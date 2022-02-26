package service

import (
	"context"
	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

var _ Balance = (*balanceImpl)(nil)

const (
	balanceServiceName = "Balance Service"

	ErrInsufficientFunds errors.ErrorCode = iota
	ErrOrderNotFound
	ErrOrderWrongOwner
)

type balanceImpl struct {
	storage.Balance
	storage.Orders
}

func (b *balanceImpl) Get(ctx context.Context, userID user.ID) (balance.Balance, error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceServiceName)
	logger.UpdateContext(logging.ContextWith(userID))
	logger.Info().Msg("querying balance")

	return b.Balance.Get(ctx, userID)
}

func (b *balanceImpl) Withdraw(ctx context.Context, userID user.ID, withdraw balance.Withdraw) error {
	ctx, logger := logging.ServiceLogger(ctx, balanceServiceName)
	logger.UpdateContext(logging.ContextWith(withdraw.Number))
	logger.Info().Msg("serving withdraw")

	ord, err := b.OrderByNumber(ctx, withdraw.Number)
	if err != nil {
		logger.Err(err).Msg("failed to query order")
		return errors.Err(err)
	}
	if ord == nil {
		logger.Warn().Msg("order not found")
		return errors.New(ErrOrderNotFound, "order not found")
	}
	if ord.UserID != userID {
		logger.Warn().Msg("order owner mismatch")
		return errors.New(ErrOrderWrongOwner, "wrong owner")
	}
	ok, err := b.Balance.Withdraw(ctx, withdraw)
	if err != nil {
		logger.Err(err).Msg("failed to withdraw")
		return errors.Err(err)
	}
	if !ok {
		logger.Warn().Msg("insufficient funds")
		return errors.New(ErrInsufficientFunds, "insufficient funds")
	}
	return nil
}

func (b *balanceImpl) Withdrawals(ctx context.Context, userID user.ID) (balance.Withdrawals, error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceServiceName)
	logger.Info().Msg("querying withdrawals")

	return b.Balance.Withdrawals(ctx, userID)
}

func NewBalance(balanceStorage storage.Balance, ordersStorage storage.Orders) Balance {
	return &balanceImpl{
		Balance: balanceStorage,
		Orders:  ordersStorage,
	}
}
