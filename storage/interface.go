package storage

import (
	"context"
	"github.com/zhupanovdm/gophermart/model"
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
		Store(context.Context, *order.Order) (*order.Order, bool, error)
		OrderByNumber(ctx context.Context, number order.Number) (*order.Order, error)
		OrdersByUser(context.Context, user.ID) (order.Orders, error)
		OrdersByStatus(context.Context, ...order.Status) (order.Orders, error)
		Update(context.Context, order.ID, order.Status, *model.Money) error
	}

	Balance interface {
		Get(context.Context, user.ID) (balance.Balance, error)
		Withdraw(context.Context, order.ID, model.Money) (bool, error)
		Withdrawals(context.Context, user.ID) (balance.Withdrawals, error)
	}
)
