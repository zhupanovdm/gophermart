package service

import (
	"context"
	"sync"

	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
)

type (
	Auth interface {
		Register(context.Context, user.Credentials) error
		Login(context.Context, user.Credentials) (user.Token, error)
		Authorize(context.Context, user.Token) (user.ID, error)
	}

	Balance interface {
		Get(context.Context, user.ID) (balance.Balance, error)
		Withdraw(context.Context, user.ID, balance.Withdraw) error
		Withdrawals(context.Context, user.ID) (balance.Withdrawals, error)
	}

	JWT interface {
		Token(context.Context, *user.User) (user.Token, error)
		Authenticate(context.Context, user.Token) (user.ID, error)
	}

	Orders interface {
		Register(context.Context, order.Number, user.ID) error
		GetAll(context.Context, user.ID) (order.Orders, error)
	}

	Accruals interface {
		Start(context.Context, *sync.WaitGroup) error
	}
)
