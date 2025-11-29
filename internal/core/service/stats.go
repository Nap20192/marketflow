package service

import (
	"context"
	"errors"
	"time"

	"marketflow/internal/adapters/secondary/storage"
	"marketflow/internal/core"
)

type Stats struct {
	repo core.Repository
}

func (s *Stats) GetAveragePrice(ctx context.Context, exchange string, symbol string, period string) (float64, error) {
	if period == "" {
		period = "24h"
	}

	interval, err := time.ParseDuration(period)
	if err != nil {
		return 0, err
	}

	if interval <= 0 {
		return 0, errors.New("incorrect period")
	}

	params := storage.Params{
		PairName: symbol,
		Exchange: exchange,
		Interval: interval,
	}

	return s.repo.GetAverage(ctx, params)
}

func (s *Stats) GetLatestPrice(ctx context.Context, exchange, symbol string) (float64, error) {
	return s.repo.GetLatest(ctx, exchange, symbol)
}

func (s *Stats) GetHighestPrice(ctx context.Context, exchange, symbol, period string) (float64, error) {
	if period == "" {
		period = "24h"
	}

	interval, err := time.ParseDuration(period)
	if err != nil {
		return 0, err
	}

	if interval <= 0 {
		return 0, errors.New("incorrect period")
	}

	params := storage.Params{
		PairName: symbol,
		Exchange: exchange,
		Interval: interval,
	}

	return s.repo.GetMax(ctx, params)
}

func (s *Stats) GetLowestPrice(ctx context.Context, exchange, symbol, period string) (float64, error) {
	if period == "" {
		period = "24h"
	}

	interval, err := time.ParseDuration(period)
	if err != nil {
		return 0, err
	}

	if interval <= 0 {
		return 0, errors.New("incorrect period")
	}

	params := storage.Params{
		PairName: symbol,
		Exchange: exchange,
		Interval: interval,
	}

	return s.repo.GetMin(ctx, params)
}

func NewStats(repo core.Repository) *Stats {
	return &Stats{
		repo: repo,
	}
}
