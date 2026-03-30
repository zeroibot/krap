package task

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeroibot/krap/authn"
	"github.com/zeroibot/krap/root"
	"github.com/zeroibot/krap/sys"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

type baseActorConfig[A Actor, P any] struct {
	initialize func(P) (*ze.Request, *A, error)
	errorFn    func(P, *ze.Request, error)
}

type baseTokenConfig[P any] struct {
	initialize func(P) (*ze.Request, *authn.Token, error)
	errorFn    func(P, *ze.Request, error)
}

// Cmd DisplayError adapter
func cmdDisplayError(args []string, rq *ze.Request, err error) {
	sys.DisplayError(err)
}

// Create new CmdConfig using task.CmdHandler()
func Cmd[T CmdHandler](command string, minParams int, docs string, task T) *root.CmdConfig {
	return &root.CmdConfig{
		Command:   command,
		MinParams: minParams,
		Docs:      docs,
		Handler:   task.CmdHandler(),
	}
}

// Create gin.HandlerFunc from task.WebHandler()
func Web[T WebHandler](task T) gin.HandlerFunc {
	return task.WebHandler()
}

// Create new CmdConfig from Router
func CmdRoute[T CmdHandler](command string, minParams int, docs string, router map[string]T) *root.CmdConfig {
	// Build the handlers of each router option
	handlerOf := make(map[string]root.CmdHandler)
	for key, task := range router {
		handlerOf[key] = task.CmdHandler()
	}
	routerHandler := func(args []string) {
		option := strings.ToLower(args[0])
		handler, ok := handlerOf[option]
		if !ok {
			sys.DisplayError(ErrInvalidOption)
			return
		}
		handler(args)
	}
	return &root.CmdConfig{
		Command:   command,
		MinParams: minParams,
		Docs:      docs,
		Handler:   routerHandler,
	}
}

// Create gin.HandlerFunc from Router
func Fork[T WebHandler](router map[string]T, response *web.ResponseType) gin.HandlerFunc {
	// Build handlers of each router option
	handlerOf := make(map[string]gin.HandlerFunc)
	for key, task := range router {
		handlerOf[key] = task.WebHandler()
	}
	return func(c *gin.Context) {
		option := sys.WebForkParam(c)
		handler, ok := handlerOf[option]
		if !ok {
			response.SendErrorFn(c, nil, ErrInvalidOption)
			return
		}
		handler(c)
	}
}

// Create new CmdConfig from AddRouter
func AddRoute[T CmdHandler](command string, minParams int, docs string, router map[[2]string]T) *root.CmdConfig {
	// Build handlers of each router pair option
	handlerOf := make(map[[2]string]root.CmdHandler)
	for pair, handler := range router {
		handlerOf[pair] = handler.CmdHandler()
	}
	routerHandler := func(args []string) {
		option1 := strings.ToLower(args[0])
		option2 := strings.ToLower(args[1])
		key := [2]string{option1, option2}
		handler, ok := handlerOf[key]
		if !ok {
			sys.DisplayError(ErrInvalidOption)
			return
		}
		handler(args)
	}
	return &root.CmdConfig{
		Command:   command,
		MinParams: minParams,
		Docs:      docs,
		Handler:   routerHandler,
	}
}

// Create gin.HandlerFunc from AddRouter,
// [option1, option2] = Option1 is from :Fork, Option2 is from ?add query option
func AddFork[T WebHandler](router map[[2]string]T, response *web.ResponseType) gin.HandlerFunc {
	// Build handlers of each router pair option
	handlerOf := make(map[[2]string]gin.HandlerFunc)
	for pair, handler := range router {
		handlerOf[pair] = handler.WebHandler()
	}
	return func(c *gin.Context) {
		option1 := sys.WebForkParam(c)
		option2 := sys.WebAddOption(c)
		key := [2]string{option1, option2}
		handler, ok := handlerOf[key]
		if !ok {
			response.SendErrorFn(c, nil, ErrInvalidOption)
			return
		}
		handler(c)
	}
}

// Fork the web router using the given option fn
func OptionFork[T WebHandler](router map[string]T, response *web.ResponseType, optionFn func(*gin.Context) string) gin.HandlerFunc {
	// Build handlers of each route option
	handlerOf := make(map[string]gin.HandlerFunc)
	for key, handler := range router {
		handlerOf[key] = handler.WebHandler()
	}
	return func(c *gin.Context) {
		option := optionFn(c)
		handler, ok := handlerOf[option]
		if !ok {
			handler, ok = handlerOf[sys.DEFAULT_OPTION]
		}
		if !ok {
			response.SendErrorFn(c, nil, ErrInvalidOption)
			return
		}
		handler(c)
	}
}
