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
	ctx, cancel := context.WithCancel(context.Background())

	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(appName))
	logger.Info().Msg("starting app")

	cfg, err := config.Load(cli)
	if err != nil {
		logger.Err(err).Msg("failed to load agent config")
		return
	}

	if err := migrate.Prepare(ctx, cfg); err != nil {
		logger.Err(err).Msg("failed to prepare app db")
		return
	}

	db, err := psql.NewConnection(ctx, cfg)
	if err != nil {
		logger.Err(err).Msg("failed to connect to app db")
		return
	}
	defer db.Close()

	ordersStorage := psql.Orders(db)

	var wg1 sync.WaitGroup

	pending := service.NewPendingOrders()

	if err = service.NewAccruals(cfg, ordersStorage, accrualsClient.Factory(cfg), pending).Start(ctx, &wg1); err != nil {
		logger.Err(err).Msg("failed to start accruals service")
		return
	}

	jwt := service.NewJWT(cfg)
	auth := service.NewAuth(cfg, psql.Users(db), jwt)
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

	srvGroup := server.Start(ctx, cfg.RunAddress, handler, serverName)

	<-app.TerminationSignal()
	logger.Info().Msg("got shutdown signal")

	cancel()
	srvGroup.Wait()
	pending.Stop()

	wg1.Wait()
}
