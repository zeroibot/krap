package conk

import (
	"context"
	"time"

	"github.com/zeroibot/rdb/ze"
	"golang.org/x/sync/errgroup"
)

type (
	ActionFn    = func() error
	ActionCtxFn = func(context.Context) error
	RequestFn   = func(*ze.Request) error
)

// Perform actions (func() error) sequentially
func ActionsLinear(actions ...ActionFn) error {
	for _, action := range actions {
		if err := action(); err != nil {
			return err
		}
	}
	return nil
}

// Perform actions (func() error) concurrently, return first error
// Waits for all actions to finish even if one has already returned error
func Actions(actions ...ActionFn) error {
	var eg errgroup.Group
	for _, action := range actions {
		eg.Go(action)
	}
	return eg.Wait()
}

// Perform contexed actions sequentially
func ActionsCtxLinear(ctxActions ...ActionCtxFn) error {
	ctx := context.Background()
	for _, ctxAction := range ctxActions {
		if err := ctxAction(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Perform contexed actions concurrently, return first error
// Actions have to check if context has cancelled to end early
func ActionsCtx(timeoutSeconds float64, ctxActions ...ActionCtxFn) error {
	ctx := context.Background()
	if timeoutSeconds > 0 {
		var cancel context.CancelFunc
		duration := time.Duration(timeoutSeconds) * time.Second
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, ctxAction := range ctxActions {
		eg.Go(func() error {
			return ctxAction(ctx)
		})
	}
	return eg.Wait()
}

// Perform requests (func(*Request) error) sequentially
func RequestsLinear(rq *ze.Request, requests ...RequestFn) error {
	for _, request := range requests {
		if err := request(rq); err != nil {
			return err
		}
	}
	return nil
}

// Perform requests (func(*Request) error) concurrently, return first error
// Waits for all requests to finish even if one has already returned error
func Requests(rq *ze.Request, requests ...RequestFn) error {
	var eg errgroup.Group
	for _, request := range requests {
		eg.Go(func() error {
			srq := rq.SubRequest()
			err := request(srq)
			rq.MergeLogs(srq)
			return err
		})
	}
	return eg.Wait()
}
