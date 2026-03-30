package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type dataConfig[T any, P any] struct {
	*baseTokenConfig[P]
	outputFn func(P, *T, *ze.Request, error)
}

type codedDataConfig[A Actor, T any, P any] struct {
	*baseActorConfig[A, P]
	outputFn func(P, *T, *ze.Request, error)
}

type DataTask[T any] struct {
	*BaseDataTokenTask
	Fn DataFn[T]
}

type CodedDataTask[A Actor, T any] struct {
	*BaseDataTask[A]
	Fn        DataFn[T]
	Validator HookFn[A]
	CodeIndex int
}

// Creates new DataTask
func NewDataTask[T any](item string, fn DataFn[T]) *DataTask[T] {
	task := &DataTask[T]{
		BaseDataTokenTask: &BaseDataTokenTask{},
	}
	task.Target = item
	task.Fn = fn
	return task
}

// Creates new CodedDataTask
func NewCodedDataTask[A Actor, T any](item string, fn DataFn[T], codeIndex int) *CodedDataTask[A, T] {
	task := &CodedDataTask[A, T]{
		BaseDataTask: &BaseDataTask[A]{},
	}
	task.Target = item
	task.Fn = fn
	task.CodeIndex = codeIndex
	return task
}

// Attach HookFn to CodedDataTask
func (task *CodedDataTask[A, T]) WithValidator(hookFn HookFn[A]) {
	task.Validator = hookFn
}

// DataTask CmdHandler
func (task DataTask[T]) CmdHandler() root.CmdHandler {
	cfg := &dataConfig[T, []string]{
		baseTokenConfig: &baseTokenConfig[[]string]{},
	}
	cfg.initialize = task.cmdInitialize
	cfg.errorFn = cmdDisplayError
	cfg.outputFn = func(args []string, data *T, rq *ze.Request, err error) {
		sys.DisplayData(data, rq, err)
	}
	return dataTaskHandler(&task, cfg)
}

// DataTask WebHandler
func (task DataTask[T]) WebHandler() gin.HandlerFunc {
	cfg := &dataConfig[T, *gin.Context]{
		baseTokenConfig: &baseTokenConfig[*gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return dataTaskHandler(&task, cfg)
}

// CodedDataTask CmdHandler
func (task CodedDataTask[A, T]) CmdHandler() root.CmdHandler {
	cfg := &codedDataConfig[A, T, []string]{
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
	return codedDataTaskHandler(&task, cfg, codeFn)
}

// CodedDataTask WebHandler
func (task CodedDataTask[A, T]) WebHandler() gin.HandlerFunc {
	cfg := &codedDataConfig[A, T, *gin.Context]{
		baseActorConfig: &baseActorConfig[A, *gin.Context]{},
	}
	cfg.initialize = task.webInitialize
	cfg.errorFn = web.SendDataError
	cfg.outputFn = web.SendDataResponse
	return codedDataTaskHandler(&task, cfg, sys.WebCodeOption)
}
