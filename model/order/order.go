package order

import (
	"github.com/zhupanovdm/gophermart/model"
	"time"
)

type (
	Number int64
	Status string

	Order struct {
		Number     Number
		Status     Status
		UploadedAt time.Time
		Accrual    model.Money
	}

	Orders []*Order
)

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)
