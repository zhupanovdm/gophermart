package order

import (
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

var _ logging.ContextUpdater = (*Number)(nil)

type (
	ID     int64
	Number string
	Status string

	Order struct {
		ID         ID           `json:"-"`
		Number     Number       `json:"number"`
		UserID     user.ID      `json:"-"`
		Status     Status       `json:"status"`
		Accrual    *model.Money `json:"accrual,omitempty"`
		UploadedAt time.Time    `json:"uploaded_at"`
	}

	Orders []*Order
)

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func (s Status) position() int {
	switch s {
	case StatusNew:
		return 0
	case StatusProcessing:
		return 1
	case StatusInvalid:
		return 2
	case StatusProcessed:
		return 2
	}
	return -1
}

func (s Status) Compare(right Status) int {
	return right.position() - s.position()
}

func (n Number) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	return ctx.Str(logging.OrderNumberKey, string(n))
}

func (n Number) Validate(validators ...func(string) bool) error {
	for _, validator := range validators {
		if !validator(string(n)) {
			return errors.New("invalid number format")
		}
	}
	return nil
}

func New(number Number, userID user.ID) *Order {
	return &Order{
		Number:     number,
		UserID:     userID,
		Status:     StatusNew,
		UploadedAt: time.Now().Local(),
	}
}
