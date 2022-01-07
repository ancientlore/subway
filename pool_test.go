package subway

import (
	"context"
	"testing"
)

func TestWorker(t *testing.T) {
	ctx := context.Background()
	result := make(chan Result[int])
	go Worker(ctx, generate(ctx, 20), func(ctx context.Context, x int) (int, error) {
		return x * x, nil
	}, result)
	for i := 0; i < 20; i++ {
		x, ok := <-result
		if !ok {
			t.Error("Channel closed too soon")
		}
		if x.Err != nil {
			t.Error("Did not expect error")
		}
		if x.R != i*i {
			t.Errorf("Expected %d, got %d", i*i, x)
		}
	}
}

func TestPool(t *testing.T) {
	ctx := context.Background()
	result := Pool(ctx, generate(ctx, 30), 5, func(ctx context.Context, x int) (int, error) {
		return x * x, nil
	})
	for i := 0; i < 30; i++ {
		x, ok := <-result
		if !ok {
			t.Error("Channel closed too soon")
		}
		if x.Err != nil {
			t.Error("Did not expect error")
		}
	}
	_, ok := <-result
	if ok {
		t.Error("Expected channel to be closed")
	}
}
