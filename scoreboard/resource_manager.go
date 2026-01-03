package scoreboard

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Generic fetch function signature.
type FetchFunc[T any] func(ctx context.Context) (T, error)

// Resource manages a single resource of type T:
// - producer runs in background and stores the latest value in an atomic pointer
// - Done channel is closed when the producer finishes (normal stop or error)
type Resource[T any] struct {
	name          string
	fetch         FetchFunc[T]
	interval      time.Duration
	maxIterations int

	ptr    atomic.Pointer[T]
	doneCh chan struct{}
}

// NewResource constructs a Resource.
func NewResource[T any](name string, interval time.Duration, maxIterations int, fetch FetchFunc[T]) *Resource[T] {
	return &Resource[T]{
		name:          name,
		fetch:         fetch,
		interval:      interval,
		maxIterations: maxIterations,
		doneCh:        make(chan struct{}),
	}
}

// Start runs the producer goroutine. It increments the provided WaitGroup.
func (r *Resource[T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			// ensure doneCh is closed exactly once
			select {
			case <-r.doneCh:
			default:
				close(r.doneCh)
			}
		}()

		count := 0
		for {
			// quick exit on context cancel
			select {
			case <-ctx.Done():
				fmt.Printf("%s: shutting down due to context cancel\n", r.name)
				return
			default:
			}

			count++
			fmt.Printf("%s: fetching iteration %d\n", r.name, count)
			val, err := r.fetch(ctx)
			if err != nil {
				fmt.Printf("%s: fetch error: %v\n", r.name, err)
				// stop producer on error (adjust if you want retries instead)
				return
			}

			// store a copy atomically
			tmp := val
			r.ptr.Store(&tmp)
			fmt.Printf("%s: stored snapshot\n", r.name)

			// stop after configured iterations (0 => run forever)
			if r.maxIterations > 0 && count >= r.maxIterations {
				fmt.Printf("%s: reached max iterations (%d), finishing\n", r.name, r.maxIterations)
				return
			}

			// sleep but respect context cancellation
			select {
			case <-ctx.Done():
				fmt.Printf("%s: shutting down during sleep\n", r.name)
				return
			case <-time.After(r.interval):
			}
		}
	}()
}

// Latest returns the latest stored value pointer (may be nil).
func (r *Resource[T]) Latest() *T {
	return r.ptr.Load()
}

// Done returns a channel that's closed when the producer finishes.
func (r *Resource[T]) Done() <-chan struct{} {
	return r.doneCh
}
