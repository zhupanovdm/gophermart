package main

import (
	"context"
	"flag"
	"sync"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/handlers"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	accrualsClient "github.com/zhupanovdm/gophermart/providers/accruals/http"
	"github.com/zhupanovdm/gophermart/service"
	"github.com/zhupanovdm/gophermart/storage/psql"
	"github.com/zhupanovdm/gophermart/storage/psql/migrate"
)

const (
	appName    = "GopherMart"
	serverName = "HTTP Server"
)

func cli(cfg *config.Config, flag *flag.FlagSet) {
	flag.StringVar(&cfg.RunAddress, "a", config.DefaultRunAddress, "Server address")
	flag.StringVar(&cfg.DatabaseURI, "d", config.DefaultDatabaseURI, "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", config.DefaultAccrualSystemAddress, "Accrual System Address")
}

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(appName))
	logger.Info().Msg("starting app")

	cfg, err := config.Load(cli)
	if err != nil {
		logger.Err(err).Msg("failed to load app config")
		return
	}

	if err := migrate.Prepare(ctx, cfg); err != nil {
		logger.Err(err).Msg("failed to prepare app db")
		return
	}

	db, err := psql.NewConnection(ctx, cfg)
	if err != nil {
		logger.Err(err).Msg("failed to connect to app's db")
		return
	}
	defer db.Close()

	ordersStorage := psql.Orders(db)

	pending := service.NewPendingOrders()
	defer pending.Stop()

	if err = service.NewAccruals(cfg, ordersStorage, accrualsClient.Factory(cfg), pending).Start(ctx, &wg); err != nil {
		logger.Err(err).Msg("failed to start accruals service")
		return
	}

	auth := service.NewAuth(cfg, psql.Users(db), service.NewJWT(cfg))
	orders := service.NewOrders(ordersStorage)
	balance := service.NewBalance(psql.Balance(db), ordersStorage)

	permitted := server.NewURLMatcher()
	if err := permitted.SetPattern("/api/user/login", "/api/user/register"); err != nil {
		logger.Err(err).Msg("failed to set permitted urls")
		return
	}

	handler := server.Handler("/api/user",
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

	srv := server.Start(ctx, cfg.RunAddress, handler, serverName, &wg)
	defer func() {
		logger.Info().Msg("closing server")
		if err := srv.Close(); err != nil {
			logger.Err(err).Msg("server close failed")
		}
	}()

	<-app.TerminationSignal()
	logger.Info().Msg("got shutdown signal")
}
