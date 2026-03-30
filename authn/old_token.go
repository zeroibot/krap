package authn

import (
	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/list"
	"github.com/zeroibot/fn/str"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

const (
	authTokenGlue string = "/"
	authHeaderKey string = "Authorization"
)

// Creates authn.Token from string "Type/Code"
func NewToken(authToken string) *Token {
	parts := str.CleanSplit(authToken, authTokenGlue)
	if len(parts) != 2 || list.Any(parts, str.IsEmpty) {
		return nil
	}
	return &Token{
		Type: parts[0],
		Code: parts[1],
	}
}

// Checks if authToken string can be a valid authn.Token
func IsToken(authToken string) bool {
	parts := str.CleanSplit(authToken, authTokenGlue)
	return len(parts) == 2 && list.All(parts, str.NotEmpty)
}

// Get the authn.Token from the Authorization header
func WebAuthToken(c *gin.Context) *Token {
	authHeader := c.GetHeader(authHeaderKey)
	return NewToken(authHeader)
}

// Get the authn.Token from the Authorizatio header;
// On error, send error response
func ReqAuthToken(c *gin.Context, response *web.ResponseType) *Token {
	authToken := WebAuthToken(c)
	if authToken == nil {
		rq := &ze.Request{Status: ze.Err401}
		response.SendErrorFn(c, rq, ErrInvalidSession)
		return nil
	}
	return authToken
}
