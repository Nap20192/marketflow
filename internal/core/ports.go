package core

import (
	"context"
	"time"

	"marketflow/internal/adapters/secondary/storage"
	"marketflow/internal/core/model"
)

type Repository interface {
	GetAverage(ctx context.Context, arg storage.Params) (float64, error)
	GetMax(ctx context.Context, arg storage.Params) (float64, error)
	GetMin(ctx context.Context, arg storage.Params) (float64, error)
	GetLatest(ctx context.Context, exchange string, symbol string) (float64, error)
	InsertMarket(ctx context.Context, arg storage.InsertMarketParams) (model.AgregetedData, error)
}

type Cache interface {
	GetCollection(ctx context.Context) ([]string, []string, error)
	SaveRawData(ctx context.Context, exchanger string, data model.Trade) error
	GetRawData(ctx context.Context, exchanger string, symbol string, interval time.Duration) ([]model.Trade, error)
}
