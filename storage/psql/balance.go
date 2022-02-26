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
}

func (b *balanceStorage) Get(ctx context.Context, userId user.ID) (balance.Balance, error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName)
	logger.UpdateContext(logging.ContextWith(userId))
	logger.Info().Msg("querying client balance")

	sql := `
SELECT
    COALESCE(a.sum, 0) - COALESCE(w.sum, 0), COALESCE(w.sum, 0)
FROM
    (SELECT sum(accrual) sum FROM orders WHERE user_id = $1) a,
    (SELECT sum(sum) sum FROM withdrawals w, orders o WHERE w.order_id = o.id AND o.user_id = $1) w`

	var sum balance.Balance
	if err := b.QueryRow(ctx, sql, userId).Scan(&sum.Current, &sum.Withdrawn); err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn().Msg("balance record not found")
			return balance.Balance{}, nil
		}
		logger.Err(err).Msg("failed to query client balance")
		return balance.Balance{}, err
	}
	return sum, nil
}

func (b *balanceStorage) Withdraw(ctx context.Context, withdraw balance.Withdraw) (ok bool, err error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName)
	logger.UpdateContext(logging.ContextWith(withdraw.Number))
	logger.Info().Msg("processing withdraw")

	err = b.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		sum, err := orderBalance(ctx, tx, withdraw.Number)
		if err != nil {
			logger.Err(err).Msg("failed to query order balance")
			return err
		}
		if sum < withdraw.Sum {
			return nil
		}
		if _, err = tx.Exec(ctx, "INSERT INTO withdrawals(order_id, sum, processed_at) VALUES((SELECT id FROM orders WHERE number = $1),$2,$3)",
			withdraw.Number, withdraw.Sum, time.Now().In(time.Local)); err != nil {
			logger.Err(err).Msg("failed to withdraw")
			return err
		}
		ok = true
		return nil
	})
	return
}

func (b *balanceStorage) Withdrawals(ctx context.Context, userId user.ID) (balance.Withdrawals, error) {
	ctx, logger := logging.ServiceLogger(ctx, balanceStorageName)
	logger.UpdateContext(logging.ContextWith(userId))
	logger.Info().Msg("querying client withdrawals")

	sql := `
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

	rows, err := b.Query(ctx, sql, userId)
	if err != nil {
		logger.Err(err).Msg("failed to query withdrawals")
		return nil, err
	}
	defer rows.Close()

	list := make(balance.Withdrawals, 0)
	for rows.Next() {
		w := balance.Withdrawal{}
		list = append(list, &w)
		if err := rows.Scan(&w.Sum, &w.Number, w.ProcessedAt); err != nil {
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

func orderBalance(ctx context.Context, db queryExecutor, number order.Number) (model.Money, error) {
	sql := `
SELECT
    COALESCE(accrual, 0) - COALESCE((SELECT sum(sum) FROM withdrawals WHERE order_id = o.id), 0)
FROM
    orders o
WHERE
    number = $1`

	var sum model.Money
	if err := db.QueryRow(ctx, sql, number).Scan(&sum); err != nil {
		if err == pgx.ErrNoRows {
			return model.Money(0), nil
		}
		return model.Money(0), err
	}
	return sum, nil
}

func Balance(conn *Connection) storage.Balance {
	return &balanceStorage{conn}
}
