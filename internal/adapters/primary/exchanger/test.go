package exchanger

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"marketflow/internal/core/model"
	"marketflow/pkg/conc"
)

type TestExchanger struct {
	Name     string
	interval time.Duration
	cancel   context.CancelFunc
}

func NewTestExchanger(name string, interval time.Duration) (*TestExchanger, error) {
	if interval <= 0 {
		interval = time.Millisecond * 500
	}
	return &TestExchanger{
		Name:     name,
		interval: interval,
	}, nil
}

func (t *TestExchanger) Stream(ctx context.Context, out chan<- conc.Task, results chan<- Result) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t.cancel = cancel
	ticker := time.NewTicker(t.interval)

	for {
		select {
		case <-ticker.C:
			out <- generateTestData(t.Name)
		case <-ctx.Done():
			results <- Result{Name: t.Name, Err: nil}
			return
		}
	}
}

func (t *TestExchanger) Stop() error {
	t.cancel()
	return nil
}

func generateTestData(name string) conc.Task {
	r := rand.Intn(5)

	var data model.Trade

	switch r {

	case 0:
		data = model.Trade{
			Symbol:    "BTCUSDT",
			Price:     100000 + rand.Float64()*10000,
			Timestamp: time.Now().Unix(),
		}
	case 1:
		data = model.Trade{
			Symbol:    "ETHUSDT",
			Price:     2000 + rand.Float64()*100,
			Timestamp: time.Now().Unix(),
		}
	case 2:
		data = model.Trade{
			Symbol:    "SOLUSDT",
			Price:     100 + rand.Float64()*10,
			Timestamp: time.Now().Unix(),
		}
	case 3:
		data = model.Trade{
			Symbol:    "DOGEUSDT",
			Price:     0.1 + rand.Float64()*0.1,
			Timestamp: time.Now().Unix(),
		}
	case 4:
		data = model.Trade{
			Symbol:    "TONUSDT",
			Price:     1.2 + rand.Float64(),
			Timestamp: time.Now().Unix(),
		}
	default:
		data = model.Trade{
			Symbol:    "BTCUSDT",
			Price:     100000 + rand.Float64()*1000,
			Timestamp: time.Now().Unix(),
		}
	}

	d, _ := json.Marshal(data)

	return conc.Task{
		From: name,
		Data: string(d),
	}
}
