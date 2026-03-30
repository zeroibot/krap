package task

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/ds"
	"github.com/zeroibot/krap/authn"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/rdb/ze"
)

// Note: A type is for Actor

type Actor interface {
	GetRole() string
}

type (
	TaskFn[A Actor, T any] = func(*ze.Request, *A) (*T, error)
	ActionFn[A Actor]      = func(*ze.Request, *A) error
	DataFn[T any]          = func(*ze.Request) (*T, error)
	ListFn[T any]          = func(*ze.Request) (*ds.List[*T], error)
)

type BaseTask[A Actor] struct {
	ze.Task
	CmdDecorator CmdDecorator[A]
	WebDecorator WebDecorator[A]
}

type BaseTokenTask struct {
	ze.Task
	CmdDecorator CmdTokenDecorator
	WebDecorator WebTokenDecorator
}

type BaseDataTask[A Actor] struct {
	ze.Task
	CmdDecorator CmdDataDecorator[A]
	WebDecorator WebDataDecorator[A]
}

type BaseDataTokenTask struct {
	ze.Task
	CmdDecorator CmdDataTokenDecorator
	WebDecorator WebDataTokenDecorator
}

type (
	Decorator[A Actor, P any] = func(*ze.Request, P) (*A, error)
	CmdDecorator[A Actor]     = func(*ze.Request, []string) (*A, error)
	WebDecorator[A Actor]     = func(*ze.Request, *gin.Context) (*A, error)

	TokenDecorator[P any] = func(*ze.Request, P) (*authn.Token, error)
	CmdTokenDecorator     = func(*ze.Request, []string) (*authn.Token, error)
	WebTokenDecorator     = func(*ze.Request, *gin.Context) (*authn.Token, error)

	DataDecorator[A Actor, P any] = func(*ze.Request, P) (*A, bool, error)
	CmdDataDecorator[A Actor]     = func(*ze.Request, []string) (*A, bool, error)
	WebDataDecorator[A Actor]     = func(*ze.Request, *gin.Context) (*A, bool, error)

	DataTokenDecorator[P any] = func(*ze.Request, P) (*authn.Token, bool, error)
	CmdDataTokenDecorator     = func(*ze.Request, []string) (*authn.Token, bool, error)
	WebDataTokenDecorator     = func(*ze.Request, *gin.Context) (*authn.Token, bool, error)
)

type CmdHandler interface {
	CmdHandler() root.CmdHandler
}

type WebHandler interface {
	WebHandler() gin.HandlerFunc
}

type Handler interface {
	CmdHandler
	WebHandler
}

type (
	Router    = map[string]Handler
	AddRouter = map[[2]string]Handler
)

// Request, Params, Actor, Code, ID
type HookFn[A Actor] = func(*ze.Request, *A, string) error

// Request, Params, Actor, Schema, Code, ID
type TypedHookFn[A Actor, T any] = func(*ze.Request, *A, *ze.Schema[T], Store[T], string) error

type Store[T any] interface {
	GetByCode(string) (*T, bool)
	Add(*T)
}
