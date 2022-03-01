package service

import (
	"sync"

	"github.com/zhupanovdm/gophermart/model/order"
)

type PendingOrders struct {
	mu      sync.Mutex
	cond    *sync.Cond
	stopped bool
	orders  order.Orders
}

func (p *PendingOrders) AddAll(orders order.Orders) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for _, ord := range orders {
		p.orders = append(p.orders, ord)
	}
	p.cond.Signal()
}

func (p *PendingOrders) Get() *order.Order {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for !p.stopped && len(p.orders) == 0 {
		p.cond.Wait()
	}
	if p.stopped {
		return nil
	}

	ord := p.orders[0]
	p.orders = p.orders[1:]
	return ord
}

func (p *PendingOrders) Stop() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.stopped = true
	p.cond.Broadcast()
}

func NewPendingOrders() *PendingOrders {
	p := &PendingOrders{
		orders: make(order.Orders, 0),
	}
	p.cond = sync.NewCond(&p.mu)
	return p
}
