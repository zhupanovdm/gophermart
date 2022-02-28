package service

import "github.com/zhupanovdm/gophermart/pkg/errors"

const (
	ErrBadCredentials errors.ErrorCode = iota
	ErrUserAlreadyRegistered

	ErrOrderNotFound
	ErrOrderAlreadyRegistered
	ErrOrderNumberCollision
	ErrOrderWrongOwner

	ErrInsufficientFunds
)
