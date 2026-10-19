package orders

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	okPay     = PayFunc(func(context.Context, Order) error { return nil })
	okDeliver = DeliverFunc(func(context.Context, Order) error { return nil })
)

// startService runs svc in the background and stops it when the test ends.
func startService(t *testing.T, svc *Service) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})
	return cancel
}

// collect reads events until n orders reached a final status, or fails after 2s.
func collect(t *testing.T, events <-chan Event, n int) map[string][]Event {
	t.Helper()
	got := make(map[string][]Event)
	final := 0
	timeout := time.After(2 * time.Second)
	for final < n {
		select {
		case e, ok := <-events:
			if !ok {
				t.Fatalf("events closed after %d/%d final", final, n)
			}
			got[e.OrderID] = append(got[e.OrderID], e)
			if e.Status == StatusDelivered || e.Status == StatusFailed {
				final++
			}
		case <-timeout:
			t.Fatalf("timed out after %d/%d final", final, n)
		}
	}
	return got
}

func submitAll(t *testing.T, svc *Service, n int) {
	t.Helper()
	for i := range n {
		if err := svc.Submit(context.Background(), Order{ID: fmt.Sprintf("o%d", i)}); err != nil {
			t.Fatalf("submit: %v", err)
		}
	}
}

func TestService_DeliversAllOrders(t *testing.T) {
	svc := NewService(okPay, okDeliver, Config{Workers: 3, QueueSize: 5})
	events, _ := svc.Subscribe(100)
	startService(t, svc)

	const n = 20
	submitAll(t, svc, n)
	got := collect(t, events, n)

	if len(got) != n {
		t.Fatalf("expected %d orders, got %d", n, len(got))
	}
	for id, evs := range got {
		if len(evs) != 2 || evs[0].Status != StatusPaid || evs[1].Status != StatusDelivered {
			t.Errorf("%s: expected [paid delivered], got %v", id, evs)
		}
	}
}

func TestService_Failures(t *testing.T) {
	errDeclined := errors.New("declined")

	tests := []struct {
		name    string
		pay     PayFunc
		deliver DeliverFunc
		wantErr error
	}{
		{
			name:    "payment declined",
			pay:     func(context.Context, Order) error { return errDeclined },
			deliver: okDeliver,
			wantErr: errDeclined,
		},
		{
			name: "delivery exceeds order timeout",
			pay:  okPay,
			deliver: func(ctx context.Context, _ Order) error {
				<-ctx.Done()
				return ctx.Err()
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(tc.pay, tc.deliver, Config{Workers: 1, OrderTimeout: 50 * time.Millisecond})
			events, _ := svc.Subscribe(10)
			startService(t, svc)

			submitAll(t, svc, 1)
			evs := collect(t, events, 1)["o0"]

			last := evs[len(evs)-1]
			if last.Status != StatusFailed {
				t.Fatalf("expected failed, got %s", last.Status)
			}
			if !errors.Is(last.Err, tc.wantErr) {
				t.Errorf("expected error %v, got %v", tc.wantErr, last.Err)
			}
		})
	}
}

func TestService_RespectsWorkerLimit(t *testing.T) {
	const workers = 3
	var running, maxSeen atomic.Int32

	pay := PayFunc(func(context.Context, Order) error {
		cur := running.Add(1)
		defer running.Add(-1)
		for {
			m := maxSeen.Load()
			if cur <= m || maxSeen.CompareAndSwap(m, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	svc := NewService(pay, okDeliver, Config{Workers: workers, QueueSize: 20})
	events, _ := svc.Subscribe(100)
	startService(t, svc)

	submitAll(t, svc, 20)
	collect(t, events, 20)

	if got := maxSeen.Load(); got > workers {
		t.Errorf("expected at most %d concurrent payments, got %d", workers, got)
	}
}

func TestService_StopClosesSubscribersAndRejectsSubmit(t *testing.T) {
	svc := NewService(okPay, okDeliver, Config{Workers: 2})
	events, _ := svc.Subscribe(1)
	stop := startService(t, svc)

	stop()
	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("expected closed channel, got an event")
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber channel not closed after stop")
	}

	if err := svc.Submit(context.Background(), Order{ID: "late"}); !errors.Is(err, ErrStopped) {
		t.Errorf("expected ErrStopped, got %v", err)
	}
}

func TestService_SubmitHonoursContextWhenQueueFull(t *testing.T) {
	svc := NewService(okPay, okDeliver, Config{QueueSize: 0}) // Run not started: nobody reads

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := svc.Submit(ctx, Order{ID: "x"}); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestBus_SlowSubscriberDoesNotBlock(t *testing.T) {
	b := NewBus()
	slow, _ := b.Subscribe(0) // never read
	fast, _ := b.Subscribe(10)

	done := make(chan struct{})
	go func() {
		for range 5 {
			b.Publish(Event{OrderID: "o"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a slow subscriber")
	}
	if len(fast) != 5 {
		t.Errorf("fast subscriber: expected 5 events, got %d", len(fast))
	}
	if b.Dropped() != 5 {
		t.Errorf("expected 5 dropped, got %d", b.Dropped())
	}
	_ = slow
}

func TestBus_UnsubscribeClosesOnlyThatChannel(t *testing.T) {
	b := NewBus()
	a, unsubA := b.Subscribe(1)
	c, _ := b.Subscribe(1)

	unsubA()
	unsubA() // second call must not panic

	if _, ok := <-a; ok {
		t.Error("expected a to be closed")
	}
	b.Publish(Event{OrderID: "o"})
	if e := <-c; e.OrderID != "o" {
		t.Errorf("c: expected event o, got %v", e)
	}
}
