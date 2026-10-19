// How to solve the communication between goroutines problem that wait groups can't solve?

package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func sendMessage(ch chan<- string, num int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Sending message %d\n", num)

	time.Sleep(time.Second * time.Duration(num)) // Simulate some work
	ch <- fmt.Sprintf("✅ Message %d sent!", num)
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

// sendAll starts one sender per num; a closer goroutine closes the channel
// once every sender is done (wg.Wait), which ends the range in receiveMessage.
func sendAll(nums ...int) []string {
	msgs := make(chan string)

	wg := sync.WaitGroup{}
	wg.Add(len(nums))
	for _, n := range nums {
		go sendMessage(msgs, n, &wg)
	}

	go func() {
		wg.Wait()
		close(msgs)
	}()

	return receiveMessage(msgs)
}

func main() {
	sendAll(1, 2)

	log.Println("Work done")
}
