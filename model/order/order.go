package order

import (
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

const NumberLogKey = "order_number"

var _ logging.ContextUpdater = (*Order)(nil)
var _ logging.ContextUpdater = (*Number)(nil)

type (
	Number string
	Status string

	Order struct {
		Number     Number     `json:"number"`
		UserID     user.ID    `json:"-"`
		Status     Status     `json:"status"`
		Accrual    *model.Sum `json:"accrual,omitempty"`
		UploadedAt time.Time  `json:"uploaded_at"`
	}

	Orders []*Order
)

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func (o *Order) UpdateLogContext(ctx zerolog.Context) zerolog.Context {
	return logging.ContextWith(o.Number).UpdateLogContext(ctx)
}

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
	return ctx.Stringer(NumberLogKey, n)
}

func (n Number) String() string {
	return string(n)
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

func StatusesToStrings(statuses []Status) []string {
	stringStatuses := make([]string, 0, len(statuses))
	for _, status := range statuses {
		stringStatuses = append(stringStatuses, string(status))
	}
	return stringStatuses
}
