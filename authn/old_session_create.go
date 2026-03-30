package authn

import (
	"strings"

	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/fn/fail"
	"github.com/zeroibot/fn/hash"
	"github.com/zeroibot/fn/str"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

// Function hook for post-authentication actions to account
type PostAuthHook[T any] = func(*ze.Request, *T)

// Authenticate account
func authenticateAccount[T Checkable](rq *ze.Request, params *Params, schema *ze.Schema[T], condition rdb.Condition) (*T, error) {
	if condition == nil {
		rq.AddLog("Missing condition for authenticate account")
		return nil, fail.MissingParams
	}

	account, err := schema.Get(rq, condition)
	if err != nil {
		rq.AddErrorLog(err)
		rq.Status = ze.Err401
		return nil, ErrNotFoundAccount
	}

	hashPassword := (*account).GetPassword()
	if ok := hash.MatchPassword(params.Password, hashPassword); !ok {
		rq.AddLog("Password doesnt match")
		rq.Status = ze.Err401
		return nil, ErrFailedAuth
	}

	return account, nil
}

// Creates new session
func newSession[T identifiable](rq *ze.Request, accountRef *T, origin *web.RequestOrigin) (*Session, error) {
	if Sessions == nil {
		return nil, ze.ErrMissingSchema
	}

	if accountRef == nil {
		return nil, fail.MissingParams
	}
	account := *accountRef
	accountID := account.GetID()

	// Prepare session object
	var browserInfo *string = nil
	var ipAddress *string = nil
	if origin != nil {
		browserInfo = str.NonEmptyRefString(origin.BrowserInfo)
		ipAddress = str.NonEmptyRefString(origin.IPAddress)
	}
	now, expiry := clock.DateTimeNowWithExpiry(sessionDuration)

	s := &Session{}
	s.ID = 0
	s.CreatedAt = now
	s.Type = strings.ToUpper(account.GetType())
	s.Code = str.RandomString(sessionCodeLength, true, true, true)
	s.AccountID = accountID
	s.LastUpdatedAt = now
	s.ExpiresAt = expiry
	s.Status = sessionActive
	s.BrowserInfo = browserInfo
	s.IPAddress = ipAddress

	// Insert session
	id, err := Sessions.InsertID(rq, s)
	if err != nil {
		return nil, err
	}
	s.ID = id
	storeAddSession(s) // add to session store

	return s, nil
}
