package audit

import (
	"fmt"
	"strings"

	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/list"
	"github.com/zeroibot/fn/str"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

const detailsGlue string = "|"

// Initialize audit package
func Initialize() error {
	errs := make([]error, 0)

	ActionLogs = ze.AddSharedSchema(&ActionLog{}, errs)
	BatchLogs = ze.AddSchema(&BatchLog{}, "logs_batch", errs)
	BatchLogItems = ze.AddSchema(&BatchLogItem{}, "logs_batch_items", errs)

	if len(errs) > 0 {
		return fmt.Errorf("%d errors encountered: %w", len(errs), errs[0])
	}

	return nil
}

// Creates new details string from items separated by |
func NewDetails(items ...string) string {
	return strings.Join(items, detailsGlue)
}

// Creates new action details list
func NewUpdateActionDetails(action, itemCode string, updates rdb.FieldUpdates) [][2]string {
	return list.Map(dict.SortedEntries(updates), func(entry dict.Entry[string, rdb.FieldUpdate]) [2]string {
		oldValue, newValue := entry.Value.Tuple()
		details := NewDetails(itemCode, entry.Key, str.Any(oldValue), str.Any(newValue))
		return [2]string{action, details}
	})
}

// Creates new ActionLog
func newActionLog(actorID ze.ID, action, details string) *ActionLog {
	actionLog := &ActionLog{}
	actionLog.CreatedAt = clock.DateTimeNow()
	actionLog.ActorID = actorID
	actionLog.Action = action
	actionLog.Details = details
	return actionLog
}

// Creates list of new ActionLogs
func newActionLogs(actorID ze.ID, actionDetails [][2]string) []*ActionLog {
	actionLogs := make([]*ActionLog, len(actionDetails))
	for i, pair := range actionDetails {
		action, details := pair[0], pair[1]
		actionLogs[i] = newActionLog(actorID, action, details)
	}
	return actionLogs
}

// Inserts ActionLog using transaction at given table
func AddActionLogTx(rqtx *ze.Request, actorID ze.ID, action, details, table string) error {
	if ActionLogs == nil {
		rqtx.Status = ze.Err500
		return ze.ErrMissingSchema
	}
	actionLog := newActionLog(actorID, action, details)
	return ActionLogs.InsertTxAt(rqtx, actionLog, table)
}

// Inserts ActionLog rows using transaction at given table
func AddActionLogsTx(rqtx *ze.Request, actorID ze.ID, actionDetails [][2]string, table string) error {
	if ActionLogs == nil {
		rqtx.Status = ze.Err500
		return ze.ErrMissingSchema
	}
	actionLogs := newActionLogs(actorID, actionDetails)
	return ActionLogs.InsertTxRowsAt(rqtx, actionLogs, table)
}

// Creates new BatchLog
func newBatchLog(action, details, actionGlue string) *BatchLog {
	now := clock.Now()
	batchLog := &BatchLog{}
	batchLog.CreatedAt = clock.StandardFormat(now)
	batchLog.Code = fmt.Sprintf("%s-%s", clock.TimestampFormat(now), str.PartInitials(action, actionGlue))
	batchLog.Action = action
	batchLog.Details = details
	return batchLog
}

// Creates rows of new BatchLog items
func newBatchLogItems(batchCode string, detailsList []string) []*BatchLogItem {
	batchItems := make([]*BatchLogItem, len(detailsList))
	for i, details := range detailsList {
		batchItem := &BatchLogItem{}
		batchItem.Code = batchCode
		batchItem.Details = details
		batchItems[i] = batchItem
	}
	return batchItems
}

// Inserts BatchLog using transaction
func AddBatchLogTx(rqtx *ze.Request, action, details, actionGlue string) (string, error) {
	if BatchLogs == nil {
		rqtx.Status = ze.Err500
		return "", ze.ErrMissingSchema
	}
	batchLog := newBatchLog(action, details, actionGlue)
	err := BatchLogs.InsertTx(rqtx, batchLog)
	return batchLog.Code, err
}

// Inserts BatchLogItems using transaction
func AddBatchLogItemsTx(rqtx *ze.Request, batchCode string, detailsList []string) error {
	if BatchLogItems == nil {
		rqtx.Status = ze.Err500
		return ze.ErrMissingSchema
	}
	batchItems := newBatchLogItems(batchCode, detailsList)
	return BatchLogItems.InsertTxRows(rqtx, batchItems)
}
