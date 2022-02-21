package orders

import "github.com/zhupanovdm/gophermart/model/order"

type Service struct {
}

func (s *Service) Register(number order.Number) error {
	return nil
}

func (s Service) GetAll() (order.Orders, error) {
	return nil, nil
}

func New() *Service {
	return nil
}
