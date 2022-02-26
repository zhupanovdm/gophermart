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
	Number string
	Status string

	Order struct {
		ID         int64        `json:"-"`
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

func (n Number) String() string {
	return string(n)
}

func (n Number) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	return ctx.Stringer(logging.OrderNumberKey, n)
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
