// Step 02: Using WaitGroup for Coordination
// Better: We can track when all waiters are done!

package main

import (
	"fmt"
	"sync"
	"time"
)

type Order struct {
	TableNumber int
	PrepTime    time.Duration
}

func processOrder(waiterID int, order Order) string {
	// Simulate cooking time
	fmt.Printf("Waiter %d: Preparing order for table %d...\n", waiterID, order.TableNumber)

	time.Sleep(order.PrepTime)

	fmt.Printf("Waiter %d: Order ready for table %d!\n\n", waiterID, order.TableNumber)

	return "order has been done"
}

func main() {
	orders := []Order{
		{TableNumber: 1, PrepTime: 2 * time.Second},
		{TableNumber: 2, PrepTime: 3 * time.Second},
		{TableNumber: 3, PrepTime: 1 * time.Second},
		{TableNumber: 4, PrepTime: 2 * time.Second},
		{TableNumber: 5, PrepTime: 4 * time.Second},
	}

	wg := sync.WaitGroup{}

	for waiterID, order := range orders {
		wg.Add(1)

		go func() {
			defer wg.Done()
			processOrder(waiterID, order)
		}()
	}

	wg.Wait()

	fmt.Println("all orders done")
}
