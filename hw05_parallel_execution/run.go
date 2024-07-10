package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
// If m <= 0, then the function should ignore errors.
func Run(tasks []Task, n, m int) error {
	taskChan := make(chan Task, len(tasks))
	errorCount := 0
	stop := atomic.Bool{}
	lock := sync.Mutex{}
	var wg sync.WaitGroup

	worker := func(id int) {
		defer wg.Done()
		defer fmt.Printf("worker %d done\n", id)

		for task := range taskChan {
			if stop.Load() {
				return
			}
			err := task()
			fmt.Printf("worker %d: finished task\n", id)
			if err != nil {
				lock.Lock()
				errorCount++
				if m > 0 && errorCount >= m {
					stop.Store(true)
					lock.Unlock()
					return
				}
				lock.Unlock()
			}
		}
	}

	// start workers
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(i)
	}

	// launch tasks
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// wait for workers to finish
	wg.Wait()

	if stop.Load() {
		return ErrErrorsLimitExceeded
	}
	return nil
}
