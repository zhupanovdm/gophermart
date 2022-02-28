package service

import (
	"context"
	"github.com/zhupanovdm/gophermart/model"
	"runtime"
	"sync"
	"time"

	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/task"
	"github.com/zhupanovdm/gophermart/providers/accruals"
	"github.com/zhupanovdm/gophermart/storage"
)

const DefaultInterval = 5 * time.Second

const accrualsServiceName = "Accruals Service"

var _ Accruals = (*accrualsImpl)(nil)

type accrualsImpl struct {
	storage.Orders
	clientFactory func() accruals.Accruals
	pending       *PendingOrders
	interval      time.Duration
	*sync.WaitGroup
}

func (a *accrualsImpl) Start(ctx context.Context) {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(accrualsServiceName))
	logger.Info().Msg("starting accruals processing")

	for i := 0; i < runtime.NumCPU(); i++ {
		NewWorker(a.Orders, a.clientFactory(), a.pending).Start(ctx)
	}

	background := task.Task(a.ProceedProcessing).
		With(task.PeriodicRun(a.interval)).
		With(task.CompletionWait(a.WaitGroup))

	go background(ctx)
}

func (a *accrualsImpl) ProceedProcessing(ctx context.Context) {
	ctx, logger := logging.ServiceLogger(ctx, accrualsServiceName)
	logger.Info().Msg("serving order process")

	orders, err := a.OrdersByStatus(ctx, order.StatusNew, order.StatusProcessing)
	if err != nil {
		logger.Err(err).Msg("failed to get orders to process")
		return
	}

	a.pending.AddAll(orders)
}

func NewAccruals(ordersStorage storage.Orders, clientFactory func() accruals.Accruals, wg *sync.WaitGroup) Accruals {
	return &accrualsImpl{
		Orders:        ordersStorage,
		pending:       NewPendingOrders(),
		clientFactory: clientFactory,
		WaitGroup:     wg,
		interval:      DefaultInterval,
	}
}

type PendingOrders struct {
	sync.Mutex
	sequence []order.ID
	set      map[order.ID]*order.Order
	*sync.Cond
}

func NewPendingOrders() *PendingOrders {
	pendingOrders := PendingOrders{
		sequence: make([]order.ID, 0),
		set:      make(map[order.ID]*order.Order),
	}
	pendingOrders.Cond = sync.NewCond(&pendingOrders)
	return &pendingOrders
}

func (p *PendingOrders) Add(orders ...*order.Order) {
	p.Lock()
	defer p.Unlock()

	for _, ord := range orders {
		if _, ok := p.set[ord.ID]; ok {
			continue
		}
		p.set[ord.ID] = ord
		p.sequence = append(p.sequence, ord.ID)
	}

	p.Signal()
}

func (p *PendingOrders) AddAll(orders order.Orders) {
	p.Add(orders...)
}

func (p *PendingOrders) Get() *order.Order {
	p.Lock()
	defer p.Unlock()
	for len(p.sequence) == 0 {
		p.Wait()
	}
	id := p.sequence[0]
	p.sequence = p.sequence[1:]

	ord := p.set[id]
	delete(p.set, id)

	return ord
}

type Worker struct {
	client  accruals.Accruals
	pending *PendingOrders
	storage.Orders
}

func (w *Worker) Start(ctx context.Context) {
	for {
		ord := w.pending.Get()
		resp, err := w.client.Get(ctx, string(ord.Number))
		if err != nil {
			if code, e := accruals.GetError(err); code == accruals.ErrTooManyRequest {
				<-time.After(e.RetryAfter)
				w.pending.Add(ord)
				continue
			}
		}

		err = w.Orders.Update(ctx, ord.ID, resp.Status.ToCanonical(), (*model.Money)(resp.Accrual))
		if err != nil {

		}

	}
}

func NewWorker(ordersStorage storage.Orders, client accruals.Accruals, pending *PendingOrders) *Worker {
	return &Worker{
		Orders:  ordersStorage,
		client:  client,
		pending: pending,
	}
}
