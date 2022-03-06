package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/service"
)

const authenticationHandlerName = "Authentication Handler"

type authenticationHandler struct {
	service.Auth
}

func (h *authenticationHandler) Register(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), authenticationHandlerName)
	logger.Info().Msg("handling client registration request")

	if cred, ok := h.decodeAuthRequest(ctx, resp, req); ok {
		logger.UpdateContext(logging.ContextWith(cred))

		if err := h.Auth.Register(ctx, cred); err != nil {
			if errors.Is(err, service.ErrUserAlreadyRegistered) {
				logger.Err(err).Msg("specified login already in use")
				server.Error(resp, http.StatusConflict, err)
			} else {
				logger.Err(err).Msg("failed to register user")
				server.Error(resp, http.StatusInternalServerError, nil)
			}
			return
		}
		logger.Info().Msg("user registered")
		h.authenticate(ctx, cred, resp)
	}
}

func (h *authenticationHandler) Login(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), authenticationHandlerName)
	logger.Info().Msg("handling client authentication request")

	if cred, ok := h.decodeAuthRequest(ctx, resp, req); ok {
		h.authenticate(ctx, cred, resp)
	}
}

func (h *authenticationHandler) authenticate(ctx context.Context, cred user.Credentials, resp http.ResponseWriter) {
	_, logger := logging.GetOrCreateLogger(ctx)
	logger.UpdateContext(logging.ContextWith(cred))

	token, err := h.Auth.Login(ctx, cred)
	if err != nil {
		if errors.Is(err, service.ErrBadCredentials) {
			logger.Err(err).Msg("client authentication failed")
			server.Error(resp, http.StatusUnauthorized, err)
			return
		}
		logger.Err(err).Msg("failed to register user")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	resp.Header().Set(AuthorizationHeader, fmt.Sprint(TokenPrefix, " ", token))
	logger.Info().Msg("user authenticated")
}

func (h *authenticationHandler) decodeAuthRequest(ctx context.Context, resp http.ResponseWriter, req *http.Request) (cred user.Credentials, ok bool) {
	_, logger := logging.GetOrCreateLogger(ctx)

	if err := json.NewDecoder(req.Body).Decode(&cred); err != nil {
		logger.Err(err).Msg("failed to decode request")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}
	if err := cred.Validate(); err != nil {
		logger.Err(err).Msg("request validation failed")
		server.Error(resp, http.StatusBadRequest, err)
		return
	}
	ok = true
	return
}

func NewAuthenticationHandler(auth service.Auth) http.Handler {
	h := &authenticationHandler{auth}

	router := chi.NewRouter()
	router.Post("/register", h.Register)
	router.Post("/login", h.Login)
	return router
}
