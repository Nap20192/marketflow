package conc

import (
	"context"
	"sync"
)

func FanOut(ctx context.Context, src <-chan Task, dests ...chan Task) {
	var wg sync.WaitGroup

	for task := range src {
		for _, dest := range dests {
			wg.Add(1)
			go func(task Task, dest chan Task) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				case dest <- task:
				default:
				}
			}(task, dest)
		}
	}

	wg.Wait()
	for _, dest := range dests {
		close(dest)
	}
}
