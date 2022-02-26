package balance

import (
	"time"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/order"
)

type (
	Withdraw struct {
		Number order.Number `json:"order"`
		Sum    model.Money  `json:"sum"`
	}

	Withdrawal struct {
		Withdraw
		ProcessedAt time.Time `json:"processed_at"`
	}

	Withdrawals []*Withdrawal
)
