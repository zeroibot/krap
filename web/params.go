package web

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/fail"
)

// Web request origin: BrowserInfo and IP address
type RequestOrigin struct {
	BrowserInfo *string
	IPAddress   *string
}

// Gets the web request origin
func GetRequestOrigin(c *gin.Context) *RequestOrigin {
	browserInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	return &RequestOrigin{
		BrowserInfo: &browserInfo,
		IPAddress:   &ipAddress,
	}
}

// Gets the request body object from web request
func GetRequestBody[T any](c *gin.Context, response *ResponseType) (*T, error) {
	var item *T
	err := c.BindJSON(&item)
	if err != nil {
		response.SendErrorFn(c, nil, fail.MissingParams)
		return nil, err
	}
	return item, nil
}
