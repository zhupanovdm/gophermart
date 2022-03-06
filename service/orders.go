package service

import (
	"context"

	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const ordersServiceName = "Orders Service"

var _ Orders = (*ordersImpl)(nil)

type ordersImpl struct {
	orders storage.Orders
}

func (o *ordersImpl) Register(ctx context.Context, number order.Number, userID user.ID) error {
	ctx, logger := logging.ServiceLogger(ctx, ordersServiceName, logging.With(number, userID))
	logger.Info().Msg("serving order registration")

	ord, ok, err := o.orders.Create(ctx, order.New(number, userID))
	if err != nil {
		logger.Err(err).Msg("order creation failed")
		return err
	}
	if !ok {
		if ord.UserID == userID {
			logger.Warn().Msg("order exists")
			return ErrOrderAlreadyRegistered
		}
		logger.Warn().Msg("order registered by another user")
		return ErrOrderNumberCollision
	}

	logger.UpdateContext(logging.ContextWith(ord))
	logger.Trace().Msg("order registered")
	return nil
}

func (o *ordersImpl) GetAll(ctx context.Context, userID user.ID) (order.Orders, error) {
	ctx, logger := logging.ServiceLogger(ctx, ordersServiceName, logging.With(userID))
	logger.Info().Msg("querying client orders")

	return o.orders.OrdersByUser(ctx, userID)
}

func NewOrders(ordersStorage storage.Orders) Orders {
	return &ordersImpl{
		orders: ordersStorage,
	}
}
