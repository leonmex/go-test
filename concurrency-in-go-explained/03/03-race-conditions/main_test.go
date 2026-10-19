package main

import "testing"

// Run with -race: removing db.Lock() in add makes this test fail with "DATA RACE"
// (or crash with "concurrent map writes").
func TestFill_ConcurrentWritersAreSafe(t *testing.T) {
	db := &database{users: make(map[int]string)}

	fill(db, 8, 1000)

	if got := db.count(); got != 1000 {
		t.Errorf("expected 1000 users, got %d", got)
	}
	if db.users[999] != "user-999" {
		t.Errorf("unexpected value %q", db.users[999])
	}
}
