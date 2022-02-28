package service

import (
	"context"
	"fmt"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const authServiceName = "Auth Service"

var _ Auth = (*authImpl)(nil)

type authImpl struct {
	storage.Users
	JWT
	passwdHash hash.StringFunc
}

func (a *authImpl) Register(ctx context.Context, cred user.Credentials) error {
	cred.HashPassword(a.passwdHash)

	ctx, logger := logging.ServiceLogger(ctx, authServiceName)
	logger.UpdateContext(logging.ContextWith(cred))
	logger.Info().Msg("register user")

	ok, err := a.CreateUser(ctx, cred)
	if err != nil {
		logger.Err(err).Msg("failed to persist user")
		return errors.Err(err)
	}
	if !ok {
		logger.Warn().Msg("user already exists")
		return errors.New(ErrUserAlreadyRegistered, fmt.Sprintf("user already exists: %v", cred))
	}
	return nil
}

func (a *authImpl) Login(ctx context.Context, cred user.Credentials) (user.Token, error) {
	cred.HashPassword(a.passwdHash)

	ctx, logger := logging.ServiceLogger(ctx, authServiceName)
	logger.UpdateContext(logging.ContextWith(cred))
	logger.Info().Msg("signing in")

	usr, err := a.UserByLogin(ctx, cred.Login)
	if err != nil {
		logger.Err(err).Msg("failed to query user")
		return user.VoidToken, errors.Err(err)
	}
	if usr == nil {
		logger.Warn().Msg("user not found")
		return user.VoidToken, errors.New(ErrBadCredentials, "invalid credentials")
	}

	logger.UpdateContext(logging.ContextWith(usr))

	if cred.Password != usr.Password {
		logger.Warn().Msg("invalid credentials")
		return user.VoidToken, errors.New(ErrBadCredentials, "invalid credentials")
	}

	token, err := a.Token(ctx, usr)
	if err != nil {
		logger.Err(err).Msg("failed to retrieve auth token")
		return user.VoidToken, errors.Err(err)
	}

	return token, nil
}

func (a *authImpl) Authorize(ctx context.Context, token user.Token) (user.ID, error) {
	ctx, logger := logging.ServiceLogger(ctx, authServiceName)
	logger.Info().Msg("authorizing client request")

	userId, err := a.Authenticate(ctx, token)
	if err != nil {
		logger.Err(err).Msg("invalid token")
		return user.VoidID, errors.New(ErrBadCredentials, "invalid token")
	}
	return userId, nil
}

func NewAuth(users storage.Users, jwt JWT, passwdHash hash.StringFunc) Auth {
	return &authImpl{
		Users:      users,
		JWT:        jwt,
		passwdHash: passwdHash,
	}
}
