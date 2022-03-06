package server

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/zhupanovdm/gophermart/pkg/logging"
)

func Start(ctx context.Context, addr string, handler http.Handler, name string, wg *sync.WaitGroup) *http.Server {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(name))

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
		ConnContext: func(ctx context.Context, _ net.Conn) context.Context {
			ctx, _ = logging.GetOrCreateLogger(ctx, logging.WithService(name))
			return ctx
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info().Msgf("running server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			logger.Err(err).Msg("server stopped")
		}
	}()
	return srv
}

func Handler(pattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	router := chi.NewRouter()
	for _, mw := range middlewares {
		router.Use(mw)
	}
	router.Mount(pattern, handler)
	return router
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	_, logger := logging.GetOrCreateLogger(r.Context())
	logger.Err(err).Msg(msg)
	Error(w, http.StatusInternalServerError, nil)
}
