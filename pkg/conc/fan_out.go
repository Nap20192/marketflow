package conc

import (
	"context"
	"sync"
)

func FanOut(ctx context.Context, src <-chan Task, dests ...chan Task) {
	var wg sync.WaitGroup

	for s := range src {
		for _, d := range dests {
			wg.Add(1)
			go func(val Task, dest chan Task) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				case dest <- val:
				default:
				}
			}(s, d)
		}
	}

	wg.Wait()
	for _, d := range dests {
		close(d)
	}
}
