package storage

import (
	"context"

	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
)

type (
	Users interface {
		// UserByLogin получает пользователя по заданному логину
		UserByLogin(context.Context, string) (*user.User, error)

		UserByID(context.Context, user.ID) (*user.User, error)

		// CreateUser добавляет пользователя. Если пользователь с таким логином уже существует, то пользователь не будет создан.
		// ok - создан пользователь или нет
		CreateUser(context.Context, user.Credentials) (ok bool, err error)
	}

	Orders interface {
		Add(context.Context, *order.Order) (*order.Order, bool, error)
		OrderByNumber(ctx context.Context, number order.Number) (*order.Order, error)
		GetAll(context.Context, user.ID) (order.Orders, error)
	}

	Balance interface {
		Get(context.Context, user.ID) (balance.Balance, error)
		Withdraw(context.Context, balance.Withdraw) (bool, error)
		Withdrawals(context.Context, user.ID) (balance.Withdrawals, error)
	}
)
