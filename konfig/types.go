package konfig

import (
	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/rdb/ze"
)

var (
	KVSchema *ze.Schema[KV]
)

type KV struct {
	Key           string `col:"AppKey"`
	Value         string `col:"AppValue"`
	LastUpdatedAt ze.DateTime
}

type Defaults struct {
	UintMap       map[string]uint
	IntMap        map[string]int
	StringMap     dict.StringMap
	StringListMap dict.StringListMap
}
