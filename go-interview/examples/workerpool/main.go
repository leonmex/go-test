// Example 4: CONCURRENCY — WORKER POOL, CHANNELS, WAITGROUP, CONTEXT TIMEOUT.
// The #1 live-coding question for Go backend interviews.
//
// Interview points:
//   - Unbuffered vs buffered channels; the SENDER closes a channel, never the receiver.
//   - sync.WaitGroup waits for goroutines; close(results) only after wg.Wait().
//   - select + ctx.Done() lets every goroutine stop on timeout -> no goroutine leaks.
//   - Run tests with -race to detect data races.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type result struct {
	job int
	out int
}

func worker(ctx context.Context, id int, jobs <-chan int, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs { // loop ends when jobs is closed
		select {
		case <-ctx.Done():
			return
		case results <- result{job: j, out: j * j}: // "work"
		}
	}
}

func process(ctx context.Context, nums []int, workers int) []result {
	jobs := make(chan int)
	results := make(chan result)
	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go worker(ctx, w, jobs, results, &wg)
	}

	// producer
	go func() {
		defer close(jobs)
		for _, n := range nums {
			select {
			case <-ctx.Done():
				return
			case jobs <- n:
			}
		}
	}()

	// closer: when all workers finish, close results so the range below ends
	go func() {
		wg.Wait()
		close(results)
	}()

	var out []result
	for r := range results {
		out = append(out, r)
	}
	return out
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res := process(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8}, 3)
	fmt.Println("processed:", len(res))
	for _, r := range res {
		fmt.Printf("job %d -> %d\n", r.job, r.out)
	}
}
