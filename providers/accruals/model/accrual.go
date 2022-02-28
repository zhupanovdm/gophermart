package model

import "github.com/zhupanovdm/gophermart/model/order"

type (
	AccrualResponse struct {
		Order   string   `json:"order"`
		Status  Status   `json:"status"`
		Accrual *float64 `json:"accrual"`
	}

	Status string
)

const (
	StatusRegistered Status = "REGISTERED"
	StatusInvalid    Status = "INVALID"
	StatusProcessing Status = "PROCESSING"
	StatusProcessed  Status = "PROCESSED"
)

func (s Status) ToCanonical() order.Status {
	switch s {
	case StatusRegistered:
		return order.StatusNew
	case StatusProcessing:
		return order.StatusProcessing
	case StatusInvalid:
		return order.StatusInvalid
	case StatusProcessed:
		return order.StatusProcessed
	}
	return ""
}
