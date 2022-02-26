package psql

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const usersStorageName = "Users PSQL Storage"

var _ storage.Users = (*usersStorage)(nil)

type usersStorage struct {
	*Connection
}

func (u *usersStorage) UserByLogin(ctx context.Context, login string) (*user.User, error) {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(usersStorageName))
	logger.Trace().Msg("query user by login")

	usr, err := userByLogin(ctx, u, login)
	if err != nil {
		logger.Err(err).Msg("failed to query user")
		return nil, err
	}
	return usr, nil
}

func (u *usersStorage) UserByID(ctx context.Context, id user.ID) (*user.User, error) {
	ctx, logger := logging.GetOrCreateLogger(ctx, logging.WithService(usersStorageName))
	logger.UpdateContext(logging.ContextWith(id))
	logger.Trace().Msg("query user")

	return userByID(ctx, u, id)
}

func (u *usersStorage) CreateUser(ctx context.Context, cred user.Credentials) (ok bool, err error) {
	_, logger := logging.GetOrCreateLogger(ctx, logging.WithService(usersStorageName))
	logger.UpdateContext(logging.ContextWith(cred))
	logger.Info().Msg("persist new user")

	err = u.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		var usr *user.User
		if usr, err = userByLogin(ctx, tx, cred.Login); err != nil {
			logger.Err(err).Msg("failed to query user")
			return err
		}
		if usr != nil {
			logger.Trace().Msg("user with the same login already exists")
			return nil
		}
		_, err = tx.Exec(ctx, "INSERT INTO users(login, password) VALUES($1, $2)", cred.Login, cred.Password)
		if err != nil {
			logger.Err(err).Msg("failed to persist user")
			return err
		}
		ok = true
		logger.Trace().Msg("new user persisted")
		return nil
	})
	return
}

func userByLogin(ctx context.Context, db queryExecutor, login string) (*user.User, error) {
	var u user.User
	err := db.QueryRow(ctx, "SELECT id, login, password FROM users WHERE login=$1", login).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func userByID(ctx context.Context, db queryExecutor, id user.ID) (*user.User, error) {
	var u user.User
	err := db.QueryRow(ctx, "SELECT id, login, password FROM users WHERE id=$1", id).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func Users(conn *Connection) storage.Users {
	return &usersStorage{conn}
}
