package storage

import "github.com/zhupanovdm/gophermart/model/user"

type GopherMart interface {
	// UserByLogin получает пользователя по заданному логину
	UserByLogin(string) (*user.User, error)

	// AddUser создает пользователя, если пользователь с таким логином существует, то пользователь не будет создан.
	// ok - создан пользователь или нет
	AddUser(user.Credentials) (ok bool, err error)
}
