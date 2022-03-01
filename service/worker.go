package service

import (
	"context"
	"errors"
	"time"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/task"
	"github.com/zhupanovdm/gophermart/providers/accruals"
	"github.com/zhupanovdm/gophermart/storage"
)

type AccrualsWorker struct {
	name     string
	accruals accruals.Accruals
	pending  *PendingOrders
	orders   storage.Orders
}

func (w *AccrualsWorker) job(ctx context.Context) {
	ctx, logger := logging.GetOrCreateLogger(ctx, logging.WithService(w.name))
	logger.Info().Msg("worker is up")

	for {
		ord := w.pending.Get()
		if ord == nil {
			logger.Info().Msg("job is completed")
			return
		}
		ctx, logger := logging.ServiceLogger(ctx, w.name, logging.With(ord))
		logger.Trace().Msg("got order to process")

		w.process(ctx, ord)
	}
}

func (w *AccrualsWorker) process(ctx context.Context, ord *order.Order) {
	ctx, logger := logging.GetOrCreateLogger(ctx, logging.WithService(w.name), logging.With(ord))
	logger.Info().Msg("processing order")

	done := ctx.Done()
	for {
		resp, err := w.accruals.Get(ctx, string(ord.Number))
		if err != nil {
			code, e := accruals.GetError(err)
			if code != accruals.ErrTooManyRequest {
				logger.Err(err).Msg("loyalty service execution failure")
				return
			}
			logger.Warn().Msgf("got response: too many requests: will sleep for %v", e.RetryAfter)

			select {
			case <-done:
				logger.Err(errors.New("interrupted")).Msg("order processing interrupted")
				return
			case <-time.After(e.RetryAfter):
				logger.Info().Msg("resuming processing")
				continue
			}
		}

		if resp == nil {
			logger.Warn().Msg("got void result")
			break
		}

		logger.Info().Msg("updating order")
		if err = w.orders.Update(ctx, ord.ID, resp.Status.ToCanonical(), (*model.Sum)(resp.Accrual)); err != nil {
			logger.Err(err).Msg("failed to update order")
		}

		logger.Trace().Msg("updated order")
		break
	}
}

func NewWorker(name string, ordersStorage storage.Orders, client accruals.Accruals, pending *PendingOrders) task.Task {
	w := &AccrualsWorker{
		name:     name,
		pending:  pending,
		orders:   ordersStorage,
		accruals: client,
	}
	//goland:noinspection GoRedundantConversion
	return task.Task(w.job)
}
