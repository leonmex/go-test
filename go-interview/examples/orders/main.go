// Example 5 demo: run with  go run ./examples/orders
// 8 orders, 3 workers. ORDER-4 is declined, ORDER-7 is slower than the per-order timeout.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/example/go-interview/internal/orders"
)

var errDeclined = errors.New("card declined")

// sleep is the original processPayment timer, reusable for any step.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, 7*time.Second)
	defer cancelTimeout()

	pay := orders.PayFunc(func(ctx context.Context, o orders.Order) error {
		if o.ID == "ORDER-4" {
			return errDeclined
		}
		return sleep(ctx, time.Duration(200+rand.IntN(300))*time.Millisecond)
	})
	deliver := orders.DeliverFunc(func(ctx context.Context, o orders.Order) error {
		d := time.Duration(300+rand.IntN(500)) * time.Millisecond
		if o.ID == "ORDER-7" {
			d = 5 * time.Second
		}
		return sleep(ctx, d)
	})

	svc := orders.NewService(pay, deliver, orders.Config{
		Workers:      3,
		QueueSize:    10,
		OrderTimeout: 2 * time.Second,
	})

	const total = 8
	events, unsubscribe := svc.Subscribe(total * 3)
	defer unsubscribe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.Run(ctx)
	}()

	for i := 1; i <= total; i++ {
		o := orders.Order{ID: fmt.Sprintf("ORDER-%d", i), AmountCents: int64(i) * 1000}
		if err := svc.Submit(ctx, o); err != nil {
			fmt.Println("submit:", err)
			break
		}
	}

	// Notifications: stop once every order reached a final state (or the 7s deadline hits).
	done := 0
	for e := range events {
		switch e.Status {
		case orders.StatusPaid:
			fmt.Printf("[notify] %s paid\n", e.OrderID)
		case orders.StatusDelivered:
			fmt.Printf("[notify] %s DELIVERED\n", e.OrderID)
			done++
		case orders.StatusFailed:
			fmt.Printf("[notify] %s FAILED: %v\n", e.OrderID, e.Err)
			done++
		}
		if done == total {
			cancelTimeout()
		}
	}

	wg.Wait()
	fmt.Printf("finished: %d/%d orders final, %d events dropped\n", done, total, svc.DroppedEvents())
}
