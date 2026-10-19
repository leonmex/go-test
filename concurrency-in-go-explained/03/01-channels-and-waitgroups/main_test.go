package main

import (
	"slices"
	"testing"
	"time"
)

func TestSendAll_ReceivesEveryMessageAndReturns(t *testing.T) {
	done := make(chan []string)
	go func() { done <- sendAll(0, 0, 0) }()

	select {
	case got := <-done:
		if len(got) != 3 {
			t.Fatalf("expected 3 messages, got %d: %v", len(got), got)
		}
		for _, m := range got {
			if m != "✅ Message 0 sent!" {
				t.Errorf("unexpected message %q", m)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("sendAll never returned: channel not closed")
	}
}

func TestSendAll_ArrivesInCompletionOrder(t *testing.T) {
	got := sendAll(1, 0) // num is seconds of work: 0 finishes first

	want := []string{"✅ Message 0 sent!", "✅ Message 1 sent!"}
	if !slices.Equal(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}
