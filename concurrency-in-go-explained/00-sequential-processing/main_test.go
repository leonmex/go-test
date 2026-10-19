package main

import (
	"testing"
	"time"
)

// Sequential: total time is the SUM of all prep times.
func TestProcessAll_TakesSumOfPrepTimes(t *testing.T) {
	orders := []Order{
		{TableNumber: 1, PrepTime: 20 * time.Millisecond},
		{TableNumber: 2, PrepTime: 30 * time.Millisecond},
		{TableNumber: 3, PrepTime: 10 * time.Millisecond},
	}

	start := time.Now()
	processAll(orders)
	elapsed := time.Since(start)

	if elapsed < 60*time.Millisecond {
		t.Errorf("expected at least 60ms (sum), got %v", elapsed)
	}
}
