package sys

import "github.com/zeroibot/fn/ds"

const (
	DEFAULT_OPTION string = "."
	ANY_TYPE       string = "*"
	LIST_CURRENT   string = "*"
	LIST_ARCHIVE   string = "archive"
	LIST_FUTURE    string = "future"
	toggleOn       string = "on"
	toggleOff      string = "off"
	viewAll        string = "all"
	viewActive     string = "active"
)

// Response for creating multiple items
type BulkCreateResult[T any] struct {
	BulkActionResult
	Items *ds.List[*T]
}

// Response for performing action on multiple items
type BulkActionResult struct {
	Success int
	Fail    int
	Fails   []string
}
