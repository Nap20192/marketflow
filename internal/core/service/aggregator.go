package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"marketflow/internal/adapters/secondary/storage"
	"marketflow/internal/core"
)

const TimeTicker = 1 * time.Second

type Aggregator struct {
	cache core.Cache
	repo  core.Repository
	err   error
}

func NewAggregator(cache core.Cache, repo core.Repository) *Aggregator {
	return &Aggregator{
		cache: cache,
		repo:  repo,
	}
}

func (a *Aggregator) Start(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {

		case <-ticker.C:

			aCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := a.aggregate(aCtx, time.Minute); err != nil {
				continue
			}

			if a.err != nil {
				slog.Error("aggregation error", "error", a.err)
				a.err = nil
				continue
			}

		case <-ctx.Done():
			aCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			a.aggregate(aCtx, time.Minute)

			return nil
		}
	}
}

func (a *Aggregator) aggregate(ctx context.Context, interval time.Duration) error {
	exchangers, symbols, err := a.cache.GetCollection(ctx)
	if err != nil {
		return err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(exchangers) * len(symbols))

	for _, exchanger := range exchangers {
		for _, symbol := range symbols {
			go func(exchanger, symbol string) {
				defer wg.Done()

				rawData, err := a.cache.GetRawData(ctx, exchanger, symbol, interval)
				if err != nil {
					slog.Error("failed to get raw data", "error", err)
					return
				}

				if len(rawData) == 0 {
					return
				}

				var sum float64
				var min, max float64
				min = rawData[0].Price
				max = rawData[0].Price
				for _, data := range rawData {
					sum += data.Price
					if data.Price < min {
						min = data.Price
					}
					if data.Price > max {
						max = data.Price
					}
				}
				average := sum / float64(len(rawData))
				_, err = a.repo.InsertMarket(ctx, storage.InsertMarketParams{
					PairName:     symbol,
					Exchange:     exchanger,
					AveragePrice: average,
					MinPrice:     min,
					MaxPrice:     max,
				})
				if err != nil {
					mu.Lock()
					a.err = err
					mu.Unlock()
					return
				}
			}(exchanger, symbol)
		}
	}

	wg.Wait()

	return nil
}
