package service

import (
	"context"

	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

var _ Balance = (*balanceImpl)(nil)

const balanceServiceName = "Balance Service"

type balanceImpl struct {
	balance storage.Balance
	orders  storage.Orders
}

func (b *balanceImpl) Get(ctx context.Context, userID user.ID) (balance.Balance, error) {
	_, logger := logging.ServiceLogger(ctx, balanceServiceName, logging.With(userID))
	logger.Info().Msg("querying client balance")

	return b.balance.Get(ctx, userID)
}

func (b *balanceImpl) Withdraw(ctx context.Context, userID user.ID, withdraw balance.Withdraw) error {
	ctx, logger := logging.ServiceLogger(ctx, balanceServiceName, logging.With(userID, withdraw.Order))
	logger.Info().Msg("serving withdraw")

	ok, err := b.balance.Withdraw(ctx, userID, withdraw.Order, withdraw.Sum)
	if err != nil {
		logger.Err(err).Msg("failed to post withdraw transaction")
		return err
	}
	if !ok {
		logger.Warn().Msg("insufficient balance")
		return ErrInsufficientFunds
	}
	logger.Trace().Msg("withdraw successful")
	return nil
}

func (b *balanceImpl) Withdrawals(ctx context.Context, userID user.ID) (balance.Withdrawals, error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceServiceName, logging.With(userID))
	logger.Info().Msg("querying client withdrawals")

	return b.balance.Withdrawals(ctx, userID)
}

func NewBalance(balanceStorage storage.Balance, ordersStorage storage.Orders) Balance {
	return &balanceImpl{
		balance: balanceStorage,
		orders:  ordersStorage,
	}
}
