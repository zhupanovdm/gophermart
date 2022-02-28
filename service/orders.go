package service

import (
	"context"

	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const ordersServiceName = "Orders Service"

var _ Orders = (*ordersImpl)(nil)

type ordersImpl struct {
	storage.Orders
}

func (o *ordersImpl) Register(ctx context.Context, number order.Number, userId user.ID) error {
	ctx, logger := logging.ServiceLogger(ctx, ordersServiceName)
	logger.UpdateContext(logging.ContextWith(userId, number))
	logger.Info().Msg("serving order registration")

	ord, ok, err := o.Store(ctx, order.New(number, userId))
	if err != nil {
		logger.Err(err).Msg("order creation failed")
		return errors.Err(err)
	}

	if !ok {
		if ord.UserID == userId {
			logger.Warn().Msg("order exists")
			return errors.New(ErrOrderAlreadyRegistered, "already registered")
		} else {
			logger.Warn().Msg("order registered by another user")
			return errors.New(ErrOrderNumberCollision, "already registered by another user")
		}
	}

	logger.Trace().Msg("order registered")
	return nil
}

func (o *ordersImpl) GetAll(ctx context.Context, userId user.ID) (order.Orders, error) {
	ctx, logger := logging.ServiceLogger(ctx, ordersServiceName)
	logger.UpdateContext(logging.ContextWith(userId))
	logger.Info().Msg("querying orders")

	return o.Orders.OrdersByUser(ctx, userId)
}

func NewOrders(orders storage.Orders) Orders {
	return &ordersImpl{orders}
}
