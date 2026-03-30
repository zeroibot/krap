package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/krap/authz"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type viewConfig[T any, P any] struct {
	*baseTokenConfig[P]
	outputFn func(P, *T, *ze.Request, error)
}

type codedViewConfig[A Actor, T any, P any] struct {
	*baseActorConfig[A, P]
	outputFn func(P, *T, *ze.Request, error)
}

type ViewTask[T any] struct {
	*BaseTokenTask
	Fn DataFn[T]
}

type CodedViewTask[A Actor, T any] struct {
	*BaseTask[A]
	Fn        DataFn[T]
	Validator HookFn[A]
	CodeIndex int
}

// Creates new ViewTask
func NewViewTask[T any](item string, fn DataFn[T]) *ViewTask[T] {
	task := &ViewTask[T]{
		BaseTokenTask: &BaseTokenTask{},
	}
	task.Action = authz.VIEW
	task.Target = item
	task.Fn = fn
	return task
}

// Creates new CodedViewTask
func NewCodedViewTask[A Actor, T any](item string, fn DataFn[T], codeIndex int) *CodedViewTask[A, T] {
	task := &CodedViewTask[A, T]{
		BaseTask: &BaseTask[A]{},
	}
	task.Action = authz.VIEW
	task.Target = item
	task.Fn = fn
	task.CodeIndex = codeIndex
	return task
}

// Attach HookFn to CodedViewTask
func (task *CodedViewTask[A, T]) WithValidator(hookFn HookFn[A]) {
	task.Validator = hookFn
}

// ViewTask CmdHandler
func (task ViewTask[T]) CmdHandler() root.CmdHandler {
	cfg := &viewConfig[T, []string]{
		baseTokenConfig: &baseTokenConfig[[]string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, item *T, rq *ze.Request, err error) {
		sys.DisplayData(item, rq, err)
	}
	return viewTaskHandler(&task, cfg)
}

// ViewTask WebHandler
func (task ViewTask[T]) WebHandler() gin.HandlerFunc {
	cfg := &viewConfig[T, *gin.Context]{
		baseTokenConfig: &baseTokenConfig[*gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return viewTaskHandler(&task, cfg)
}

// CodedViewTask CmdHandler
func (task CodedViewTask[A, T]) CmdHandler() root.CmdHandler {
	cfg := &codedViewConfig[A, T, []string]{
		baseActorConfig: &baseActorConfig[A, []string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, item *T, rq *ze.Request, err error) {
		sys.DisplayData(item, rq, err)
	}
	codeFn := func(args []string) string {
		return getCode(args, task.CodeIndex)
	}
	return codedViewTaskHandler(&task, cfg, codeFn)
}

// CodedViewTask WebHandler
func (task CodedViewTask[A, T]) WebHandler() gin.HandlerFunc {
	cfg := &codedViewConfig[A, T, *gin.Context]{
		baseActorConfig: &baseActorConfig[A, *gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return codedViewTaskHandler(&task, cfg, sys.WebCodeParam)
}
