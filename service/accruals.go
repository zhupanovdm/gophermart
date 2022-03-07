package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/task"
	"github.com/zhupanovdm/gophermart/providers/accruals"
	"github.com/zhupanovdm/gophermart/storage"
)

const accrualsServiceName = "Accruals Service"

var _ Accruals = (*accrualsImpl)(nil)

type accrualsImpl struct {
	orders               storage.Orders
	createAccrualsClient func() (accruals.Accruals, error)
	pending              *PendingOrders
	wg                   *sync.WaitGroup
	interval             time.Duration
	workerCount          int
}

func (a *accrualsImpl) Start(ctx context.Context, wg *sync.WaitGroup) error {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(accrualsServiceName))
	logger.Info().Msg("start serving orders")

	for i := 0; i < a.workerCount; i++ {
		worker, err := a.worker(fmt.Sprintf("%s (worker %d)", accrualsServiceName, i))
		if err != nil {
			logger.Err(err).Msg("failed to create worker")
			return err
		}
		worker.With(task.WaitGroup(wg)).Go(ctx)
	}

	task.Task(a.fetchOrders).
		With(task.PeriodicRun(a.interval)).
		With(task.WaitGroup(wg)).
		Go(ctx)

	return nil
}

func (a *accrualsImpl) worker(name string) (task.Task, error) {
	client, err := a.createAccrualsClient()
	if err != nil {
		return nil, err
	}
	return NewWorker(name, a.orders, client, a.pending), nil
}

func (a *accrualsImpl) fetchOrders(ctx context.Context) {
	ctx, logger := logging.ServiceLogger(ctx, accrualsServiceName)
	logger.Info().Msg("polling unprocessed orders")

	orders, err := a.orders.OrdersByStatus(ctx, order.StatusNew, order.StatusProcessing)
	if err != nil {
		logger.Err(err).Msg("failed to poll orders for processing")
		return
	}

	logger.Trace().Msgf("got %d orders", len(orders))
	a.pending.AddAll(orders)
}

func NewAccruals(cfg *config.Config, ordersStorage storage.Orders, clientFactory accruals.ClientFactory, pending *PendingOrders) Accruals {
	return &accrualsImpl{
		orders:               ordersStorage,
		createAccrualsClient: clientFactory,
		pending:              pending,
		interval:             cfg.AccrualsPollingInterval,
		workerCount:          cfg.AccrualsWorkersCount,
	}
}
