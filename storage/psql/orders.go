package psql

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const ordersStorageName = "Orders PSQL Storage"

var _ storage.Orders = (*ordersStorage)(nil)

type ordersStorage struct {
	*Connection
}

func (o *ordersStorage) Add(ctx context.Context, newOrder *order.Order) (ord *order.Order, ok bool, err error) {
	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName)
	logger.UpdateContext(logging.ContextWith(newOrder.Number))
	logger.Info().Msg("storing new order")

	err = o.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		if ord, err = orderByNumber(ctx, tx, newOrder.Number); err != nil {
			logger.Err(err).Msg("failed to query order")
			return err
		}
		if ord != nil {
			logger.Warn().Msg("order already exists")
			return nil
		}
		if _, err = tx.Exec(ctx, "INSERT INTO orders(number,user_id,status,uploaded_at) VALUES($1,$2,$3,$4)",
			newOrder.Number, newOrder.UserID, newOrder.Status, newOrder.UploadedAt); err != nil {
			logger.Err(err).Msg("failed to persist new order")
			return err
		}
		ord = newOrder
		ok = true
		return nil
	})
	return
}

func (o *ordersStorage) OrderByNumber(ctx context.Context, number order.Number) (*order.Order, error) {
	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName)
	logger.UpdateContext(logging.ContextWith(number))
	logger.Info().Msg("query order by number")

	return orderByNumber(ctx, o, number)
}

func (o *ordersStorage) GetAll(ctx context.Context, userId user.ID) (order.Orders, error) {
	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName)
	logger.Info().Msg("querying client orders")

	sql := `
SELECT
	id,
	number,
	status,
	uploaded_at,
	accrual
FROM
	orders
WHERE
	user_id = $1
ORDER BY
	uploaded_at`

	rows, err := o.Query(ctx, sql, userId)
	if err != nil {
		logger.Err(err).Msg("failed to query client orders")
		return nil, err
	}
	defer rows.Close()

	list := make(order.Orders, 0)
	for rows.Next() {
		var ord order.Order
		list = append(list, &ord)
		if err := rows.Scan(&ord.ID, &ord.Number, &ord.Status, &ord.UploadedAt, &ord.Accrual); err != nil {
			logger.Err(err).Msg("failed to query client orders")
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		logger.Err(err).Msg("failed to query client orders")
		return nil, err
	}

	logger.Trace().Msgf("got %d records", len(list))
	return list, nil
}

func orderByNumber(ctx context.Context, db queryExecutor, number order.Number) (*order.Order, error) {
	var ord order.Order
	err := db.QueryRow(ctx, "SELECT id, number, user_id, status, uploaded_at FROM orders WHERE number = $1", number).
		Scan(&ord.ID, &ord.Number, &ord.UserID, &ord.Status, &ord.UploadedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ord, nil
}

func Orders(conn *Connection) storage.Orders {
	return &ordersStorage{conn}
}
