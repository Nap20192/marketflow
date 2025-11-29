package storage

import (
	"context"
	"time"

	"marketflow/internal/core/model"
)

type Params struct {
	PairName string
	Exchange string
	Interval time.Duration
}

type InsertMarketParams struct {
	PairName     string
	Exchange     string
	AveragePrice float64
	MinPrice     float64
	MaxPrice     float64
}

type DBRepository interface {
	GetAverage(ctx context.Context, arg Params) (float64, error)
	GetMax(ctx context.Context, arg Params) (float64, error)
	GetMin(ctx context.Context, arg Params) (float64, error)
	InsertMarket(ctx context.Context, arg InsertMarketParams) (model.AgregetedData, error)
}
