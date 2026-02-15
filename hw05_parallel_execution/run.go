package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	var wg sync.WaitGroup

	var errCount int64
	var errLimit int64

	if m <= 0 {
		errLimit = int64(1)
	} else {
		errLimit = int64(m)
	}

	// каналы.
	taskCh := make(chan Task)
	doneCh := make(chan struct{}, 1)

	// пишем задачи в канал - 1 горутина писатель.
	wg.Add(1)
	go func() {
		defer func() {
			close(taskCh)
			wg.Done()
		}()

		for _, task := range tasks {
			select {
			case <-doneCh:
				return
			default:
			}

			select {
			case <-doneCh:
				return
			case taskCh <- task:
			}
		}
	}()

	// запускаем n воркеров на чтение из канала - n горутин читателей.
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range taskCh {
				if err := task(); err != nil {
					if atomic.AddInt64(&errCount, 1) == errLimit {
						doneCh <- struct{}{}
					}
				}
			}
		}()
	}

	wg.Wait()

	// errCount без atomic: горутины завершились, за атомарность можно больше не переживать.
	if errCount >= errLimit {
		return ErrErrorsLimitExceeded
	}

	return nil
}
