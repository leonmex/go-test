// How to solve the communication between goroutines problem that wait groups can't solve?

package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func sendMessage(ch chan<- string, num int, wg *sync.WaitGroup) {
	fmt.Printf("Sending message %d\n", num)

	time.Sleep(time.Second * time.Duration(num)) // Simulate some work
	ch <- fmt.Sprintf("✅ Message %d sent!", num)
	wg.Done()
}

func receiveMessage(msgs <-chan string) {
	fmt.Println("Waiting for message")

	for msg := range msgs {
		fmt.Println("Received:", msg)
	}
}

func main() {
	msgs := make(chan string)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go sendMessage(msgs, 1, &wg)
	go sendMessage(msgs, 2, &wg)

	go func() {
		receiveMessage(msgs)
	}()

	wg.Wait()
	close(msgs)

	log.Println("Work done")

	fmt.Scanln() // keep the program running
}
