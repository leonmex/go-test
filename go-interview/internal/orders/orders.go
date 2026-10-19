// Package orders is Example 5: A SERVICE BUILT ON GOROUTINES, WAITGROUP AND CHANNELS.
//
// Interview points:
//   - Bounded worker pool: N goroutines read one queue channel -> concurrency is capped,
//     the queue size gives backpressure (Submit blocks when it is full).
//   - Every goroutine has an exit: workers stop on ctx.Done(); Run waits for them (WaitGroup)
//     before closing the subscriber channels -> no leaks, no send on closed channel.
//   - The SENDER closes a channel. Many goroutines call Submit, so the queue is never closed;
//     shutdown is signalled with ctx + the stopped channel instead.
//   - Per-order timeout: context.WithTimeout derived from the service ctx.
//   - Dependencies are interfaces (Payer, Deliverer); PayFunc/DeliverFunc are adapters
//     like http.HandlerFunc, so tests pass plain functions.
package orders

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrStopped = errors.New("orders: service stopped")

type Status string

const (
	StatusPaid      Status = "paid"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Order struct {
	ID          string
	AmountCents int64
}

type Payer interface {
	Pay(ctx context.Context, o Order) error
}

type Deliverer interface {
	Deliver(ctx context.Context, o Order) error
}

type PayFunc func(ctx context.Context, o Order) error

func (f PayFunc) Pay(ctx context.Context, o Order) error { return f(ctx, o) }

type DeliverFunc func(ctx context.Context, o Order) error

func (f DeliverFunc) Deliver(ctx context.Context, o Order) error { return f(ctx, o) }

type Config struct {
	Workers      int
	QueueSize    int
	OrderTimeout time.Duration
}

type Service struct {
	payer     Payer
	deliverer Deliverer
	cfg       Config
	queue     chan Order
	bus       *Bus
	stopped   chan struct{}
	runOnce   sync.Once
}

func NewService(p Payer, d Deliverer, cfg Config) *Service {
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.QueueSize < 0 {
		cfg.QueueSize = 0
	}
	if cfg.OrderTimeout <= 0 {
		cfg.OrderTimeout = 10 * time.Second
	}
	return &Service{
		payer:     p,
		deliverer: d,
		cfg:       cfg,
		queue:     make(chan Order, cfg.QueueSize),
		bus:       NewBus(),
		stopped:   make(chan struct{}),
	}
}

// Subscribe returns a channel of events for every order. The channel is closed when Run returns.
func (s *Service) Subscribe(buffer int) (<-chan Event, func()) {
	return s.bus.Subscribe(buffer)
}

// DroppedEvents counts notifications lost because a subscriber was too slow.
func (s *Service) DroppedEvents() int64 { return s.bus.Dropped() }

// Run starts the workers and blocks until ctx is cancelled and every worker has exited.
// Orders still in the queue at that point are not processed.
func (s *Service) Run(ctx context.Context) {
	s.runOnce.Do(func() {
		var wg sync.WaitGroup
		for range s.cfg.Workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.worker(ctx)
			}()
		}
		wg.Wait()
		close(s.stopped)
		s.bus.Close()
	})
}

// Submit queues an order. It blocks while the queue is full (backpressure).
func (s *Service) Submit(ctx context.Context, o Order) error {
	select {
	case <-s.stopped:
		return ErrStopped
	default:
	}
	select {
	case s.queue <- o:
		return nil
	case <-s.stopped:
		return ErrStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-s.queue:
			s.handle(ctx, o)
		}
	}
}

func (s *Service) handle(ctx context.Context, o Order) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.OrderTimeout)
	defer cancel()

	if err := s.payer.Pay(ctx, o); err != nil {
		s.publish(o, StatusFailed, fmt.Errorf("payment: %w", err))
		return
	}
	s.publish(o, StatusPaid, nil)

	if err := s.deliverer.Deliver(ctx, o); err != nil {
		s.publish(o, StatusFailed, fmt.Errorf("delivery: %w", err))
		return
	}
	s.publish(o, StatusDelivered, nil)
}

func (s *Service) publish(o Order, st Status, err error) {
	s.bus.Publish(Event{OrderID: o.ID, Status: st, Err: err, At: time.Now()})
}
