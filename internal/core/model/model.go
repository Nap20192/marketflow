package model

import (
	"time"
)

const TimeOfAverage = 1 * time.Minute

type AgregetedData struct {
	ID           int64
	PairName     string
	Exchange     string
	Timestamp    time.Time
	AveragePrice float64
	MinPrice     float64
	MaxPrice     float64
}

type Trade struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

const (
	BTCUSDT  = "BTCUSDT"
	DOGEUSDT = "DOGEUSDT"
	TONUSDT  = "TONUSDT"
	SOLUSDT  = "SOLUSDT"
	ETHUSDT  = "ETHUSDT"
)
