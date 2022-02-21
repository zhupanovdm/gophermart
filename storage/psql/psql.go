package psql

import (
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/storage"
)

var _ storage.GopherMart = (*client)(nil)

type client struct {
}

func (p *client) UserByLogin(s string) (*user.User, error) {
	return &user.User{
		ID: 777,
		Credentials: user.Credentials{
			Login:    s,
			Password: "********",
		},
	}, nil
}

func (p client) AddUser(credentials user.Credentials) (ok bool, err error) {
	return true, nil
}

func New() storage.GopherMart {
	return &client{}
}
