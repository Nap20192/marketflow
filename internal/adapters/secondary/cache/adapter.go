package cache

import (
	"context"
	"errors"
	"time"

	"marketflow/infrastucture/redis"
	"marketflow/internal/core/model"
)

type CacheAdapter struct {
	cache    Cache
	fallback Cache
}

func NewCacheAdapter(cache Cache, fallback Cache) *CacheAdapter {
	return &CacheAdapter{
		cache:    cache,
		fallback: fallback,
	}
}

func (c *CacheAdapter) GetCollection(ctx context.Context) ([]string, []string, error) {
	exchangers, symbols, err := c.cache.GetCollection(ctx)
	if err != nil {
		if errors.Is(err, redis.ErrNoConnection) {
			exchangers, symbols, err = c.fallback.GetCollection(ctx)
			if err != nil {
				return nil, nil, err
			}
			return exchangers, symbols, nil
		}
		return nil, nil, err
	}

	return exchangers, symbols, nil
}

func (c *CacheAdapter) SaveRawData(ctx context.Context, exchanger string, data model.Trade) error {
	err := c.cache.SaveRawData(ctx, exchanger, data)
	if err != nil {
		if errors.Is(err, redis.ErrNoConnection) {
			if err := c.fallback.SaveRawData(ctx, exchanger, data); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

func (c *CacheAdapter) GetRawData(ctx context.Context, exchanger string, symbol string, interval time.Duration) ([]model.Trade, error) {
	data, err := c.cache.GetRawData(ctx, exchanger, symbol, interval)
	if err != nil {
		if errors.Is(err, redis.ErrNoConnection) {
			return c.fallback.GetRawData(ctx, exchanger, symbol, interval)
		}
		return nil, err
	}

	return data, nil
}
