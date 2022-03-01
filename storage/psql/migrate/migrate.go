package migrate

import (
	"context"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

const migrationToolName = "PG Migration Tool"

func Prepare(ctx context.Context, cfg *config.Config) error {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(migrationToolName))
	logger.Info().Msg("preparing db")

	opt, err := pg.ParseURL(cfg.DatabaseURI)
	if err != nil {
		return err
	}
	db := pg.Connect(opt)

	_, _, err = migrations.Run(db, "init")
	if err != nil {
		logger.Err(err).Msg("init migration tool failed")
		return err
	}

	oldVersion, newVersion, err := migrations.Run(db)
	if err != nil {
		logger.Err(err).Msg("migration failed")
		return err
	}
	if newVersion != oldVersion {
		logger.Info().Msgf("schema migration: v.%d to v.%d", oldVersion, newVersion)
	} else {
		logger.Info().Msgf("current schema v.%d", oldVersion)
	}
	return nil
}
