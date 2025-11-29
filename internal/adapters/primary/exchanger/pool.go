package exchanger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"marketflow/pkg/conc"
)

type Exchanger interface {
	Stream(ctx context.Context, out chan<- conc.Task, results chan<- Result)
	Stop() error
}

type Pool struct {
	MaxCount   int
	Exchangers map[string]Exchanger
	numClients int

	wg     *sync.WaitGroup
	out    chan conc.Task
	result chan Result
	mu     sync.Mutex
}

func NewPool(maxCount int) *Pool {
	pool := &Pool{
		MaxCount:   maxCount,
		Exchangers: make(map[string]Exchanger),
		numClients: 0,
		wg:         &sync.WaitGroup{},
		out:        make(chan conc.Task),
		result:     make(chan Result),
	}
	return pool
}

func (p *Pool) GetConnectedExchangers() map[string]bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	connected := make(map[string]bool)
	for name := range p.Exchangers {
		connected[name] = true
	}
	return connected
}

func (p *Pool) Add(ctx context.Context, name, host, port string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := p.numClients + 1
	if n >= p.MaxCount {
		return fmt.Errorf("max exchangers limit reached: %d", p.MaxCount)
	}
	p.numClients = n

	if _, exists := p.Exchangers[name]; exists {
		return fmt.Errorf("exchanger with name %s already exists", name)
	}

	worker, err := NewLiveExchanger(name, host, port)
	if err != nil {
		return err
	}
	p.Exchangers[name] = worker

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		worker.Stream(ctx, p.out, p.result)
		p.mu.Lock()
		defer p.mu.Unlock()
		p.numClients--
		delete(p.Exchangers, name)
	}()

	return nil
}

func (p *Pool) AddTest(ctx context.Context, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := p.numClients + 1

	if n >= p.MaxCount {
		return fmt.Errorf("max exchangers limit reached: %d", p.MaxCount)
	}

	p.numClients = n

	if _, exists := p.Exchangers[name]; exists {
		return fmt.Errorf("exchanger with name %s already exists", name)
	}

	worker, err := NewTestExchanger(name, 100*time.Millisecond)
	if err != nil {
		return err
	}

	p.Exchangers[name] = worker

	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		worker.Stream(ctx, p.out, p.result)
		p.mu.Lock()
		defer p.mu.Unlock()
		p.numClients--
		delete(p.Exchangers, name)
	}()
	return nil
}

func (p *Pool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if exchanger, ok := p.Exchangers[name]; ok {
		exchanger.Stop()
		p.numClients--
		delete(p.Exchangers, name)
	}
}

func (p *Pool) StopPool() {
	p.mu.Lock()
	fmt.Println(len(p.Exchangers))
	for n, exchanger := range p.Exchangers {
		slog.Warn("stopping exchanger...", "name", n)
		exchanger.Stop()
	}
	p.mu.Unlock()

	p.wg.Wait()
	close(p.out)
	close(p.result)
}

func (p *Pool) Out() <-chan conc.Task {
	return p.out
}

type Result struct {
	Name          string
	Host          string
	Port          string
	ReceivedTasks int
	Err           error
}

func (p *Pool) Results() <-chan Result {
	return p.result
}
