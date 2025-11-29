package cache

import (
	"context"
	"time"

	"marketflow/internal/core/model"
)

type Cache interface {
	GetCollection(ctx context.Context) ([]string, []string, error)
	SaveRawData(ctx context.Context, exchanger string, data model.Trade) error
	GetRawData(ctx context.Context, exchanger string, symbol string, interval time.Duration) ([]model.Trade, error)
	GetLatest(ctx context.Context, exchange string, symbol string) (float64, error)
}
