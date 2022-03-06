package service

import "errors"

var (
	ErrBadCredentials         = errors.New("bad credentials")
	ErrUserAlreadyRegistered  = errors.New("user is already registered")
	ErrOrderAlreadyRegistered = errors.New("order already registered")
	ErrOrderNumberCollision   = errors.New("order number collision")
	ErrInsufficientFunds      = errors.New("insufficient funds")
)
