package main

import (
	"slices"
	"testing"
	"time"
)

func TestSendMessage_DeliversFormattedMessage(t *testing.T) {
	ch := make(chan string, 1) // buffered: the send does not need a receiver yet
	sendMessage(0, ch)

	if got := <-ch; got != "✅ Message 0 sent!" {
		t.Errorf("unexpected message %q", got)
	}
}

func TestReceiveMessage_StopsWhenChannelIsClosed(t *testing.T) {
	ch := make(chan string, 2)
	ch <- "a"
	ch <- "b"
	close(ch) // without this, range never ends: the deadlock in main

	done := make(chan []string)
	go func() { done <- receiveMessage(ch) }()

	select {
	case got := <-done:
		if !slices.Equal(got, []string{"a", "b"}) {
			t.Errorf("expected [a b], got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("receiveMessage did not return after close")
	}
}
