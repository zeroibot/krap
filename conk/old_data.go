package conk

import (
	"sync"

	"github.com/zeroibot/rdb/ze"
)

type (
	DataWorkerFn[I any, O any]        = func(I) (O, error)
	RequestDataWorkerFn[I any, O any] = func(*ze.Request, I) (O, error)
)

// Perform data work (func(I) (O, error)) sequentially
func DataWorkersLinear[I any, O any](items []I, fn DataWorkerFn[I, O]) *DataResult[O] {
	result := newDataResult[O]()
	for i, item := range items {
		out, err := fn(item)
		if err == nil {
			result.Success += 1
			result.Output[i] = out
		} else {
			result.Errors[i] = err
		}
	}
	return result
}

// Perform data work (func(I) (O, error)) concurrently
func DataWorkers[I any, O any](items []I, fn DataWorkerFn[I, O], numWorkers int) *DataResult[O] {
	// Work function
	work := func(index int, item I, workerCh chan<- outputData[O]) {
		out, err := fn(item)
		workerCh <- outputData[O]{index, out, err}
	}
	return dataFanOutIn(items, work, numWorkers)
}

// Perform request data work (func(*Request, I) (O, error)) sequentially
func RequestDataWorkersLinear[I any, O any](rq *ze.Request, items []I, fn RequestDataWorkerFn[I, O]) *DataResult[O] {
	result := newDataResult[O]()
	for i, item := range items {
		out, err := fn(rq, item)
		if err == nil {
			result.Success += 1
			result.Output[i] = out
		} else {
			result.Errors[i] = err
		}
	}
	return result
}

// Perform request data work (func(I) (O, error)) concurrently
func RequestDataWorkers[I any, O any](rq *ze.Request, items []I, fn RequestDataWorkerFn[I, O], numWorkers int) *DataResult[O] {
	// Work function
	work := func(index int, item I, workerCh chan<- outputData[O]) {
		srq := rq.SubRequest()
		out, err := fn(srq, item)
		workerCh <- outputData[O]{index, out, err}
		rq.MergeLogs(srq)
	}
	return dataFanOutIn(items, work, numWorkers)
}

// Common: data worker pool
type dataWorkerFn[I any, O any] = func(int, I, chan<- outputData[O])

func dataFanOutIn[I any, O any](items []I, work dataWorkerFn[I, O], numWorkers int) *DataResult[O] {
	numWorkers = max(numWorkers, 1) // lower-bound: 1 worker
	numItems := len(items)

	// Fan-out the workload
	channels := make([]<-chan outputData[O], numWorkers)
	for workerID := range numWorkers {
		workerCh := make(chan outputData[O])
		channels[workerID] = workerCh

		// Start worker goroutine
		// Labor division scheme: Process i % numWorkers for each worker
		go func() {
			for i := workerID; i < numItems; i += numWorkers {
				work(i, items[i], workerCh)
			}
			close(workerCh)
		}()
	}

	// Fan-in the results
	var wg sync.WaitGroup
	outputCh := make(chan outputData[O], numWorkers) // buffered
	for _, workerCh := range channels {
		wg.Go(func() {
			for out := range workerCh {
				outputCh <- out
			}
		})
	}

	// Wait for all channels to close and close main output channel
	go func() {
		wg.Wait()
		close(outputCh)
	}()

	// Get results
	result := newDataResult[O]()
	for out := range outputCh {
		if out.err == nil {
			result.Success += 1
			result.Output[out.index] = out.item
		} else {
			result.Errors[out.index] = out.err
		}
	}
	return result
}
