package subway

import (
	"context"
	"testing"
)

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func generate[T signed](ctx context.Context, count T) chan T {
	res := make(chan T)
	done := ctx.Done()
	go func() {
		defer close(res)
		for i := T(0); i < count; i++ {
			select {
			case <-done:
				return
			case res <- i:
			}
		}
	}()
	return res
}

func TestGenerate(t *testing.T) {
	c := generate(context.Background(), 3)
	for i := 0; i < 3; i++ {
		x := <-c
		if x != i {
			t.Errorf("Expected %d, got %d", i, x)
		}
	}
	_, ok := <-c
	if ok {
		t.Error("Expected channel to be closed")
	}
}
