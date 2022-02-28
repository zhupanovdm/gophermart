package handlers

import (
	"context"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/service"
)

const (
	SampleExists     = "exists"
	SampleWrong      = "wrong"
	SampleCrash      = "crash"
	SampleFakeString = "fake"
	SampleFakeId     = 777
)

var _ service.Auth = (*authServiceMock)(nil)

type authServiceMock struct{}

func (a *authServiceMock) Register(_ context.Context, cred user.Credentials) error {
	if cred.Login == SampleExists {
		return errors.New(service.ErrUserAlreadyRegistered, SampleFakeString)
	}
	if cred.Login == SampleCrash {
		return errors.New(errors.ErrUnknown, SampleFakeString)
	}
	return nil
}

func (a *authServiceMock) Login(_ context.Context, cred user.Credentials) (user.Token, error) {
	if cred.Login == SampleWrong {
		return user.VoidToken, errors.New(service.ErrBadCredentials, SampleFakeString)
	}
	if cred.Login == SampleCrash {
		return user.VoidToken, errors.New(errors.ErrUnknown, SampleFakeString)
	}
	return SampleFakeString, nil
}

func (a *authServiceMock) Authorize(_ context.Context, token user.Token) (user.ID, error) {
	if token == SampleWrong {
		return user.VoidID, errors.New(service.ErrBadCredentials, SampleFakeString)
	}
	return user.ID(SampleFakeId), nil
}

func NewAuthServiceMock() service.Auth {
	return &authServiceMock{}
}
