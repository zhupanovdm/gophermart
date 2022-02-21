package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/service/auth"
)

const HandlerName = "Auth Handler"

type Auth struct {
	auth *auth.Service
}

func (h *Auth) Register(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx, _ := logging.SetIfAbsentCID(req.Context(), logging.NewCID())
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithComponent(HandlerName), logging.WithCID(ctx))
	logger.Info().Msg("handling user registration request")

	var cred user.Credentials
	if err := json.NewDecoder(req.Body).Decode(&cred); err != nil {
		logger.Err(err).Msg("failed to decode request body")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}

	if err := cred.Validate(); err != nil {
		logger.Err(err).Msg("request body validation failed")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}

	if err := h.auth.Register(ctx, cred); err != nil {
		if errors.ErrCode(err) == auth.ErrAlreadyExists {
			logger.Err(err).Msg("specified login already in use")
			server.Error(resp, http.StatusConflict, "specified login already in use")
			return
		}

		logger.Err(err).Msg("failed to register user")
		server.Error(resp, http.StatusInternalServerError, nil)
	}
}

func (h *Auth) Login(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx, _ := logging.SetIfAbsentCID(req.Context(), logging.NewCID())
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithComponent(HandlerName), logging.WithCID(ctx))
	logger.Info().Msg("handling user login request")

	var cred user.Credentials
	if err := json.NewDecoder(req.Body).Decode(&cred); err != nil {
		logger.Err(err).Msg("failed to decode request body")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}

	if err := cred.Validate(); err != nil {
		logger.Err(err).Msg("request body validation failed")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}

	token, err := h.auth.Login(ctx, cred)
	if err != nil {
		if errors.ErrCode(err) == auth.ErrInvalidCredentials {
			logger.Err(err).Msg("invalid credentials")
			server.Error(resp, http.StatusUnauthorized, "invalid credentials")
			return
		}

		logger.Err(err).Msg("failed to register user")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	resp.Header().Set("Authorization", fmt.Sprintf("Bearer %v", token))
}

func NewAuth(authService *auth.Service) *Auth {
	return &Auth{
		auth: authService,
	}
}
