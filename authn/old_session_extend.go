package authn

import (
	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/krap/conk"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

// Extend the sessions stored in sessionUpdates
func extendSessions() (*ze.Request, error) {
	rq, err := ze.NewRequest("ExtendSessions")
	if err != nil {
		return rq, err
	}

	start := clock.Now()
	defer func() {
		rq.AddDurationLog(start)
	}()

	extensions := sessionUpdates.Map()
	numExtend := len(extensions)
	rq.AddFmtLog("Extend: %d", numExtend)
	if numExtend == 0 {
		return rq, nil // end early if no extensions
	}

	if Sessions == nil {
		return rq, ze.ErrMissingSchema
	}

	// Concurrently extend sessions
	extend := func(rq *ze.Request, entry dict.Entry[string, [2]ze.DateTime]) error {
		authToken, timePair := entry.Key, entry.Value
		return extendSession(rq, authToken, timePair)
	}
	entries := dict.Entries(extensions)
	result := conk.RequestWorkers(rq, entries, extend, 4)
	success, fail := result.Success, len(result.Errors)
	for i, entry := range entries {
		if dict.NoKey(result.Errors, i) {
			continue
		}
		rq.AddFmtLog("Failed to extend %s: %s", entry.Key, result.Errors[i].Error())
	}

	// Linear implementation
	// i, success, fail := 0, 0, 0
	// for authToken, timePair := range extensions {
	// 	i++
	// 	rq.AddFmtLog("%.2d / %.2d: ExtendSession: %s - %s", i, numExtend, authToken, timePair[1])

	// 	err = extendSession(rq, authToken, timePair)
	// 	if err != nil {
	// 		fail += 1
	// 		rq.AddFmtLog("Failed: %s", err.Error())
	// 	} else {
	// 		success += 1
	// 	}
	// }

	sessionUpdates.Clear()
	rq.AddFmtLog("Success: %d, Fail: %d", success, fail)
	return rq, nil
}

// Extend one session's expiry
func extendSession(rq *ze.Request, authTokenString string, timePair [2]ze.DateTime) error {
	authToken := NewToken(authTokenString)
	if authToken == nil {
		return ErrInvalidSession
	}
	now, expiry := timePair[0], timePair[1]

	// Sessions null check is done above in extendSessions
	s := Sessions.Ref
	q := rdb.NewUpdateQuery[Session](Sessions.Table)
	q.Where(rdb.And(
		rdb.Equal(&s.Type, authToken.Type),
		rdb.Equal(&s.Code, authToken.Code),
	))
	rdb.Update(q, &s.LastUpdatedAt, now)
	rdb.Update(q, &s.ExpiresAt, expiry)
	rdb.Update(q, &s.Status, sessionExtended)

	_, err := rdb.Exec(q, rq.DB)
	return err
}
