package user

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

var _ logging.ContextUpdater = (*Credentials)(nil)
var _ logging.ContextUpdater = (*User)(nil)
var _ logging.ContextUpdater = (*ID)(nil)

const (
	VoidID    = ID(0)
	VoidToken = Token("")

	IDLogKey    = "user_id"
	LoginLogKey = "user_login"
)

type (
	Credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	ID int64

	User struct {
		Credentials
		ID ID
	}

	Token string
)

func (u *User) String() string {
	if u == nil {
		return model.NilObjectString
	}
	return fmt.Sprintf("%v, id: %d", u.Credentials, u.ID)
}

func (u *User) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	if u == nil {
		return ctx
	}
	return logging.ContextWith(u.ID, u.Credentials).UpdateLogContext(ctx)
}

func (c Credentials) Validate() error {
	if c.Login == "" {
		return errors.New("user login not specified")
	}
	if c.Password == "" {
		return errors.New("user password not specified")
	}
	return nil
}

func (c *Credentials) HashPassword(h hash.StringFunc) {
	c.Password = h(c.Password)
}

func (c Credentials) String() string {
	return c.Login
}

func (c Credentials) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	return ctx.Str(LoginLogKey, c.Login)
}

func (t Token) String() string {
	return string(t)
}

func (id ID) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	return ctx.Stringer(IDLogKey, id)
}

func (id ID) String() string {
	return fmt.Sprintf("%d", id)
}
