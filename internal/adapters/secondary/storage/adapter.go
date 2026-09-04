package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"marketflow/infrastucture/redis"
	"marketflow/internal/adapters/secondary/cache"
	"marketflow/internal/core/model"
)

var ErrNoData = fmt.Errorf("no data found")

type StorageAdapter struct {
	cache      cache.Cache
	fallback   cache.Cache
	repository DBRepository
}

func NewStorageAdapter(cache cache.Cache, fallback cache.Cache, repository DBRepository) *StorageAdapter {
	return &StorageAdapter{
		cache:      cache,
		fallback:   fallback,
		repository: repository,
	}
}

func (s *StorageAdapter) GetAverage(ctx context.Context, arg Params) (float64, error) {
	if arg.Interval <= 1*time.Minute {
		data, err := s.getFromCache(ctx, arg.Exchange, arg.PairName, arg.Interval)
		if err != nil {
			return 0, err
		}

		return data.AveragePrice, nil
	}

	value, err := s.repository.GetAverage(ctx, arg)
	if err != nil {
		slog.Error("failed to get average from db", "error", err, "params", arg)
		return 0, err
	}

	return value, nil
}

func (s *StorageAdapter) GetMax(ctx context.Context, arg Params) (float64, error) {
	if arg.Interval <= 1*time.Minute {
		data, err := s.getFromCache(ctx, arg.Exchange, arg.PairName, arg.Interval)
		if err != nil {
			return 0, err
		}

		return data.MaxPrice, nil
	}
	return s.repository.GetMax(ctx, arg)
}

func (s *StorageAdapter) GetMin(ctx context.Context, arg Params) (float64, error) {
	if arg.Interval <= 1*time.Minute {
		if data, err := s.getFromCache(ctx, arg.Exchange, arg.PairName, arg.Interval); err == nil {
			return data.MinPrice, nil
		} else {
			return 0, err
		}
	}
	return s.repository.GetMin(ctx, arg)
}

func (s *StorageAdapter) GetLatest(ctx context.Context, exchange string, symbol string) (float64, error) {
	data, err := s.cache.GetLatest(ctx, exchange, symbol)
	if err != nil {
		if errors.Is(err, redis.ErrNoConnection) {
			data, err = s.fallback.GetLatest(ctx, exchange, symbol)
			if err != nil {
				return 0, ErrNoData
			}
			return data, nil
		}

		return 0, err
	}

	return data, nil
}

func (s *StorageAdapter) InsertMarket(ctx context.Context, arg InsertMarketParams) (model.AggregatedData, error) {
	return s.repository.InsertMarket(ctx, arg)
}

func (s *StorageAdapter) getFromCache(ctx context.Context, exchanger string, symbol string, interval time.Duration) (model.AggregatedData, error) {
	data, err := s.cache.GetRawData(ctx, exchanger, symbol, interval)
	if err != nil {
		data, err = s.fallback.GetRawData(ctx, exchanger, symbol, interval)
		if err != nil {
			return model.AggregatedData{}, err
		}
	}

	if len(data) == 0 {
		return model.AggregatedData{}, ErrNoData
	}

	var sum float64
	minPrice := data[0].Price
	maxPrice := data[0].Price
	for _, trade := range data {
		sum += trade.Price
		if trade.Price < minPrice {
			minPrice = trade.Price
		}
		if trade.Price > maxPrice {
			maxPrice = trade.Price
		}
	}
	avg := sum / float64(len(data))

	return model.AggregatedData{
		PairName:     symbol,
		Exchange:     exchanger,
		AveragePrice: avg,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
	}, nil
}
