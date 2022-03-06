package handlers

import (
	"context"
	"errors"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/service"
)

const (
	SampleExists     = "exists"
	SampleWrong      = "wrong"
	SampleCrash      = "crash"
	SampleFakeString = "fake"
	SampleFakeID     = 777
)

var _ service.Auth = (*authServiceMock)(nil)

type authServiceMock struct{}

func (a *authServiceMock) Register(_ context.Context, cred user.Credentials) error {
	if cred.Login == SampleExists {
		return service.ErrUserAlreadyRegistered
	}
	if cred.Login == SampleCrash {
		return errors.New(SampleFakeString)
	}
	return nil
}

func (a *authServiceMock) Login(_ context.Context, cred user.Credentials) (user.Token, error) {
	if cred.Login == SampleWrong {
		return user.VoidToken, service.ErrBadCredentials
	}
	if cred.Login == SampleCrash {
		return user.VoidToken, errors.New(SampleFakeString)
	}
	return SampleFakeString, nil
}

func (a *authServiceMock) Authorize(_ context.Context, token user.Token) (user.ID, error) {
	if token == SampleWrong {
		return user.VoidID, service.ErrBadCredentials
	}
	return user.ID(SampleFakeID), nil
}

func NewAuthServiceMock() service.Auth {
	return &authServiceMock{}
}
