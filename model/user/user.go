package user

import (
	"errors"
	"github.com/zhupanovdm/gophermart/pkg/hash"
)

const EmptyToken Token = ""

type (
	Token string

	Credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	User struct {
		Credentials
		ID int64
	}
)

func (t Token) String() string {
	return string(t)
}

func (c *Credentials) Validate() error {
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
