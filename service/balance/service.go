package balance

import (
	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/order"
)

type Service struct {
}

func (s *Service) Balance() (*balance.Balance, error) {
	return nil, nil
}

func (s *Service) Withdraw(number order.Number, sum model.Money) error {
	return nil
}

func (s *Service) Withdrawals() (balance.Withdrawals, error) {
	return nil, nil
}
