package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const ordersStorageName = "Orders PSQL Storage"

var _ storage.Orders = (*ordersStorage)(nil)

type ordersStorage struct {
	*Connection
	timeout time.Duration
}

func (o *ordersStorage) Create(ctx context.Context, newOrd *order.Order) (ord *order.Order, ok bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(newOrd))
	logger.Info().Msg("storing new order")

	err = o.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		if ord, err = orderByNumber(ctx, tx, newOrd.Number); err != nil {
			logger.Err(err).Msg("failed to check existing order with the same number")
			return err
		} else if ord != nil {
			logger.Warn().Msg("order already exists")
			return nil
		}

		if _, err = tx.Exec(ctx, "INSERT INTO orders(number,user_id,status,uploaded_at) VALUES($1,$2,$3,$4)",
			newOrd.Number, newOrd.UserID, newOrd.Status, newOrd.UploadedAt); err != nil {
			logger.Err(err).Msg("failed to persist new order")
			return err
		}
		if ord, err = orderByNumber(ctx, tx, newOrd.Number); err != nil {
			logger.Err(err).Msg("failed to query created order")
			return err
		}

		ok = true
		logger.Trace().Msg("success")
		return nil
	})
	return
}

func (o *ordersStorage) OrderByNumber(ctx context.Context, number order.Number) (*order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(number))
	logger.Info().Msg("query order by number")

	return orderByNumber(ctx, o, number)
}

func (o *ordersStorage) OrdersByUser(ctx context.Context, userID user.ID) (order.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(userID))
	logger.Info().Msg("querying client orders")

	query := `
SELECT
	id,
	number,
	status,
	accrual,
	user_id,
	uploaded_at
FROM
	orders
WHERE
	user_id = $1
ORDER BY
	uploaded_at`

	rows, err := o.Query(ctx, query, userID)
	if err != nil {
		logger.Err(err).Msg("failed to query client orders")
		return nil, err
	}
	return fetchOrders(ctx, rows)
}

func (o *ordersStorage) OrdersByStatus(ctx context.Context, statuses ...order.Status) (order.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName)
	logger.Info().Msg("querying orders by status")

	stringStatuses := make([]string, 0, len(statuses))
	for _, status := range statuses {
		stringStatuses = append(stringStatuses, string(status))
	}

	query := `
SELECT
	id,
	number,
	status,
	accrual,
	user_id,
	uploaded_at
FROM
	orders
WHERE
	status = ANY($1)
ORDER BY
	uploaded_at`

	rows, err := o.Query(ctx, query, stringStatuses)
	if err != nil {
		logger.Err(err).Msg("failed to query orders")
		return nil, err
	}
	return fetchOrders(ctx, rows)
}

func (o *ordersStorage) Update(ctx context.Context, orderID order.ID, status order.Status, accrual *model.Sum) error {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(orderID))
	logger.Info().Msg("querying client orders")

	return o.BeginFunc(ctx, func(tx pgx.Tx) error {
		ord, err := orderById(ctx, tx, orderID)
		if err != nil {
			logger.Err(err).Msg("failed to query order by id")
			return err
		}

		if change := ord.Status.Compare(status); change == 0 {
			logger.Warn().Msg("status is unchanged: skipping update")
			return nil
		} else if change < 0 {
			err = fmt.Errorf("can't descend order status: from %v to %v", ord.Status, status)
			logger.Err(err).Msg("failed to update order")
			return err
		}

		if _, err := o.Exec(ctx, "UPDATE orders SET status = $2, accrual = $3 WHERE id = $1", orderID, status, accrual); err != nil {
			logger.Err(err).Msg("failed to update order")
			return err
		}

		logger.Trace().Msg("success")
		return nil
	})
}

func orderByNumber(ctx context.Context, db queryExecutor, number order.Number) (*order.Order, error) {
	return fetchOrder(db.QueryRow(ctx, "SELECT id, number, user_id, status, uploaded_at FROM orders WHERE number = $1", number))
}

func orderById(ctx context.Context, db queryExecutor, id order.ID) (*order.Order, error) {
	return fetchOrder(db.QueryRow(ctx, "SELECT id, number, user_id, status, uploaded_at FROM orders WHERE id = $1", id))
}

func fetchOrder(s rowScanner) (*order.Order, error) {
	var ord order.Order
	var accrual sql.NullInt64

	err := s.Scan(&ord.ID, &ord.Number, &ord.Status, &accrual, &ord.UserID, &ord.UploadedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	sum := model.Sum(accrual.Int64)
	if accrual.Valid {
		ord.Accrual = &sum
	}

	return &ord, nil
}

func fetchOrders(ctx context.Context, rows pgx.Rows) (order.Orders, error) {
	defer rows.Close()

	_, logger := logging.GetOrCreateLogger(ctx)

	orders := make(order.Orders, 0)
	for rows.Next() {
		ord, err := fetchOrder(rows)
		if err != nil {
			logger.Err(err).Msg("failed to fetch orders")
			return nil, err
		}
		orders = append(orders, ord)
	}
	if err := rows.Err(); err != nil {
		logger.Err(err).Msg("failed to fetch orders")
		return nil, err
	}

	logger.Trace().Msgf("got %d records", len(orders))
	return orders, nil
}

func Orders(conn *Connection) storage.Orders {
	return &ordersStorage{
		Connection: conn,
		timeout:    DefaultTimeout,
	}
}
