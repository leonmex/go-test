package main

import (
	"testing"
	"time"
)

// Concurrent: total time is the MAX prep time, not the sum, and every order finishes.
func TestProcessAll_WaitsForAllOrdersConcurrently(t *testing.T) {
	orders := []Order{
		{TableNumber: 1, PrepTime: 50 * time.Millisecond},
		{TableNumber: 2, PrepTime: 80 * time.Millisecond},
		{TableNumber: 3, PrepTime: 30 * time.Millisecond},
		{TableNumber: 4, PrepTime: 60 * time.Millisecond},
	}

	start := time.Now()
	results := processAll(orders)
	elapsed := time.Since(start)

	if elapsed < 80*time.Millisecond {
		t.Errorf("returned before the slowest order: %v", elapsed)
	}
	if elapsed >= 220*time.Millisecond {
		t.Errorf("expected about 80ms (max), not the 220ms sum: got %v", elapsed)
	}
	for i, r := range results {
		if r != "order has been done" {
			t.Errorf("order %d: expected done, got %q", i, r)
		}
	}
}
