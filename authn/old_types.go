package authn

import (
	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/krap/web"
	"github.com/zeroibot/rdb/ze"
)

var Sessions *ze.Schema[Session]

const (
	sessionActive   string = "ACTIVE"
	sessionExtended string = "EXTENDED"
	sessionExpired  string = "EXPIRED"
	sessionLogout   string = "LOGOUT"
)

type identifiable interface {
	GetID() ze.ID
	GetType() string
}

type Checkable interface {
	GetPassword() string
}

type Authable interface {
	identifiable
	Checkable
}

type Session struct {
	ze.UniqueItem
	ze.CreatedItem
	Token
	AccountID     ze.ID  `json:"-"`
	AccountCode_  string `col:"-" json:"AccountCode"`
	LastUpdatedAt ze.DateTime
	ExpiresAt     ze.DateTime
	Status        string
	web.RequestOrigin
}

type Params struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}

type Token struct {
	Type string `validate:"required" fx:"upper"`
	Code string `validate:"required"`
}

func (a Token) String() string {
	return a.Type + authTokenGlue + a.Code
}

func (s Session) IsExpired() bool {
	return clock.IsExpired(s.ExpiresAt)
}
