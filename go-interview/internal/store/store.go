// Package store holds data access. Example 1: INTERFACES + MUTEX.
//
// Interview points:
//   - Depend on an interface (Store), not a concrete type -> easy to mock in tests,
//     easy to swap memory -> Postgres/DynamoDB later.
//   - "Accept interfaces, return structs."
//   - sync.RWMutex: many readers OR one writer. Maps are NOT safe for concurrent use.
//   - Sentinel errors (ErrNotFound) let callers use errors.Is.
package store

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("item not found")

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Store is the contract. Every method takes a context (cancellation / timeouts).
type Store interface {
	Get(ctx context.Context, id string) (Item, error)
	Put(ctx context.Context, item Item) error
}

type Memory struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewMemory() *Memory {
	return &Memory{items: make(map[string]Item)}
}

func (m *Memory) Get(ctx context.Context, id string) (Item, error) {
	if err := ctx.Err(); err != nil { // respect cancellation
		return Item{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

	}()
	return it, nil
}

func (m *Memory) Put(ctx context.Context, item Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[item.ID] = item
	return nil
}
