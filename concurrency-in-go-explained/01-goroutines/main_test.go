package main

import (
	"testing"
	"time"
)

// The problem this step shows: dispatch returns before any order is ready,
// so without fmt.Scanln() main would exit and kill the goroutines.
func TestDispatch_ReturnsBeforeOrdersAreReady(t *testing.T) {
	orders := []Order{
		{TableNumber: 1, PrepTime: 200 * time.Millisecond},
		{TableNumber: 2, PrepTime: 300 * time.Millisecond},
	}

	start := time.Now()
	dispatch(orders)
	elapsed := time.Since(start)

	if elapsed >= 200*time.Millisecond {
		t.Errorf("expected dispatch to return immediately, took %v", elapsed)
	}
}
