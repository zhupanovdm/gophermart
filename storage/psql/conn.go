package psql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

const (
	DefaultTimeout = 15 * time.Second

	pgxDriverName = "PGX Driver"
)

type Connection struct {
	*pgxpool.Pool
}

func NewConnection(ctx context.Context, cfg *config.Config) (*Connection, error) {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(pgxDriverName))

	pool, err := pgxpool.Connect(ctx, cfg.DatabaseURI)
	if err != nil {
		logger.Err(err).Msg("failed to establish connection pool")
		return nil, err
	}

	logger.Info().Msg("connection pool established")
	return &Connection{pool}, nil
}

type queryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type rowScanner interface {
	Scan(...interface{}) error
}
