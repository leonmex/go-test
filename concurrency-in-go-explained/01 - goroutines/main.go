// Step 01: Goroutines
// Problem: No synchronization between goroutines

package main

import (
	"fmt"
	"time"
)

type Order struct {
	TableNumber int
	PrepTime    time.Duration
}

func processOrder(order Order) {
	// Simulate cooking time
	fmt.Printf("Preparing order for table %d...\n", order.TableNumber)
	
	time.Sleep(order.PrepTime)
	
	fmt.Printf("Order ready for table %d!\n\n", order.TableNumber)
}

func main() {
	orders := []Order{
		{TableNumber: 1, PrepTime: 2 * time.Second},
		{TableNumber: 2, PrepTime: 3 * time.Second},
		{TableNumber: 3, PrepTime: 1 * time.Second},
		{TableNumber: 4, PrepTime: 2 * time.Second},
		{TableNumber: 5, PrepTime: 4 * time.Second},
	}

	for _, order := range orders {
		go processOrder(order)
	}

	fmt.Scanln()
}