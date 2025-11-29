package service

import (
	"context"
	"encoding/json"
	"time"

	"marketflow/internal/core"
	"marketflow/internal/core/model"
	"marketflow/pkg/conc"
)

type TradeHandler struct {
	cache core.Cache
}

func NewTradeHandler(cache core.Cache) *TradeHandler {
	return &TradeHandler{
		cache: cache,
	}
}

func (th *TradeHandler) Handle(q int, task conc.Task, result chan<- conc.Result) {
	var data model.Trade

	err := json.Unmarshal([]byte(task.Data), &data)
	if err != nil {
		result <- conc.Result{
			Name:      task.From,
			Symbol:    data.Symbol,
			Price:     data.Price,
			Timestamp: data.Timestamp,
			Err:       err,
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = th.cache.SaveRawData(ctx, task.From, data)
	if err != nil {
		result <- conc.Result{
			Name:      task.From,
			Symbol:    data.Symbol,
			Price:     data.Price,
			Timestamp: data.Timestamp,
			Err:       err,
		}
		return
	}

	result <- conc.Result{
		Name:      task.From,
		Symbol:    data.Symbol,
		Price:     data.Price,
		Timestamp: data.Timestamp,
		Err:       nil,
	}
}
