package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/service"
)

const (
	authorizationHandlerName = "Authorization Handler"

	AuthorizationHeader = "Authorization"
	TokenPrefix         = "Bearer "

	CtxKeyUserID = app.ContextKey("UserID")
)

type authorizationHandler struct {
	service.Auth
	permitted *server.URLMatcher
}

func (h *authorizationHandler) AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx, logger := logging.ServiceLogger(req.Context(), authorizationHandlerName)
		logger.Info().Msg("authorizing client's request")

		if h.permitted.Match(req) {
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

		userID, err := h.Auth.Authorize(ctx, user.Token(token[len(TokenPrefix):]))
		if err != nil {
			logger.Err(err).Msg("client authentication failed")
			server.Error(resp, http.StatusUnauthorized, "invalid credentials")
			return
		}

		logger.UpdateContext(logging.ContextWith(userID))
		ctx = logging.SetLogger(ctx, logger)

		logger.Trace().Msg("client request authorized")

		next.ServeHTTP(resp, req.WithContext(context.WithValue(ctx, CtxKeyUserID, userID)))
	})
}

func NewAuthorizeMiddleware(auth service.Auth, permitted *server.URLMatcher) func(http.Handler) http.Handler {
	return (&authorizationHandler{
		Auth:      auth,
		permitted: permitted,
	}).AuthorizeMiddleware
}

func AuthorizedUserID(ctx context.Context) user.ID {
	if userID, ok := ctx.Value(CtxKeyUserID).(user.ID); ok {
		return userID
	}
	return user.VoidID
}
