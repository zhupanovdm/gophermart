package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zhupanovdm/gophermart/model/user"
)

const (
	authorizationHeader = "Authorization"
	tokenPrefix         = "Bearer "
)

type TokenBearerHeader http.Header

func (p TokenBearerHeader) Set(token user.Token) {
	http.Header(p).Set(authorizationHeader, fmt.Sprint(tokenPrefix, token))
}

func (p TokenBearerHeader) Get() user.Token {
	token := http.Header(p).Get(authorizationHeader)
	if strings.HasPrefix(token, tokenPrefix) {
		return user.Token(token[len(tokenPrefix):])
	}
	return user.VoidToken
}
