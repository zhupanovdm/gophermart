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

	ord, err := b.orders.OrderByNumber(ctx, withdraw.Order)
	if err != nil {
		logger.Err(err).Msg("failed to query order by number from storage")
		return err
	}
	if ord == nil {
		logger.Warn().Msg("order by number not found")
		return errors.New(ErrOrderNotFound, "order not found")
	}
	if ord.UserID != userID {
		logger.Warn().Msg("order owner mismatch")
		return errors.New(ErrOrderWrongOwner, "wrong owner")
	}

	logger = logging.ApplyOptions(logger, logging.With(ord))
	ctx = logging.SetLogger(ctx, logger)

	ok, err := b.balance.Withdraw(ctx, ord.ID, withdraw.Sum)
	if err != nil {
		logger.Err(err).Msg("failed to post withdraw transaction")
		return err
	}
	if !ok {
		logger.Warn().Msg("insufficient balance")
		return errors.New(ErrInsufficientFunds, "insufficient balance")
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
