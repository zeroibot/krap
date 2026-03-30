package authn

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeroibot/fn/check"
	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/fn/fail"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

var (
	ErrExistingSession = errors.New("public: Existing session")
	ErrExpiredSession  = errors.New("public: Expired session")
	ErrFailedAuth      = errors.New("public: Authentication failed")
	ErrInvalidSession  = errors.New("public: Invalid session")
	ErrNotFoundAccount = errors.New("public: Account not found")
	ErrNotFoundSession = errors.New("public: Session not found")
)

var (
	sessionDuration   time.Duration = time.Duration(1) * time.Hour // default: 1 hour
	sessionCodeLength uint          = 20                           // default: 20 characters
)

const tableSessionsArchive string = "sessions_archive"

// Initialize authn package
func Initialize() error {
	errs := make([]error, 0)

	Sessions = ze.AddSchema(&Session{}, "sessions", errs)

	if len(errs) > 0 {
		return fmt.Errorf("%d errors encountered: %w", len(errs), errs[0])
	}

	return nil
}

// Sets the Session duration
func SetSessionDuration(duration time.Duration) {
	sessionDuration = duration
}

// Sets the SessionCode length
func SetSessionCodeLength(length uint) {
	sessionCodeLength = length
}

// Deletes the authToken's associated session (logout)
func DeleteSession(authToken *Token) (*ze.Request, error) {
	rq, err := ze.NewRequest("DeleteSession")
	if err != nil {
		return rq, err
	}

	if !check.IsValidStruct(authToken) {
		return rq, fail.MissingParams
	}

	session, err := findSession(rq, authToken)
	if err != nil {
		rq.AddLog("Failed to find session")
		return rq, err
	}

	session.LastUpdatedAt = clock.DateTimeNow()
	err = archiveSession(rq, session, sessionLogout)
	if err != nil {
		rq.AddLog("Failed to archive session")
		return rq, err
	}

	rq.AddFmtLog("Logout: %s", authToken.String())
	return rq, nil
}

// Find session and extend if not expired, otherwise archived
func TouchSession(rq *ze.Request, authToken *Token) (*Session, error) {
	session, err := findSession(rq, authToken)
	if err != nil {
		rq.AddLog("Failed to find session")
		return nil, err
	}

	if session.IsExpired() {
		err = archiveSession(rq, session, sessionExpired)
		if err != nil {
			rq.AddLog("Failed to archive session")
			return nil, err
		}
		rq.AddFmtLog("Expired: %s", authToken.String())
		rq.Status = ze.Err401
		return nil, ErrExpiredSession
	} else {
		now, expiry := clock.DateTimeNowWithExpiry(sessionDuration)
		storeExtendSession(session.Token.String(), now, expiry)
		return session, nil
	}
}

// Checks if authn.Token is still a valid session
func IsValidSession(authToken *Token) (bool, *ze.Request, error) {
	rq, err := ze.NewRequest("IsValidSession")
	if err != nil {
		return false, rq, err
	}

	if !check.IsValidStruct(authToken) {
		return false, rq, fail.MissingParams
	}

	session, err := TouchSession(rq, authToken)
	isValid := session != nil
	return isValid, rq, err
}

// Authenticate account
func AuthenticateAccount[T Checkable](authParams *Params, schema *ze.Schema[T], condition rdb.Condition) (*T, *ze.Request, error) {
	rq, err := ze.NewRequest("Authenticate%s", schema.Name)
	if err != nil {
		return nil, rq, err
	}

	if !check.IsValidStruct(authParams) {
		return nil, rq, fail.MissingParams
	}

	account, err := authenticateAccount(rq, authParams, schema, condition)
	if err != nil {
		return nil, rq, err
	}

	return account, rq, nil
}

// Create new session, returns account and sessionCode
func NewSession[T Authable](authParams *Params, origin *web.RequestOrigin, schema *ze.Schema[T], condition rdb.Condition, hook PostAuthHook[T]) (*T, string, *ze.Request, error) {
	var sessionCode string = ""

	rq, err := ze.NewRequest("New%sSession", schema.Name)
	if err != nil {
		return nil, sessionCode, rq, err
	}

	if !check.IsValidStruct(authParams) {
		return nil, sessionCode, rq, fail.MissingParams
	}

	account, err := authenticateAccount(rq, authParams, schema, condition)
	if err != nil {
		rq.AddFmtLog("%s authentication failed", schema.Name)
		return nil, sessionCode, rq, err
	}

	session, err := newSession(rq, account, origin)
	if err != nil {
		rq.AddFmtLog("Failed to create %s session", schema.Name)
		return nil, sessionCode, rq, err
	}

	if hook != nil {
		hook(rq, account)
	}

	return account, session.Code, rq, nil
}

// Common: gets the associated session for the authToken,
// Checks cache first, if cache-miss: queries the db
func findSession(rq *ze.Request, authToken *Token) (*Session, error) {
	var err error
	authToken.Type = strings.ToUpper(authToken.Type)
	session := storeGetSession(authToken.String())
	if session == nil {
		if Sessions == nil {
			return nil, ze.ErrMissingSchema
		}

		s := Sessions.Ref
		condition := rdb.And(
			rdb.Equal(&s.Code, authToken.Code),
			rdb.Equal(&s.Type, authToken.Type),
		)

		session, err = Sessions.Get(rq, condition)
		if err != nil {
			rq.AddErrorLog(err)
			rq.Status = ze.Err401
			return nil, ErrNotFoundSession
		}
		storeAddSession(session)
	}
	return session, nil
}
