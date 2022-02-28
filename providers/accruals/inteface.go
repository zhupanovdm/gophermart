package accruals

import (
	"context"
	"fmt"
	"time"

	"github.com/zhupanovdm/gophermart/providers/accruals/model"
)

const (
	ErrUnknown = iota
	ErrTooManyRequest
)

type (
	Accruals interface {
		Get(context.Context, string) (*model.AccrualResponse, error)
	}

	Error struct {
		code       int
		message    string
		RetryAfter time.Duration
	}
)

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Code() int {
	return e.code
}

func NewError(code int, message interface{}) *Error {
	return &Error{
		code:    code,
		message: fmt.Sprint(message),
	}

}

func GetError(err error) (int, *Error) {
	if e, ok := err.(*Error); ok {
		return e.code, e
	}
	return ErrUnknown, nil
}
