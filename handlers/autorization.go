package handlers

import (
	"context"
	"github.com/zhupanovdm/gophermart/service"
	"net/http"
	"strings"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
)

const (
	authorizationHandlerName = "Authorization Handler"

	AuthorizationHeader = "Authorization"
	TokenPrefix         = "Bearer "

	CtxKeyUserID = app.ContextKey("UserID")
)

type authorizationHandler struct {
	service.Auth
	permitted *RequestMatcher
}

func (h *authorizationHandler) AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx, logger := logging.ServiceLogger(req.Context(), authorizationHandlerName)
		logger.Info().Msg("authorizing client's request")

		if h.permitted.MatchURL(req) {
			logger.Info().Msg("requested url is permitted")
			next.ServeHTTP(resp, req)
			return
		}

		token := req.Header.Get(AuthorizationHeader)
		if !strings.HasPrefix(token, TokenPrefix) {
			logger.Warn().Msg("request has no authorization token")
			server.Error(resp, http.StatusUnauthorized, "invalid credentials")
			return
		}

		userId, err := h.Auth.Authorize(ctx, user.Token(token[len(TokenPrefix):]))
		if err != nil {
			logger.Err(err).Msg("client authentication failed")
			server.Error(resp, http.StatusUnauthorized, "invalid credentials")
			return
		}

		logger.UpdateContext(logging.ContextWith(userId))
		ctx = logging.SetLogger(ctx, logger)

		logger.Trace().Msg("client request authorized")

		next.ServeHTTP(resp, req.WithContext(context.WithValue(ctx, CtxKeyUserID, userId)))
	})
}

func NewAuthorizeMiddleware(auth service.Auth, permitted *RequestMatcher) func(http.Handler) http.Handler {
	return (&authorizationHandler{
		Auth:      auth,
		permitted: permitted,
	}).AuthorizeMiddleware
}

func AuthorizedUserId(ctx context.Context) user.ID {
	if userId, ok := ctx.Value(CtxKeyUserID).(user.ID); ok {
		return userId
	}
	return user.VoidID
}
