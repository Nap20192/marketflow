package conc

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"marketflow/internal/core/model"
)

func FanIn(ctx context.Context, out chan Task, channels []chan Result) {
	var wg sync.WaitGroup

	for _, ch := range channels {
		if ch == nil {
			continue
		}

		wg.Add(1)
		go func(c <-chan Result) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-c:
					if !ok {
						return
					}

					if res.Err != nil {
						slog.Error("fan-in: received error from worker", "error", res.Err)
						continue
					}

					task, err := makeToTask(res)
					if err != nil {
						slog.Error("fan-in: failed to convert result to task", "error", err)
						continue
					}

					select {
					case <-ctx.Done():
						return
					case out <- task:
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		slog.Info("fan-in: all worker channels closed, closing output channel")
		close(out)
	}()
}

func makeToTask(result Result) (Task, error) {
	d := model.Trade{
		Symbol:    result.Symbol,
		Price:     result.Price,
		Timestamp: result.Timestamp,
	}

	data, err := json.Marshal(d)
	if err != nil {
		return Task{}, err
	}

	return Task{
		From: "global",
		Data: string(data),
	}, nil
}
