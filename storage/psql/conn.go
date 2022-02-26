package psql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/zhupanovdm/gophermart/pkg/logging"
)

const pgxDriverName = "PGX Driver"

type Connection struct {
	*pgxpool.Pool
}

func NewConnection(ctx context.Context, url string) (*Connection, error) {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(pgxDriverName))

	pool, err := pgxpool.Connect(ctx, url)
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
