package main

import (
	"fmt"
	"time"
)

func main() {
	strings := make(chan string)
	nums := make(chan int)

	go func() {
		for {
			strings <- "Hello"
			time.Sleep(time.Millisecond * 200)
		}
	}()

	go func() {
		for {
			nums <- 1
			time.Sleep(time.Second * 2)
		}
	}()

	/*
		Sequential blocking: The loop first waits for
		a string from strings channel, then waits for a
		number from nums channel. It must receive both in
			that exact order before starting the next
		iteration
	*/
	for {
		fmt.Println(<-strings)
		// This will block waiting
		fmt.Println(<-nums)
	}

	/*
		 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			// can receive from whichever channel
			// has data available first
			for {
				select {
				case msg := <-strings:
					fmt.Println(msg)
				case msg := <-nums:
					fmt.Println(msg)
				case <-ctx.Done():
					fmt.Println("Context cancelled:", ctx.Err())
					return
				}
			}
	*/
}
