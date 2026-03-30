package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/ds"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type listConfig[T any, P any] struct {
	*baseTokenConfig[P]
	outputFn func(P, *ds.List[*T], *ze.Request, error)
}

type codedListConfig[A Actor, T any, P any] struct {
	*baseActorConfig[A, P]
	outputFn func(P, *ds.List[*T], *ze.Request, error)
}

type ListTask[T any] struct {
	*BaseDataTokenTask
	Fn ListFn[T]
}

type CodedListTask[A Actor, T any] struct {
	*BaseDataTask[A]
	Fn        ListFn[T]
	Validator HookFn[A]
	CodeIndex int
}

// Creates new ListTask
func NewListTask[T any](item string, fn ListFn[T]) *ListTask[T] {
	task := &ListTask[T]{
		BaseDataTokenTask: &BaseDataTokenTask{},
	}
	task.Target = item
	task.Fn = fn
	return task
}

// Creates new CodedListTask
func NewCodedListTask[A Actor, T any](item string, fn ListFn[T], codeIndex int) *CodedListTask[A, T] {
	task := &CodedListTask[A, T]{
		BaseDataTask: &BaseDataTask[A]{},
	}
	task.Target = item
	task.Fn = fn
	task.CodeIndex = codeIndex
	return task
}

// Attach HookFn to CodedListTask
func (task *CodedListTask[A, T]) WithValidator(hookFn HookFn[A]) {
	task.Validator = hookFn
}

// ListTask CmdHandler
func (task ListTask[T]) CmdHandler() root.CmdHandler {
	cfg := &listConfig[T, []string]{
		baseTokenConfig: &baseTokenConfig[[]string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, list *ds.List[*T], rq *ze.Request, err error) {
		sys.DisplayList(list, rq, err)
	}
	return listTaskHandler(&task, cfg)
}

// ListTask WebHandler
func (task ListTask[T]) WebHandler() gin.HandlerFunc {
	cfg := &listConfig[T, *gin.Context]{
		baseTokenConfig: &baseTokenConfig[*gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return listTaskHandler(&task, cfg)
}

// CodedListTask CmdHandler
func (task CodedListTask[A, T]) CmdHandler() root.CmdHandler {
	cfg := &codedListConfig[A, T, []string]{
		baseActorConfig: &baseActorConfig[A, []string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, list *ds.List[*T], rq *ze.Request, err error) {
		sys.DisplayList(list, rq, err)
	}
	codeFn := func(args []string) string {
		return getCode(args, task.CodeIndex)
	}
	return codedListTaskHandler(&task, cfg, codeFn)
}

// CodedListTask WebHandler
func (task CodedListTask[A, T]) WebHandler() gin.HandlerFunc {
	cfg := &codedListConfig[A, T, *gin.Context]{
		baseActorConfig: &baseActorConfig[A, *gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return codedListTaskHandler(&task, cfg, sys.WebCodeOption)
}
