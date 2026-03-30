package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type taskConfig[A Actor, T any, P any] struct {
	*baseActorConfig[A, P]
	outputFn func(P, *T, *ze.Request, error)
}

type FullTask[A Actor, T any] struct {
	*BaseTask[A]
	Fn               TaskFn[A, T]
	DeferActionCheck bool
}

type CodedFullTask[A Actor, T any] struct {
	*BaseTask[A]
	Fn        TaskFn[A, T]
	Validator HookFn[A]
	CodeIndex int
}

// Create cmd taskConfig
func cmdTaskConfig[A Actor, T any](task *BaseTask[A]) *taskConfig[A, T, []string] {
	cfg := &taskConfig[A, T, []string]{
		baseActorConfig: &baseActorConfig[A, []string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, item *T, rq *ze.Request, err error) {
		sys.DisplayData(item, rq, err)
	}
	return cfg
}

// Create web taskConfig
func webTaskConfig[A Actor, T any](task *BaseTask[A]) *taskConfig[A, T, *gin.Context] {
	cfg := &taskConfig[A, T, *gin.Context]{
		baseActorConfig: &baseActorConfig[A, *gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return cfg
}

// Creates new FullTask
func NewFullTask[A Actor, T any](action, item string, fn TaskFn[A, T], deferActionCheck bool) *FullTask[A, T] {
	task := &FullTask[A, T]{
		BaseTask: &BaseTask[A]{},
	}
	task.Action = action
	task.Target = item
	task.Fn = fn
	task.DeferActionCheck = deferActionCheck
	return task
}

// Creates new CodedFullTask
func NewCodedFullTask[A Actor, T any](action, item string, fn TaskFn[A, T], codeIndex int) *CodedFullTask[A, T] {
	task := &CodedFullTask[A, T]{
		BaseTask: &BaseTask[A]{},
	}
	task.Action = action
	task.Target = item
	task.Fn = fn
	task.CodeIndex = codeIndex
	return task
}

// Attach HookFn to CodedFullTask
func (task *CodedFullTask[A, T]) WithValidator(hookFn HookFn[A]) {
	task.Validator = hookFn
}

// FullTask CmdHandler
func (task FullTask[A, T]) CmdHandler() root.CmdHandler {
	return fullTaskHandler(&task, cmdTaskConfig[A, T](task.BaseTask))
}

// FullTask WebHandler
func (task FullTask[A, T]) WebHandler() gin.HandlerFunc {
	return fullTaskHandler(&task, webTaskConfig[A, T](task.BaseTask))
}

// CodedFullTask CmdHandler
func (task CodedFullTask[A, T]) CmdHandler() root.CmdHandler {
	codeFn := func(args []string) string {
		return getCode(args, task.CodeIndex)
	}
	return codedFullTaskHandler(&task, cmdTaskConfig[A, T](task.BaseTask), codeFn)
}

// CodedFullTask WebHandler
func (task CodedFullTask[A, T]) WebHandler() gin.HandlerFunc {
	return codedFullTaskHandler(&task, webTaskConfig[A, T](task.BaseTask), sys.WebCodeParam)
}
