package errors

import "errors"

var _ error = (*Error)(nil)

const ErrUnknown ErrorCode = -1

type (
	ErrorCode int

	Error struct {
		error
		Code ErrorCode
	}
)

func (e ErrorCode) IsUnknown() bool {
	return e == ErrUnknown
}

func New(code ErrorCode, description string) *Error {
	return &Error{
		Code:  code,
		error: errors.New(description),
	}
}

func ErrCode(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ErrUnknown
}
