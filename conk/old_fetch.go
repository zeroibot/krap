package conk

import (
	"github.com/zeroibot/rdb/ze"
	"golang.org/x/sync/errgroup"
)

type RequestListFn[T any] = func(*ze.Request) ([]T, error)

// Perform request list sequentially
func RequestListsLinear[T any](rq *ze.Request, fetchers ...RequestListFn[T]) ([]T, error) {
	allResults := make([]T, 0)
	for _, fetcher := range fetchers {
		results, err := fetcher(rq)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}
	return allResults, nil
}

// Perform request list concurrently
func RequestLists[T any](rq *ze.Request, fetchers ...RequestListFn[T]) ([]T, error) {
	var eg errgroup.Group
	resultCh := make(chan T, len(fetchers))

	for _, fetcher := range fetchers {
		eg.Go(func() error {
			srq := rq.SubRequest()
			defer rq.MergeLogs(srq)

			items, subErr := fetcher(srq)
			if subErr != nil {
				return subErr
			}
			for _, item := range items {
				resultCh <- item
			}
			return nil
		})
	}

	var err error
	go func() {
		err = eg.Wait()
		close(resultCh)
	}()

	results := make([]T, 0)
	for result := range resultCh {
		results = append(results, result)
	}

	return results, err
}
