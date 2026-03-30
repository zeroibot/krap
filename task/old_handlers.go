package task

import (
	"github.com/zeroibot/fn/ds"
	"github.com/zeroibot/krap/authz"
)

// Common: create FullTask Handler
func fullTaskHandler[A Actor, T any, P any](task *FullTask[A, T], cfg *taskConfig[A, T, P]) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check Authorization if not deferred
		if !task.DeferActionCheck {
			err = authz.CheckActionAllowedFor(rq, (*actor).GetRole())
		}
		var item *T = nil
		if err == nil {
			// Perform action if authorized
			item, err = task.Fn(rq, actor)
		}
		cfg.outputFn(p, item, rq, err)
	}
}

// Common: create CodedFullTask Handler
func codedFullTaskHandler[A Actor, T any, P any](task *CodedFullTask[A, T], cfg *taskConfig[A, T, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		if task.Validator == nil {
			cfg.errorFn(p, rq, errMissingHook)
			return
		}
		// Get code and call validator
		code := codeFn(p)
		err = task.Validator(rq, actor, code)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Perform action
		item, err := task.Fn(rq, actor)
		cfg.outputFn(p, item, rq, err)
	}
}

// Common: create ActionTask Handler
func actionTaskHandler[A Actor, P any](task *ActionTask[A], cfg *actionConfig[A, P]) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check Authorization
		err = authz.CheckActionAllowedFor(rq, (*actor).GetRole())
		if err == nil {
			// Perform action if authorized
			err = task.Fn(rq, actor)
		}
		cfg.outputFn(p, rq, err)
	}
}

// Common: create CodedActionTask Handler
func codedActionTaskHandler[A Actor, P any](task *CodedActionTask[A], cfg *actionConfig[A, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		if task.Validator == nil {
			cfg.errorFn(p, rq, errMissingHook)
			return
		}
		// Get code and call validator
		code := codeFn(p)
		err = task.Validator(rq, actor, code)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Perform action
		err = task.Fn(rq, actor)
		cfg.outputFn(p, rq, err)
	}
}

// Common: create TypedActionTask Handler
func typedActionTaskHandler[A Actor, T any, P any](task *TypedActionTask[A, T], cfg *actionConfig[A, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		if task.Validator == nil {
			cfg.errorFn(p, rq, errMissingHook)
			return
		}
		// Get code and call validator
		code := codeFn(p)
		err = task.Validator(rq, actor, task.Schema, task.Store, code)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Perform action
		err = task.Fn(rq, actor)
		cfg.outputFn(p, rq, err)
	}
}

// Common: create ListTask Handler
func dataTaskHandler[T any, P any](task *DataTask[T], cfg *dataConfig[T, P]) func(P) {
	return func(p P) {
		// Initialize
		rq, authToken, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check Authorization
		err = authz.CheckActionAllowedFor(rq, authToken.Type)
		var data *T
		if err == nil {
			// Get data if authorized
			data, err = task.Fn(rq)
		}
		cfg.outputFn(p, data, rq, err)
	}
}

// Common: create CodedDataTask Handler
func codedDataTaskHandler[A Actor, T any, P any](task *CodedDataTask[A, T], cfg *codedDataConfig[A, T, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		if task.Validator == nil {
			cfg.errorFn(p, rq, errMissingHook)
			return
		}
		// Get code and call validator
		code := codeFn(p)
		err = task.Validator(rq, actor, code)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Get data after passing all checks
		data, err := task.Fn(rq)
		cfg.outputFn(p, data, rq, err)
	}
}

// Common: create ListTask Handler
func listTaskHandler[T any, P any](task *ListTask[T], cfg *listConfig[T, P]) func(P) {
	return func(p P) {
		// Initialize
		rq, authToken, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check Authorization
		err = authz.CheckActionAllowedFor(rq, authToken.Type)
		var list *ds.List[*T]
		if err == nil {
			// Get data if authorized
			list, err = task.Fn(rq)
		}
		cfg.outputFn(p, list, rq, err)
	}
}

// Common: create CodedListTask Handler
func codedListTaskHandler[A Actor, T any, P any](task *CodedListTask[A, T], cfg *codedListConfig[A, T, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		if task.Validator == nil {
			cfg.errorFn(p, rq, errMissingHook)
			return
		}
		// Get code and call validator
		code := codeFn(p)
		err = task.Validator(rq, actor, code)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Get list after passing all checks
		list, err := task.Fn(rq)
		cfg.outputFn(p, list, rq, err)
	}
}

// Common: create ViewTask Handler
func viewTaskHandler[T any, P any](task *ViewTask[T], cfg *viewConfig[T, P]) func(P) {
	return func(p P) {
		// Initialize
		rq, authToken, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check Authorization
		err = authz.CheckActionAllowedFor(rq, authToken.Type)
		var item *T
		if err == nil {
			// Get item if authorized
			item, err = task.Fn(rq)
		}
		cfg.outputFn(p, item, rq, err)
	}
}

// Common: create CodedViewTask Handler
func codedViewTaskHandler[A Actor, T any, P any](task *CodedViewTask[A, T], cfg *codedViewConfig[A, T, P], codeFn func(P) string) func(P) {
	return func(p P) {
		// Initialize
		rq, actor, err := cfg.initialize(p)
		if err != nil {
			cfg.errorFn(p, rq, err)
			return
		}
		// Check validator, if it exists
		if task.Validator != nil {
			code := codeFn(p)
			err = task.Validator(rq, actor, code)
			if err != nil {
				cfg.errorFn(p, rq, err)
				return
			}
		}
		// Get item
		item, err := task.Fn(rq)
		cfg.outputFn(p, item, rq, err)
	}
}
