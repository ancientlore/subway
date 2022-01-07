package subway

import (
	"context"
	"sync"
)

// Func is a function that performs an operation on t.
type Func[T, R any] func(ctx context.Context, t T) (R, error)

// Result holds the data returned by a Func.
type Result[R any] struct {
	R   R
	Err error
}

// Worker reads c and calls f for each item until c is closed
// or the context is cancelled.
func Worker[T, R any](ctx context.Context, c <-chan T, f Func[T, R], res chan<- Result[R]) {
	done := ctx.Done()
	for {
		select {
		case <-done:
			// processing cancelled; return
			return
		case t, ok := <-c:
			if !ok {
				// channel closed; all done
				return
			}
			// Here we do the work
			r, err := f(ctx, t)
			select {
			case res <- Result[R]{R: r, Err: err}:
			case <-done:
				return
			}
		}
	}
}

// Pool spawns size workers to processes messages on c using f,
// returning the results in a channel. The resulting channel must
// be read until it is closed or the process will hang.
func Pool[T, R any](ctx context.Context, c <-chan T, size int, f Func[T, R]) <-chan Result[R] {
	res := make(chan Result[R])

	go func() {
		defer close(res)
		var wg sync.WaitGroup
		wg.Add(size)
		// Start workers to process the channel T
		for i := 0; i < size; i++ {
			go func() {
				defer wg.Done()
				Worker(ctx, c, f, res)
			}()
		}
		wg.Wait()
	}()

	return res
}
