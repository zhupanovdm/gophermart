package auth

import (
	"github.com/zhupanovdm/gophermart/model/user"
	"hash"
)

type JWTProvider struct {
	secret string
	hash.Hash
}

func (j *JWTProvider) Create(*user.User) (user.Token, error) {
	return user.EmptyToken, nil

}

func (j *JWTProvider) User(token user.Token) (*user.User, error) {
	return &user.User{}, nil
}

func JWT(secret string) *JWTProvider {
	return &JWTProvider{
		secret: secret,
	}
}
