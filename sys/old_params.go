package sys

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/io"
)

// Reads the patch object from path as type T, then convert to dict.Object
func CmdReadPatchObject[T any](path string) (dict.Object, error) {
	patchItem, err := io.ReadJSON[T](path)
	if err != nil {
		return nil, err
	}
	return dict.ToObject(patchItem)
}

// Reads the patch object from request body as type T, then convert to dict.Object
func WebReadPatchObject[T any](c *gin.Context) (dict.Object, error) {
	var patchItem T
	err := c.BindJSON(&patchItem)
	if err != nil {
		return nil, err
	}
	return dict.ToObject(&patchItem)
}

// False if option is 'all', otherwise true
func MustBeActiveOption(option string) bool {
	mustBeActive := strings.ToLower(option) != viewAll
	return mustBeActive
}

// Return toggle on/off (boolean), hasToggleOption (ok flag)
func ToggleOption(option string) (bool, bool) {
	switch strings.ToLower(option) {
	case toggleOn:
		return true, true
	case toggleOff:
		return false, true
	default:
		return false, false
	}
}

// Returns uppercase type at params[limit] if it exists,
// Defaults to ANY_TYPE (*)
func CmdTypeOption(params []string, limit int) string {
	typ := ANY_TYPE
	if len(params) > limit {
		typ = strings.ToUpper(params[limit])
	}
	return typ
}

// Return list type (current, archive, future)
func ListTypeOption(option string) string {
	option = strings.ToLower(option)
	listType := LIST_CURRENT
	switch option {
	case LIST_ARCHIVE:
		listType = option
	case LIST_FUTURE:
		listType = option
	}
	return listType
}

// Gets the lowercase :Fork param
func WebForkParam(c *gin.Context) string {
	return strings.ToLower(c.Param("Fork"))
}

// Gets the uppercase :Code param
func WebCodeParam(c *gin.Context) string {
	return strings.ToUpper(c.Param("Code"))
}

// Gets the lowercase :Type param
func WebTypeParam(c *gin.Context) string {
	return strings.ToLower(c.Param("Type"))
}

// False if option is 'all', otherwise true,
// From ?view=option query string
func WebMustBeActiveOption(c *gin.Context) bool {
	option := c.DefaultQuery("view", viewActive)
	return MustBeActiveOption(option)
}

// Return toggle on/off (boolean), hasToggleOption (ok flag),
// From ?toggle=option query string
func WebToggleOption(c *gin.Context) (bool, bool) {
	option := c.Query("toggle")
	return ToggleOption(option)
}

// Gets the uppercase code option from ?code=option query string
func WebCodeOption(c *gin.Context) string {
	return strings.ToUpper(c.Query("code"))
}

// Gets the uppercase type option from ?type=option query string,
// Defaults to ANY_TYPE (*)
func WebTypeOption(c *gin.Context) string {
	option := c.DefaultQuery("type", ANY_TYPE)
	return strings.ToUpper(option)
}

// Gets the lowercase add option from ?add=option query string,
// Defaults to DEFAULT_OPTION (.)
func WebAddOption(c *gin.Context) string {
	option := c.DefaultQuery("add", DEFAULT_OPTION)
	return strings.ToLower(option)
}

// Gets the lowercase by option from ?by=option query string,
// Defaults to DEFAULT_OPTION (.)
func WebByOption(c *gin.Context) string {
	option := c.DefaultQuery("by", DEFAULT_OPTION)
	return strings.ToLower(option)
}

// Gets the lowercase as option from ?as=option query string,
// Defaults to DEFAULT_OPTION (.)
func WebAsOption(c *gin.Context) string {
	option := c.DefaultQuery("as", DEFAULT_OPTION)
	return strings.ToLower(option)
}

// Gets lowercase list option from ?list=option query string,
// Defaults to LIST_CURRENT (*)
func WebListTypeOption(c *gin.Context) string {
	option := c.Query("list")
	return ListTypeOption(option)
}
