// How to solve the communication between goroutines problem that wait groups can't solve?

package main

import (
	"fmt"
	"time"
)

func sendMessage(num int, msgChan chan<- string) {
	fmt.Printf("Sending message %d\n", num)

	time.Sleep(time.Second * time.Duration(num)) // Simulate some work

	msg := fmt.Sprintf("✅ Message %d sent!", num)

	msgChan <- msg
}

func receiveMessage(msgs <-chan string) []string {
	fmt.Println("Waiting for message")

	var got []string
	for msg := range msgs {
		fmt.Println("Received:", msg)
		got = append(got, msg)
	}
	return got
}

// Known problem (fixed in 01-channels-and-waitgroups): nobody closes msgChan before
// receiveMessage ranges over it, so after the 2 messages the program dies with
// "fatal error: all goroutines are asleep - deadlock!".
func main() {
	msgChan := make(chan string)

	go sendMessage(2, msgChan)
	go sendMessage(3, msgChan)

	receiveMessage(msgChan)

	close(msgChan)
}
