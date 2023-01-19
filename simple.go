package subway

import (
	"context"
	"sync"

	"golang.org/x/exp/constraints"
)

// Merge returns all the messages from the given input channels on a single channel.
// The output channel is closed when all of the input channels are closed or when
// the context is cancelled.
func Merge[T any](ctx context.Context, input ...<-chan T) <-chan T {
	done := ctx.Done()
	output := make(chan T)
	go func() {
		defer close(output)
		var wg sync.WaitGroup
		wg.Add(len(input))
		for _, c := range input {
			go func(ch <-chan T) {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					case item, ok := <-ch:
						if !ok {
							return
						}
						select {
						case output <- item:
						case <-done:
							return
						}
					}
				}
			}(c)
		}
		wg.Wait()
	}()
	return output
}

// Split reads from input and sends messages to any one of the output channels.
// The output channels are closed when the input channel is closed or when
// the context is cancelled.
func Split[T any](ctx context.Context, input <-chan T, output ...chan<- T) {
	done := ctx.Done()
	for _, c := range output {
		go func(outCh chan<- T) {
			defer close(outCh)
			for {
				select {
				case <-done:
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					select {
					case outCh <- item:
					case <-done:
						return
					}
				}
			}
		}(c)
	}
}

// Copy reads from the input channel and sends messages to all of the output channels.
// The output channels are closed when the input channel is closed or when
// the context is cancelled.
func Copy[T any](ctx context.Context, input <-chan T, output ...chan<- T) {
	done := ctx.Done()
	go func() {
		defer func() {
			for _, c := range output {
				close(c)
			}
		}()
		for {
			select {
			case <-done:
				return
			case item, ok := <-input:
				if !ok {
					return
				}
				for i := range output {
					select {
					case output[i] <- item:
					case <-done:
						return
					}
				}
			}
		}
	}()
}

func Tee[T any](ctx context.Context, c chan T) (chan T, chan T) {
	a := make(chan T)
	b := make(chan T)
	Copy(ctx, c, a, b)
	return a, b
}

func Filter[T any](ctx context.Context, c <-chan T, f func(T) bool) <-chan T {
	done := ctx.Done()
	res := make(chan T)
	go func() {
		defer close(res)
		for {
			select {
			case <-done:
				return
			case item, ok := <-c:
				if !ok {
					return
				}
				if f(item) {
					select {
					case res <- item:
					case <-done:
						return
					}
				}
			}
		}
	}()
	return res
}

func Less[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item < value })
}

func LessOrEqual[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item <= value })
}

func Greater[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item > value })
}

func GreaterOrEqual[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item >= value })
}

func Equal[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item == value })
}

func NotEqual[T constraints.Ordered](ctx context.Context, c <-chan T, value T) <-chan T {
	return Filter(ctx, c, func(item T) bool { return item != value })
}

func Make[T any](ctx context.Context, f func() (T, bool)) chan T {
	done := ctx.Done()
	result := make(chan T)
	go func() {
		defer close(result)
		for {
			x, ok := f()
			if !ok {
				return
			}
			select {
			case <-done:
				return
			case result <- x:
			}
		}
	}()
	return result
}
