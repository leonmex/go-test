package main

import (
	"context"
	"fmt"
	"time"
)

// produce sends v every interval until ctx is cancelled, then closes the channel.
func produce[T any](ctx context.Context, v T, every time.Duration) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for {
			select {
			case ch <- v:
			case <-ctx.Done():
				return
			}
			select {
			case <-time.After(every):
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// listen receives from whichever channel has data available first, until ctx ends.
// A closed channel is set to nil: a nil channel blocks forever, so select skips it.
func listen(ctx context.Context, strings <-chan string, nums <-chan int) (gotStrings, gotNums int) {
	for {
		select {
		case msg, ok := <-strings:
			if !ok {
				strings = nil
				continue
			}
			fmt.Println(msg)
			gotStrings++
		case msg, ok := <-nums:
			if !ok {
				nums = nil
				continue
			}
			fmt.Println(msg)
			gotNums++
		case <-ctx.Done():
			fmt.Println("Context cancelled:", ctx.Err())
			return
		}
	}
}

func main() {
	strings := produce(context.Background(), "Hello", time.Millisecond*200)
	nums := produce(context.Background(), 1, time.Second*2)

	/*
		Sequential blocking: The loop first waits for
		a string from strings channel, then waits for a
		number from nums channel. It must receive both in
			that exact order before starting the next
		iteration
	*/
	for {
		fmt.Println(<-strings)
		// This will block waiting
		fmt.Println(<-nums)
	}

	// The fix is listen(ctx, strings, nums) with a 3s timeout: see main_test.go.
}
