package main

import (
	"fmt"
	"sync"
)

type database struct {
	// Maps in go are not thread safe
	users map[int]string
	sync.RWMutex
}

func (db *database) add(i int) {
	db.Lock()
	defer db.Unlock()
	db.users[i] = fmt.Sprintf("user-%d", i)
}

func (db *database) count() int {
	db.RLock()
	defer db.RUnlock()
	return len(db.users)
}

// fill runs `writers` goroutines that each add users 0..n-1 at the same time.
func fill(db *database, writers, n int) {
	wg := sync.WaitGroup{}
	wg.Add(writers)
	for range writers {
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				db.add(i)
			}
		}()
	}
	wg.Wait()
}

// go run -race main.go
func main() {
	db := database{
		users: make(map[int]string),
	}

	fill(&db, 2, 1000)

	fmt.Println(db.users)
}
