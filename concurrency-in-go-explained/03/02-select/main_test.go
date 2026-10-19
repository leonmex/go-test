package main

import (
	"context"
	"testing"
	"time"
)

// With select the fast channel is not held back by the slow one.
// The sequential loop in main would receive exactly as many strings as nums.
func TestListen_FastChannelIsNotBlockedBySlowOne(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	strings := produce(ctx, "Hello", 10*time.Millisecond)
	nums := produce(ctx, 1, 100*time.Millisecond)

	gotStrings, gotNums := listen(ctx, strings, nums)

	if gotNums == 0 {
		t.Fatal("expected at least one num")
	}
	if gotStrings < 3*gotNums {
		t.Errorf("expected many more strings than nums, got %d strings, %d nums", gotStrings, gotNums)
	}
}

func TestListen_ReturnsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	never := make(chan string)
	nothing := make(chan int)

	done := make(chan struct{})
	go func() {
		listen(ctx, never, nothing)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listen ignored ctx cancellation")
	}
}

func TestProduce_ClosesChannelOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := produce(ctx, "x", time.Hour)

	<-ch // first value is sent immediately
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("producer goroutine leaked: channel never closed")
	}
}
