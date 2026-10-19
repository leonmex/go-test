package orders

import (
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	OrderID string
	Status  Status
	Err     error
	At      time.Time
}

// Bus fans one event out to every subscriber (one channel each).
// Publish never blocks: a slow subscriber loses events instead of stalling the workers.
type Bus struct {
	mu      sync.RWMutex
	subs    map[int]chan Event
	nextID  int
	closed  bool
	dropped atomic.Int64
}

func NewBus() *Bus {
	return &Bus{subs: make(map[int]chan Event)}
}

func (b *Bus) Subscribe(buffer int) (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, buffer)
	if b.closed {
		close(ch)
		return ch, func() {}
	}
	id := b.nextID
	b.nextID++
	b.subs[id] = ch

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if c, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(c)
		}
	}
	return ch, unsubscribe
}

func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
			b.dropped.Add(1)
		}
	}
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for id, ch := range b.subs {
		close(ch)
		delete(b.subs, id)
	}
}

func (b *Bus) Dropped() int64 { return b.dropped.Load() }
