package task

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/lang"
	"github.com/zeroibot/krap/authn"
	"github.com/zeroibot/krap/authz"
	"github.com/zeroibot/rdb/ze"
)

// const actionGlue string = "-"

var (
	ErrInvalidActor  = errors.New("public: Invalid actor")
	ErrInvalidOption = errors.New("public: Invalid option")
	errMissingHook   = errors.New("missing hook")
)

// Attach CmdDecorator to BaseTask
func (t *BaseTask[A]) WithCmd(cmdDecorator CmdDecorator[A]) {
	t.CmdDecorator = cmdDecorator
}

// Attach WebDecorator to BaseTask
func (t *BaseTask[A]) WithWeb(webDecorator WebDecorator[A]) {
	t.WebDecorator = webDecorator
}

// Attach CmdDecorator to BaseTokenTask
func (t *BaseTokenTask) WithCmd(cmdDecorator CmdTokenDecorator) {
	t.CmdDecorator = cmdDecorator
}

// Attach WebDecorator to BaseTokenTask
func (t *BaseTokenTask) WithWeb(webDecorator WebTokenDecorator) {
	t.WebDecorator = webDecorator
}

// Attach CmdDecorator to BaseDataTask
func (t *BaseDataTask[A]) WithCmd(cmdDecorator CmdDataDecorator[A]) {
	t.CmdDecorator = cmdDecorator
}

// Attach WebDecorator to BaseDataTask
func (t *BaseDataTask[A]) WithWeb(webDecorator WebDataDecorator[A]) {
	t.WebDecorator = webDecorator
}

// Attach CmdDecorator to BaseDataTokenTask
func (t *BaseDataTokenTask) WithCmd(cmdDecorator CmdDataTokenDecorator) {
	t.CmdDecorator = cmdDecorator
}

// Attach WebDecorator to BaseDataTokenTask
func (t *BaseDataTokenTask) WithWeb(webDecorator WebDataTokenDecorator) {
	t.WebDecorator = webDecorator
}

// Initialize BaseTask CmdHandler
func (task BaseTask[A]) cmdInitialize(args []string) (*ze.Request, *A, error) {
	return initialize(&task, args, task.CmdDecorator)
}

// Initialize BaseTask WebHandler
func (task BaseTask[A]) webInitialize(c *gin.Context) (*ze.Request, *A, error) {
	return initialize(&task, c, task.WebDecorator)
}

// Common BaseTask initialize
func initialize[A Actor, P any](task *BaseTask[A], p P, decorator Decorator[A, P]) (*ze.Request, *A, error) {
	// Create request
	name := task.FullName()
	rq, err := ze.NewRequest(name)
	if err != nil {
		return rq, nil, err
	}
	// Attach task to request
	rq.Task = task.Task
	// Decorate params
	actor, err := decorator(rq, p)
	if err == nil && actor == nil {
		err = ErrInvalidActor
	}
	if err != nil {
		return rq, nil, err
	}
	return rq, actor, nil
}

// Initialize BaseTokenTask CmdHandler
func (task BaseTokenTask) cmdInitialize(args []string) (*ze.Request, *authn.Token, error) {
	return initializeToken(&task, args, task.CmdDecorator)
}

// Initialize BaseTokenTask WebHandler
func (task BaseTokenTask) webInitialize(c *gin.Context) (*ze.Request, *authn.Token, error) {
	return initializeToken(&task, c, task.WebDecorator)
}

// Common BaseTokenTask initialize
func initializeToken[P any](task *BaseTokenTask, p P, decorator TokenDecorator[P]) (*ze.Request, *authn.Token, error) {
	// Create request
	name := task.FullName()
	rq, err := ze.NewRequest(name)
	if err != nil {
		return rq, nil, err
	}
	// Attach task to request
	rq.Task = task.Task
	// Decorate params
	authToken, err := decorator(rq, p)
	if err == nil && authToken == nil {
		err = authn.ErrInvalidSession
	}
	if err != nil {
		return rq, nil, err
	}
	return rq, authToken, nil
}

// Initialize BaseDataTask CmdHandler
func (task BaseDataTask[A]) cmdInitialize(args []string) (*ze.Request, *A, error) {
	return initializeData(&task, args, task.CmdDecorator)
}

// Initialize BaseDataTask WebHandler
func (task BaseDataTask[A]) webInitialize(c *gin.Context) (*ze.Request, *A, error) {
	return initializeData(&task, c, task.WebDecorator)
}

// Common BaseDataTask initialize
func initializeData[A Actor, P any](task *BaseDataTask[A], p P, decorator DataDecorator[A, P]) (*ze.Request, *A, error) {
	// Create request
	rq, err := ze.NewRequest(task.Target) // temporary name, updated below
	if err != nil {
		return rq, nil, err
	}
	// Decorate params
	actor, mustBeActive, err := decorator(rq, p)
	if err == nil && actor == nil {
		err = ErrInvalidActor
	}
	if err != nil {
		return rq, nil, err
	}
	// Attach action, item to request
	rq.Action = lang.Ternary(mustBeActive, authz.VIEW, authz.ROWS)
	rq.Target = task.Target
	rq.Name = rq.Task.FullName()
	return rq, actor, nil
}

// Initialize BaseDataTokenTask CmdHandler
func (task BaseDataTokenTask) cmdInitialize(args []string) (*ze.Request, *authn.Token, error) {
	return initializeDataToken(&task, args, task.CmdDecorator)
}

// Initialize BaseDataTokenTask WebHandler
func (task BaseDataTokenTask) webInitialize(c *gin.Context) (*ze.Request, *authn.Token, error) {
	return initializeDataToken(&task, c, task.WebDecorator)
}

// Common BaseDataTokenTask initialize
func initializeDataToken[P any](task *BaseDataTokenTask, p P, decorator DataTokenDecorator[P]) (*ze.Request, *authn.Token, error) {
	// Create request
	rq, err := ze.NewRequest(task.Target) // temporary name, updated below
	if err != nil {
		return rq, nil, err
	}
	// Decorate params
	authnToken, mustBeActive, err := decorator(rq, p)
	if err == nil && authnToken == nil {
		err = authn.ErrInvalidSession
	}
	if err != nil {
		return rq, nil, err
	}
	// Attach action, item to request
	rq.Action = lang.Ternary(mustBeActive, authz.VIEW, authz.ROWS)
	rq.Target = task.Target
	rq.Name = rq.Task.FullName()
	return rq, authnToken, nil
}

// Get code from args string list on index
func getCode(args []string, index int) string {
	code := ""
	if index < len(args) {
		code = strings.ToUpper(args[index])
	}
	return code
}
