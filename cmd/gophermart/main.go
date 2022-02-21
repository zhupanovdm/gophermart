package main

import (
	"context"
	"crypto"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zhupanovdm/gophermart/handlers"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/service/auth"
	"github.com/zhupanovdm/gophermart/storage/psql"
)

const appName = "GopherMart"
const serverName = "HTTP Server"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithComponent(appName))

	gopherMartStorage := psql.New()

	jwtProvider := auth.JWT("np891yx2")
	authService := auth.New(gopherMartStorage, hash.StringWith(crypto.SHA512), jwtProvider)

	userHandler := handlers.User(handlers.NewAuth(authService))
	rootHandler := server.Handler("/api/user",
		userHandler,
		middleware.RealIP,
		server.CorrelationID,
		server.Logger,
		server.CompressGzip,
		server.DecompressGzip,
		middleware.Recoverer)

	srvGroup := server.Start(ctx, ":8080", rootHandler, serverName)

	<-app.TerminationSignal()
	logger.Info().Msg("got shutdown signal")
	cancel()
	srvGroup.Wait()
}
