package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type actionConfig[A Actor, P any] struct {
	*baseActorConfig[A, P]
	outputFn func(P, *ze.Request, error)
}

type ActionTask[A Actor] struct {
	*BaseTask[A]
	Fn ActionFn[A]
}

type CodedActionTask[A Actor] struct {
	*ActionTask[A]
	Validator HookFn[A]
	CodeIndex int
}

type TypedActionTask[A Actor, T any] struct {
	*ActionTask[A]
	Validator TypedHookFn[A, T]
	CodeIndex int
	*ze.Schema[T]
	Store Store[T]
}

// Create cmd actionConfig
func cmdActionConfig[A Actor](task *BaseTask[A]) *actionConfig[A, []string] {
	cfg := &actionConfig[A, []string]{
		baseActorConfig: &baseActorConfig[A, []string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, rq *ze.Request, err error) {
		sys.DisplayOutput(rq, err)
	}
	return cfg
}

// Create web actionConfig
func webActionConfig[A Actor](task *BaseTask[A]) *actionConfig[A, *gin.Context] {
	cfg := &actionConfig[A, *gin.Context]{
		baseActorConfig: &baseActorConfig[A, *gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendActionError
	cfg.outputFn = web.SendActionResponse
	return cfg
}

// Creates new ActionTask
func NewActionTask[A Actor](action, item string, fn ActionFn[A]) *ActionTask[A] {
	task := &ActionTask[A]{
		BaseTask: &BaseTask[A]{},
	}
	task.Action = action
	task.Target = item
	task.Fn = fn
	return task
}

// Creates new CodedActionTask
func NewCodedActionTask[A Actor](action, item string, fn ActionFn[A], codeIndex int) *CodedActionTask[A] {
	task := &CodedActionTask[A]{
		ActionTask: NewActionTask(action, item, fn),
	}
	task.CodeIndex = codeIndex
	return task
}

// Creates new TypedActionTask
func NewTypedActionTask[A Actor, T any](action, item string, fn ActionFn[A], codeIndex int, schema *ze.Schema[T], store Store[T]) *TypedActionTask[A, T] {
	task := &TypedActionTask[A, T]{
		ActionTask: NewActionTask(action, item, fn),
	}
	task.CodeIndex = codeIndex
	task.Schema = schema
	task.Store = store
	return task
}

// Attach HookFn to CodedActionTask
func (task *CodedActionTask[A]) WithValidator(hookFn HookFn[A]) {
	task.Validator = hookFn
}

// Attach TypedHookFn to TypedActionTask
func (task *TypedActionTask[A, T]) WithValidator(hookFn TypedHookFn[A, T]) {
	task.Validator = hookFn
}

// ActionTask CmdHandler
func (task ActionTask[A]) CmdHandler() root.CmdHandler {
	return actionTaskHandler(&task, cmdActionConfig(task.BaseTask))
}

// ActionTask WebHandler
func (task ActionTask[A]) WebHandler() gin.HandlerFunc {
	return actionTaskHandler(&task, webActionConfig(task.BaseTask))
}

// CodedActionTask CmdHandler
func (task CodedActionTask[A]) CmdHandler() root.CmdHandler {
	codeFn := func(args []string) string {
		return getCode(args, task.CodeIndex)
	}
	return codedActionTaskHandler(&task, cmdActionConfig(task.BaseTask), codeFn)
}

// CodedActionTask WebHandler
func (task CodedActionTask[A]) WebHandler() gin.HandlerFunc {
	return codedActionTaskHandler(&task, webActionConfig(task.BaseTask), sys.WebCodeParam)
}

// TypedActionTask CmdHandler
func (task TypedActionTask[A, T]) CmdHandler() root.CmdHandler {
	codeFn := func(args []string) string {
		return getCode(args, task.CodeIndex)
	}
	return typedActionTaskHandler(&task, cmdActionConfig(task.BaseTask), codeFn)
}

// TypedActionTask WebHandler
func (task TypedActionTask[A, T]) WebHandler() gin.HandlerFunc {
	return typedActionTaskHandler(&task, webActionConfig(task.BaseTask), sys.WebCodeParam)
}
