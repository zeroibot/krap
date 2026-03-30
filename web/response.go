package web

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/fail"
	"github.com/zeroibot/fn/lang"
	"github.com/zeroibot/rdb/ze"
)

const okMessage = "OK"

// Map common error => status codes
var webErrorStatus = map[error]int{
	fail.MissingParams:  ze.Err400,
	ze.ErrMissingSchema: ze.Err500,
}

// Response for action web requests
type actionResponse struct {
	Success bool
	Message string
}

// Response for data web requests
type dataResponse struct {
	Data    any
	Message string
}

type ResponseType struct {
	SendErrorFn func(*gin.Context, *ze.Request, error)
}

var (
	Action = &ResponseType{
		SendErrorFn: SendActionError,
	}
	Data = &ResponseType{
		SendErrorFn: SendDataError,
	}
)

// Sends actionResponse
func SendActionResponse(c *gin.Context, rq *ze.Request, err error) {
	// TODO: APILog output
	fmt.Println("Output:\n" + getOutput(rq, err)) // temporary

	if err == nil {
		// url := c.Request.URL.Path
		// TODO: APILog url, status
		c.JSON(rq.Status, actionResponse{
			Success: true,
			Message: okMessage,
		})
	} else {
		sendActionError(c, rq, err)
	}
}

// Sends dataResponse
func SendDataResponse[T any](c *gin.Context, data *T, rq *ze.Request, err error) {
	// TODO: APILog output
	fmt.Println("Output:\n" + getOutput(rq, err)) // temporary

	if err == nil {
		// url := c.Request.URL.Path
		// TODO: APILog url, status, include data in logs?
		c.JSON(rq.Status, dataResponse{
			Data:    data,
			Message: okMessage,
		})

	} else {
		sendDataError(c, rq, err)
	}
}

// Sends actionResponse with given error
func SendActionError(c *gin.Context, rq *ze.Request, err error) {
	// TODO: APILog output
	fmt.Println("Output:\n" + getOutput(rq, err)) // temporary
	sendActionError(c, rq, err)
}

// Sends dataResponse with given error
func SendDataError(c *gin.Context, rq *ze.Request, err error) {
	// TODO: APILog output
	fmt.Println("Output:\n" + getOutput(rq, err)) // temporary
	sendDataError(c, rq, err)
}

// Common: sends an actionResponse with the given error
func sendActionError(c *gin.Context, rq *ze.Request, err error) {
	// url := c.Request.URL.Path
	status, message := getStatusMessage(rq, err)
	c.JSON(status, actionResponse{
		Success: false,
		Message: message,
	})
}

// Common: sends a dataResponse with the given error
func sendDataError(c *gin.Context, rq *ze.Request, err error) {
	// url := c.Request.URL.Path
	status, message := getStatusMessage(rq, err)
	c.JSON(status, dataResponse{
		Data:    nil,
		Message: message,
	})
}

// Combines the request output and the error message into one string, joined by newline
func getOutput(rq *ze.Request, err error) string {
	output := make([]string, 0)
	if rq != nil {
		output = append(output, rq.Output())
	}
	if err != nil {
		output = append(output, fmt.Sprintf("Error: %s", err.Error()))
	}
	return strings.Join(output, "\n")
}

// Get error message and status code
func getStatusMessage(rq *ze.Request, err error) (int, string) {
	message, ok := fail.PublicMessage(err)
	// default status: 400 if has public error message, else 500
	status := lang.Ternary(ok, ze.Err400, ze.Err500)
	// Check request status
	if rq != nil {
		status = rq.Status // get request status by default
		// If has error but status remained at OK200, find from web errors
		if err != nil && status == ze.OK200 {
			found := false
			for errKey := range webErrorStatus {
				if errors.Is(err, errKey) {
					status = webErrorStatus[errKey]
					found = true
					break
				}
			}
			if !found {
				status = ze.Err500 // defaults to 500 if not found in web errors
			}
		}
	}
	return status, message
}
