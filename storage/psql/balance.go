package psql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const balanceStorageName = "Balance PSQL Storage"

var _ storage.Balance = (*balanceStorage)(nil)

type balanceStorage struct {
	*Connection
	timeout time.Duration
}

func (b *balanceStorage) Get(ctx context.Context, userID user.ID) (balance.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName, logging.With(userID))
	logger.Info().Msg("querying client balance")

	query := `
SELECT
    COALESCE(a.sum, 0) - COALESCE(w.sum, 0), COALESCE(w.sum, 0)
FROM
    (SELECT sum(accrual) sum FROM orders WHERE user_id = $1) a,
    (SELECT sum(sum) sum FROM withdrawals w, orders o WHERE w.order_id = o.id AND o.user_id = $1) w`

	var sum balance.Balance
	if err := b.QueryRow(ctx, query, userID).Scan(&sum.Current, &sum.Withdrawn); err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn().Msg("client's balance record not found")
			return balance.Balance{}, nil
		}

		logger.Err(err).Msg("failed to query client balance")
		return balance.Balance{}, err
	}

	logger.Trace().Msg("success")
	return sum, nil
}

func (b *balanceStorage) Withdraw(ctx context.Context, orderID order.ID, requested model.Sum) (ok bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName, logging.With(orderID))
	logger.Info().Msg("processing withdraw")

	err = b.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		sum, err := orderBalance(ctx, tx, orderID)
		if err != nil {
			logger.Err(err).Msg("failed to query order balance")
			return err
		}
		if sum < requested {
			logger.Warn().Msg("requested sum exceeds available client balance")
			return nil
		}
		if _, err = tx.Exec(ctx, "INSERT INTO withdrawals(order_id, sum, processed_at) VALUES($1,$2,$3)",
			orderID, requested, time.Now().Local()); err != nil {
			logger.Err(err).Msg("failed to store withdraw transaction")
			return err
		}

		ok = true
		logger.Trace().Msg("success")
		return nil
	})
	return
}

func (b *balanceStorage) Withdrawals(ctx context.Context, userID user.ID) (balance.Withdrawals, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName, logging.With(userID))
	logger.Info().Msg("querying client withdrawals")

	query := `
SELECT
	sum,
	order_number,
	processed_at
FROM
	withdrawals,
	orders
WHERE
	order_number = number
	AND user_id = $1`

	rows, err := b.Query(ctx, query, userID)
	if err != nil {
		logger.Err(err).Msg("failed to query withdrawals")
		return nil, err
	}
	defer rows.Close()

	list := make(balance.Withdrawals, 0)
	for rows.Next() {
		var w balance.Withdrawal
		list = append(list, &w)
		if err := rows.Scan(&w.Sum, &w.Order, w.ProcessedAt); err != nil {
			logger.Err(err).Msg("failed to query withdrawals")
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		logger.Err(err).Msg("failed to query withdrawals")
		return nil, err
	}
	return list, nil
}

func orderBalance(ctx context.Context, db queryExecutor, orderID order.ID) (model.Sum, error) {
	query := `
SELECT
    COALESCE(accrual, 0) - COALESCE((SELECT sum(sum) FROM withdrawals WHERE order_id = $1), 0)
FROM
    orders o
WHERE
    id = $1`

	var sum model.Sum
	if err := db.QueryRow(ctx, query, orderID).Scan(&sum); err != nil {
		if err == pgx.ErrNoRows {
			return model.Sum(0), nil
		}
		return model.Sum(0), err
	}
	return sum, nil
}

func Balance(conn *Connection) storage.Balance {
	return &balanceStorage{
		Connection: conn,
		timeout:    DefaultTimeout,
	}
}
