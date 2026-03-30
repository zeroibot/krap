package audit

import (
	"github.com/zeroibot/rdb/ze"
)

var (
	ActionLogs    *ze.Schema[ActionLog]
	BatchLogs     *ze.Schema[BatchLog]
	BatchLogItems *ze.Schema[BatchLogItem]
)

type ActionLog struct {
	ze.CreatedItem
	ActorID    ze.ID  `json:"-"`
	ActorCode_ string `col:"-" json:"ActorCode"`
	ActionDetails
}

type BatchLog struct {
	ze.CodedItem
	ze.CreatedItem
	ActionDetails
}

type BatchLogItem struct {
	ze.CodedItem
	Details string
}

type ActionDetails struct {
	Action  string
	Details string
}
