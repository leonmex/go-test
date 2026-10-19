package store

import (
	"context"
	"errors"
	"testing"
)

var (
	book = Item{ID: "1", Name: "book"}
	pen  = Item{ID: "2", Name: "pen"}
)

func newSeededMemory(seed map[string]Item) *Memory {
	m := NewMemory()
	for k, v := range seed {
		m.items[k] = v
	}
	return m
}

func testContext(t *testing.T, cancelled bool) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	if cancelled {
		cancel()
	}
	return ctx
}

func TestMemory_Get(t *testing.T) {
	tests := []struct {
		name      string
		seed      map[string]Item
		id        string
		cancelled bool
		want      Item
		wantErr   error
	}{
		{
			name:    "Item not exist",
			seed:    nil,
			id:      "23",
			wantErr: ErrNotFound,
		},
		{
			name: "Item Exist",
			seed: map[string]Item{"1": book},
			id:   "1",
			want: book,
		},
		{
			name:      "Cancelled context",
			seed:      map[string]Item{"1": book},
			id:        "1",
			cancelled: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newSeededMemory(tc.seed)

			got, err := m.Get(testContext(t, tc.cancelled), tc.id)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}

			if got != tc.want {
				t.Errorf("Expected want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestNewMemory(t *testing.T) {
	store := NewMemory()
	if store == nil {
		t.Error("Expected a store, got nil")
	}
}
