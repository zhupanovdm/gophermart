package auth

import (
	"context"
	"fmt"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/storage"
)

const (
	ErrAlreadyExists errors.ErrorCode = iota
	ErrInvalidCredentials
)

type Service struct {
	passwdHashFunc hash.StringFunc
	users          storage.GopherMart
	jwt            *JWTProvider
}

func (s *Service) Register(ctx context.Context, cred user.Credentials) error {
	cred.HashPassword(s.passwdHashFunc)

	ok, err := s.users.AddUser(cred)
	if err != nil {
		return errors.Err(err)
	}
	if !ok {
		return errors.New(ErrAlreadyExists, fmt.Sprintf("user with such login already exists: %s", cred.Login))
	}
	return nil
}

func (s *Service) Login(ctx context.Context, cred user.Credentials) (user.Token, error) {
	u, err := s.users.UserByLogin(cred.Login)
	if err != nil {
		return user.EmptyToken, errors.Err(err)
	}

	cred.HashPassword(s.passwdHashFunc)
	if cred.Password != u.Password {
		return user.EmptyToken, errors.New(ErrInvalidCredentials, "invalid credentials")
	}

	token, err := s.jwt.Create(u)
	if err != nil {
		return user.EmptyToken, errors.Err(err)
	}
	return token, nil
}

func (s *Service) Authenticate(ctx context.Context, token user.Token) (*user.User, error) {
	u, err := s.jwt.User(token)
	if err != nil {
		return nil, errors.New(ErrInvalidCredentials, "invalid token")
	}
	return u, nil
}

func New(users storage.GopherMart, passwdHash hash.StringFunc, jwtProvider *JWTProvider) *Service {
	return &Service{
		passwdHashFunc: passwdHash,
		users:          users,
		jwt:            jwtProvider,
	}
}
