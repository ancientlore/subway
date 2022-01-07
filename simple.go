package subway

import (
	"constraints"
	"context"
	"sync"
)

// Merge returns the messages from the given channels on a single channel.
func Merge[T any](ctx context.Context, t ...<-chan T) <-chan T {
	done := ctx.Done()
	result := make(chan T)
	go func() {
		defer close(result)
		var wg sync.WaitGroup
		wg.Add(len(t))
		for _, c := range t {
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
						result <- item
					}
				}
			}(c)
		}
		wg.Wait()
	}()
	return result
}

// Split reads from and sends messages to any one of the to channels.
func Split[T any](ctx context.Context, from <-chan T, to ...chan<- T) {
	done := ctx.Done()
	for _, c := range to {
		go func(ch chan<- T) {
			defer close(ch)
			for {
				select {
				case <-done:
					return
				case item, ok := <-from:
					if !ok {
						return
					}
					ch <- item
				}
			}
		}(c)
	}
}

// Copy reads from and sends messages to all of the to channels.
func Copy[T any](ctx context.Context, from <-chan T, to ...chan<- T) {
	done := ctx.Done()
	go func() {
		defer func() {
			for _, c := range to {
				close(c)
			}
		}()
		for {
			select {
			case <-done:
				return
			case item, ok := <-from:
				if !ok {
					return
				}
				for i := range to {
					select {
					case to[i] <- item:
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
					res <- item
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
