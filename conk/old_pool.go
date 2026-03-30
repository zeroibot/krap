package conk

import (
	"sync"

	"github.com/zeroibot/rdb/ze"
)

type (
	WorkerFn[T any]        = func(T) error
	RequestWorkerFn[T any] = func(*ze.Request, T) error
)

// Perform work (func(T) error) sequentially
func WorkersLinear[T any](items []T, fn WorkerFn[T]) *Result {
	result := newResult()
	for i, item := range items {
		err := fn(item)
		if err == nil {
			result.Success += 1
		} else {
			result.Errors[i] = err
		}
	}
	return result
}

// Perform work (func(T) error) concurrently
func Workers[T any](items []T, fn WorkerFn[T], numWorkers int) *Result {
	// Worker function
	worker := func(inputCh <-chan inputData[T], resultCh chan<- resultData) {
		for input := range inputCh {
			err := fn(input.item)
			resultCh <- resultData{input.index, err}
		}
	}
	return workerPool(items, worker, numWorkers)
}

// Perform request work (func(*Request, T) error) sequentially
func RequestWorkersLinear[T any](rq *ze.Request, items []T, fn RequestWorkerFn[T]) *Result {
	result := newResult()
	for i, item := range items {
		err := fn(rq, item)
		if err == nil {
			result.Success += 1
		} else {
			result.Errors[i] = err
		}
	}
	return result
}

// Perform request work (func(*Request, T) error) concurrently
func RequestWorkers[T any](rq *ze.Request, items []T, fn RequestWorkerFn[T], numWorkers int) *Result {
	// Worker function
	worker := func(inputCh <-chan inputData[T], resultCh chan<- resultData) {
		srq := rq.SubRequest()
		for input := range inputCh {
			err := fn(srq, input.item)
			resultCh <- resultData{input.index, err}
		}
		rq.MergeLogs(srq)
	}

	return workerPool(items, worker, numWorkers)
}

// Common: worker pool
type workerFn[T any] = func(<-chan inputData[T], chan<- resultData)

func workerPool[T any](items []T, worker workerFn[T], numWorkers int) *Result {
	numWorkers = max(numWorkers, 1) // lower-bound: 1 worker
	inputCh := make(chan inputData[T])
	resultCh := make(chan resultData, numWorkers) // need buffered, otherwise deadlocks

	// Spawn workers
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			worker(inputCh, resultCh)
		})
	}

	// Feed input items to input channel
	go func() {
		for i, item := range items {
			inputCh <- inputData[T]{i, item}
		}
		close(inputCh)
	}()

	// Wait for all workers to finish and close result channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Get results
	result := newResult()
	for out := range resultCh {
		if out.err == nil {
			result.Success += 1
		} else {
			result.Errors[out.index] = out.err
		}
	}
	return result
}
