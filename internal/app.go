package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"marketflow/infrastucture/postgres"
	iredis "marketflow/infrastucture/redis"
	"marketflow/internal/adapters/primary/exchanger"
	"marketflow/internal/adapters/primary/ui"
	"marketflow/internal/adapters/primary/ui/handlers"
	"marketflow/internal/adapters/primary/ui/middleware"
	"marketflow/internal/adapters/secondary/cache"
	"marketflow/internal/adapters/secondary/storage"
	"marketflow/internal/core/service"

	"marketflow/pkg/conc"
	"marketflow/pkg/group"
)

type App struct {
	config       config
	serverConfig *ui.ServerConfig

	mode   string
	modeMu *sync.Mutex

	server *ui.Server
	redis  *iredis.Queries

	postgres *postgres.Client
	repo     *postgres.Queries

	poolClients *exchanger.Pool
	handler     *handlers.Handler

	fanoutChannels []chan conc.Task
	workerChannels []chan conc.Result
	fanin          chan conc.Task
	faninResult    chan conc.Result

	tradeHandler *service.TradeHandler
	aggregator   *service.Aggregator
	stats        *service.Stats

	storageAdapter *storage.StorageAdapter
	cacheAdapter   *cache.CacheAdapter
	workerWg       sync.WaitGroup
}

func NewApp(port *ui.ServerConfig) *App {
	return &App{
		mode:           "live",
		modeMu:         &sync.Mutex{},
		fanin:          make(chan conc.Task),
		faninResult:    make(chan conc.Result),
		fanoutChannels: make([]chan conc.Task, 3),
		workerChannels: make([]chan conc.Result, 3),
		serverConfig:   port,
		workerWg:       sync.WaitGroup{},
	}
}

func (a *App) di() error {
	a.poolClients = exchanger.NewPool(4)

	a.storageAdapter = storage.NewStorageAdapter(a.redis, a.repo, a.repo)
	a.cacheAdapter = cache.NewCacheAdapter(a.redis, a.repo)

	a.aggregator = service.NewAggregator(a.cacheAdapter, a.storageAdapter)
	a.tradeHandler = service.NewTradeHandler(a.cacheAdapter)
	a.stats = service.NewStats(a.storageAdapter)

	a.handler = handlers.NewHandler(a.stats)
	routes := ui.RegisterRoutes(a.handler)
	a.server = ui.NewServer(a.serverConfig, routes)
	routesWithLogger := middleware.Logger(routes)
	a.server = ui.NewServer(a.serverConfig, routesWithLogger)

	return nil
}

func (a *App) Run(ctx context.Context) error {
	var err error
	conf, _ := LoadConfig()
	a.config = *conf

	if err = a.initRedis(ctx); err != nil {
		return err
	}

	if err = a.initPostgres(); err != nil {
		return err
	}

	if err = a.di(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = a.poolClients.Add(ctx, "exchange1", a.config.exchange1, "40101")
	if err != nil {
		return err
	}
	err = a.poolClients.Add(ctx, "exchange2", a.config.exchange2, "40102")
	if err != nil {
		return err
	}
	err = a.poolClients.Add(ctx, "exchange3", a.config.exchange3, "40103")
	if err != nil {
		return err
	}

	handlers.WithTestModeSwitch(a.SwitchToTest(ctx), a.handler)
	handlers.WithLiveModeSwitch(a.SwitchToLive(ctx), a.handler)

	handlers.WithHealthCheck(a.HealthCheck(), a.handler)

	for i := range a.fanoutChannels {
		a.fanoutChannels[i] = make(chan conc.Task)
		a.workerChannels[i] = make(chan conc.Result)
	}

	// Create 10 workers for each channel
	for i, ch := range a.fanoutChannels {
		a.handleTasks(i, 10, ch, a.workerChannels[i])
	}

	// handle FanIn
	a.handleTasks(4, 30, a.fanin, a.faninResult)

	g, gCtx := group.WithContext(ctx)

	// Fan-out incoming tasks to worker pools
	g.Go(func() error {
		conc.FanOut(context.TODO(), a.poolClients.Out(), a.fanoutChannels...)
		return nil
	})

	g.Go(func() error {
		for i := range a.poolClients.Results() {
			if i.Err != nil {
				slog.Error("pool client error", "name", i.Name, "error", i.Err)
				continue
			}
			slog.Info("pool client finished", "name", i.Name)
		}
		return nil
	})

	// Handle Pool Results
	conc.FanIn(context.TODO(), a.fanin, a.workerChannels)

	// FanIn Results
	g.Go(func() error {
		for res := range a.faninResult {
			if res.Err != nil {
				slog.Error("worker error", "name", res.Name, "error", res.Err)
				continue
			}
		}
		return nil
	})

	// Start aggregator
	g.Go(func() error {
		interval := service.TimeTicker
		if err := a.aggregator.Start(gCtx, interval); err != nil {
			slog.Error("aggregator error", "error", err)
			return err
		}
		return nil
	})

	// Start HTTP server
	g.Go(func() error {
		slog.Info("starting server on port: " + a.serverConfig.Port)
		if err := a.server.Start(); err != nil {
			slog.Error("server error", "error", err)
			cancel()
			return err
		}

		return nil
	})

	// Graceful Shutdown
	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)
		select {
		case <-gCtx.Done():
			return gCtx.Err()
		case <-c:
			slog.Info("received interrupt signal, shutting down...")
			cancel()
			return nil
		}
	})

	// Shutdown Procedures
	g.Go(func() error {
		<-gCtx.Done()
		slog.Info("shutting down exchanger pool...")
		a.poolClients.StopPool()
		a.workerWg.Wait()
		slog.Info("exchanger pool shut down complete")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		slog.Info("shutting down server...")
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
			return err
		}

		a.redis.Close()
		a.postgres.Close()

		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("application error", "error", err)
		return err
	}

	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	client, err := iredis.NewClient(ctx, a.config.redis)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		return err
	}
	a.redis = iredis.NewQueries(client)
	return nil
}

func (a *App) initPostgres() error {
	pg, err := postgres.NewClient(a.config.postgres)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		return err
	}
	a.postgres = pg
	a.repo = postgres.New(pg.Pool)

	if err := a.postgres.Pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping postgres", "error", err)
		return err
	}
	return nil
}

func (a *App) handleTasks(id int, numOfWorkers int, taskChan chan conc.Task, resultChan chan conc.Result) {
	a.workerWg.Add(1)
	go func(index int, numOfWorkers int, c chan conc.Task, Result chan conc.Result) {
		defer a.workerWg.Done()
		indexStr := strconv.Itoa(index + 1)
		workers := conc.CreateNWorkers(numOfWorkers, conc.WithName("worker-"+indexStr))

		pool := conc.NewWorkerPool(workers, a.tradeHandler)
		pool.Name = "pool-" + indexStr

		pool.Create()
		for t := range c {
			pool.Work(t, resultChan)
		}
		pool.Wait()
		pool.PrintStats()
		close(resultChan)
	}(id, numOfWorkers, taskChan, resultChan)
}

var ErrAlreadyInLiveMode = fmt.Errorf("already in live mode")

func (a *App) SwitchToLive(ctx context.Context) func() error {
	return func() error {
		if a.mode == "live" {
			return ErrAlreadyInLiveMode
		}
		fmt.Println("switching to live mode...")

		a.poolClients.Remove("exchange1")
		a.poolClients.Remove("exchange2")
		a.poolClients.Remove("exchange3")

		a.modeMu.Lock()
		a.mode = "live"
		a.modeMu.Unlock()

		if err := a.poolClients.Add(ctx, "exchange1", a.config.exchange1, "40101"); err != nil {
			return err
		}

		if err := a.poolClients.Add(ctx, "exchange2", a.config.exchange2, "40102"); err != nil {
			return err
		}

		if err := a.poolClients.Add(ctx, "exchange3", a.config.exchange3, "40103"); err != nil {
			return err
		}

		return nil
	}
}

var ErrAlreadyInTestMode = fmt.Errorf("already in test mode")

func (a *App) SwitchToTest(ctx context.Context) func() error {
	return func() error {
		if a.mode == "test" {
			return ErrAlreadyInTestMode
		}
		fmt.Println("switching to test mode...")

		a.modeMu.Lock()
		a.mode = "test"
		a.modeMu.Unlock()

		a.poolClients.Remove("exchange1")
		a.poolClients.Remove("exchange2")
		a.poolClients.Remove("exchange3")

		a.poolClients.AddTest(ctx, "exchange1")
		a.poolClients.AddTest(ctx, "exchange2")
		a.poolClients.AddTest(ctx, "exchange3")

		return nil
	}
}

type HealthCheck struct {
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

func (a *App) HealthCheck() func() []byte {
	return func() []byte {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var health HealthCheck

		if err := a.postgres.Pool.Ping(ctx); err != nil {
			health.Postgres = "Not working"
		} else {
			health.Postgres = "OK"
		}

		ctxRedis, cancelRedis := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelRedis()

		err := a.redis.Ping(ctxRedis)

		if err != nil {
			health.Redis = "Not working"
		} else {
			health.Redis = "OK"
		}

		data, _ := json.MarshalIndent(health, "", "  ")

		return data
	}
}
