package subway

import (
	"context"
	"math/rand"
	"sync"
	"testing"
)

func TestMergeNil(t *testing.T) {
	c := Merge[int](context.Background())
	for x := range c {
		t.Errorf("Received value %v from nil channel", x)
	}
}

func TestMergeOne(t *testing.T) {
	ctx := context.Background()
	c := Merge(ctx, generate(ctx, 3))
	for i := 0; i < 3; i++ {
		x, ok := <-c
		if !ok {
			t.Errorf("Channel should not be closed yet")
		}
		if i != x {
			t.Errorf("Expected %d, got %d", i, x)
		}
	}
	_, ok := <-c
	if ok {
		t.Error("Expected channel to be closed")
	}
}

func TestMergeMulti(t *testing.T) {
	ctx := context.Background()
	c := Merge(ctx, generate(ctx, 3), generate(ctx, 8), generate(ctx, 0))
	i := 0
	for range c {
		i++
	}
	if i != 11 {
		t.Errorf("Channel generated incorrect number of results")
	}
}

func TestSplit(t *testing.T) {
	ctx := context.Background()
	c := make([]chan int, 3)
	c[0] = make(chan int)
	c[1] = make(chan int)
	c[2] = make(chan int)
	results := make(chan int)
	Split(ctx, generate(ctx, 20), c[0], c[1], c[2])
	var wg sync.WaitGroup
	wg.Add(len(c))
	for _, ch := range c {
		go func(r chan int) {
			defer wg.Done()
			for item := range r {
				results <- item
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	i := 0
	for range results {
		i++
	}
	if i != 20 {
		t.Errorf("Expected 20 results, got %d", i)
	}
}

func TestCopy(t *testing.T) {
	ctx := context.Background()
	c := make([]chan int, 3)
	c[0] = make(chan int)
	c[1] = make(chan int)
	c[2] = make(chan int)
	results := make(chan int)
	Copy(ctx, generate(ctx, 20), c[0], c[1], c[2])
	var wg sync.WaitGroup
	wg.Add(len(c))
	for _, ch := range c {
		go func(r chan int) {
			defer wg.Done()
			i := 0
			for item := range r {
				if item != i {
					t.Errorf("Expected %d but got %d", i, item)
				}
				results <- item
				i++
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	i := 0
	for range results {
		i++
	}
	if i != 60 {
		t.Errorf("Expected 60 results, got %d", i)
	}
}

func TestFilter(t *testing.T) {
	ctx := context.Background()
	c := Filter(ctx, generate(ctx, 30), func(x int) bool { return x%2 == 0 })
	count := 0
	for i := range c {
		if i%2 != 0 {
			t.Errorf("Expected an even number but got %d", i)
		}
		count++
	}
	if count != 15 {
		t.Errorf("Expected 15 results but got %d", count)
	}
}

func TestMake(t *testing.T) {
	c := Make(context.Background(), func() (int, bool) {
		x := rand.Intn(100)
		return x, x < 80
	})
	for i := range c {
		t.Logf("Got %d", i)
	}
}
