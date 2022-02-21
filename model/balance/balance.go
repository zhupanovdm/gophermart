package balance

import (
	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/order"
	"time"
)

type (
	Balance struct {
		Current   model.Money
		Withdrawn model.Money
	}

	Withdrawal struct {
		Number      order.Number
		Sum         model.Money
		ProcessedAt time.Time
	}

	Withdrawals []*Withdrawal
)
