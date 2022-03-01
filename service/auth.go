package service

import (
	"context"
	"fmt"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const authServiceName = "Auth Service"

var _ Auth = (*authImpl)(nil)

type authImpl struct {
	users      storage.Users
	jwt        JWT
	passwdHash hash.StringFunc
}

func (a *authImpl) Register(ctx context.Context, cred user.Credentials) error {
	ctx, logger := logging.ServiceLogger(ctx, authServiceName, logging.With(cred))
	logger.Info().Msg("register new user")

	cred.HashPassword(a.passwdHash)
	if ok, err := a.users.CreateUser(ctx, cred); err != nil {
		logger.Err(err).Msg("failed to persist new user in storage")
		return err
	} else if !ok {
		logger.Warn().Msg("user already exists")
		return errors.New(ErrUserAlreadyRegistered, fmt.Sprintf("user already exists: %v", cred))
	}

	logger.Info().Msg("user registered")
	return nil
}

func (a *authImpl) Login(ctx context.Context, cred user.Credentials) (user.Token, error) {
	ctx, logger := logging.ServiceLogger(ctx, authServiceName, logging.With(cred))
	logger.Info().Msg("signing in")

	usr, err := a.users.UserByLogin(ctx, cred.Login)
	if err != nil {
		logger.Err(err).Msg("failed to query user from storage")
		return user.VoidToken, err
	}
	if usr == nil {
		logger.Warn().Msg("user not found")
		return user.VoidToken, errors.New(ErrBadCredentials, "invalid credentials")
	}
	logger = logging.ApplyOptions(logger, logging.With(usr))
	ctx = logging.SetLogger(ctx, logger)

	cred.HashPassword(a.passwdHash)
	if cred.Password != usr.Password {
		logger.Warn().Msg("invalid credentials")
		return user.VoidToken, errors.New(ErrBadCredentials, "invalid credentials")
	}

	token, err := a.jwt.Token(ctx, usr)
	if err != nil {
		logger.Err(err).Msg("failed to retrieve auth token")
		return user.VoidToken, err
	}

	logger.Info().Msg("authenticated")
	return token, nil
}

func (a *authImpl) Authorize(ctx context.Context, token user.Token) (user.ID, error) {
	ctx, logger := logging.ServiceLogger(ctx, authServiceName)
	logger.Info().Msg("authorizing client request")

	userId, err := a.jwt.Authenticate(ctx, token)
	if err != nil {
		logger.Err(err).Msg("invalid token")
		return user.VoidID, errors.New(ErrBadCredentials, "invalid token")
	}

	logger.Info().Msg("authorized")
	return userId, nil
}

func NewAuth(cfg *config.Config, users storage.Users, jwt JWT) Auth {
	return &authImpl{
		users:      users,
		jwt:        jwt,
		passwdHash: hash.StringWith(cfg.CryptoHash),
	}
}
