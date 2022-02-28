package main

import (
	"context"
	"crypto"
	"github.com/zhupanovdm/gophermart/providers/accruals"
	"sync"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/zhupanovdm/gophermart/handlers"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"github.com/zhupanovdm/gophermart/pkg/hash"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/providers/accruals/http"
	"github.com/zhupanovdm/gophermart/service"
	"github.com/zhupanovdm/gophermart/storage/psql"
	"github.com/zhupanovdm/gophermart/storage/psql/migrate"
)

const (
	appName    = "GopherMart"
	serverName = "HTTP Server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(appName))
	logger.Info().Msg("starting app")

	dbUrl := "postgresql://postgres:qwe54321@localhost:5432/gophermart?sslmode=disable"
	if err := migrate.Prepare(ctx, dbUrl); err != nil {
		logger.Err(err).Msg("failed to prepare app db")
		return
	}

	db, err := psql.NewConnection(ctx, dbUrl)
	if err != nil {
		logger.Err(err).Msg("failed to connect to app db")
		return
	}
	defer db.Close()

	ordersStorage := psql.Orders(db)

	clientFactory := func() accruals.Accruals {
		return http.New("http://localhost:8080")
	}

	var wg1 sync.WaitGroup

	service.NewAccruals(ordersStorage, clientFactory, &wg1).Start(ctx)

	jwt := service.NewJWT([]byte("np891yx2"))
	auth := service.NewAuth(psql.Users(db), jwt, hash.StringWith(crypto.SHA512))
	orders := service.NewOrders(ordersStorage)
	balance := service.NewBalance(psql.Balance(db), ordersStorage)

	permitted := server.NewRequestMatcher()
	if err := permitted.URLPattern("/api/user/login", "/api/user/register"); err != nil {
		logger.Err(err).Msg("failed to set permitted urls")
		return
	}

	rootHandler := server.Handler("/api/user",
		handlers.NewUserHandler(
			handlers.NewAuthenticationHandler(auth),
			handlers.NewOrders(orders),
			handlers.NewBalance(balance)),
		middleware.RealIP,
		server.CorrelationID,
		server.Logger,
		server.CompressGzip,
		server.DecompressGzip,
		handlers.NewAuthorizeMiddleware(auth, permitted),
		middleware.Recoverer)

	srvGroup := server.Start(ctx, ":8081", rootHandler, serverName)

	<-app.TerminationSignal()
	logger.Info().Msg("got shutdown signal")

	cancel()
	srvGroup.Wait()

	wg1.Wait()
}
