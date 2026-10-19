package main

import (
	"slices"
	"testing"
	"time"
)

// Same shape as main (8s, 5s, 5s, 3s with 3 waiters), scaled to milliseconds.
var testOrders = []Order{
	{TableNumber: 1, PrepTime: 80 * time.Millisecond},
	{TableNumber: 2, PrepTime: 50 * time.Millisecond},
	{TableNumber: 3, PrepTime: 50 * time.Millisecond},
	{TableNumber: 4, PrepTime: 30 * time.Millisecond},
}

func TestRunPool(t *testing.T) {
	tests := []struct {
		name    string
		waiters int
		minTime time.Duration
		maxTime time.Duration
	}{
		// table 4 waits for a free waiter (50ms) then cooks 30ms -> 80ms total, same as table 1
		{name: "3 waiters", waiters: 3, minTime: 80 * time.Millisecond, maxTime: 150 * time.Millisecond},
		{name: "1 waiter is sequential", waiters: 1, minTime: 210 * time.Millisecond, maxTime: 400 * time.Millisecond},
		{name: "more waiters than orders", waiters: 10, minTime: 80 * time.Millisecond, maxTime: 150 * time.Millisecond},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			served := runPool(testOrders, tc.waiters)
			elapsed := time.Since(start)

			slices.Sort(served)
			if !slices.Equal(served, []int{1, 2, 3, 4}) {
				t.Errorf("expected every table served once, got %v", served)
			}
			if elapsed < tc.minTime || elapsed > tc.maxTime {
				t.Errorf("expected between %v and %v, got %v", tc.minTime, tc.maxTime, elapsed)
			}
		})
	}
}
