package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/zhupanovdm/gophermart/model"
	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/storage"
)

const (
	ordersStorageName = "Orders PSQL Storage"

	ordersQuery = `SELECT number, status, accrual, user_id, uploaded_at FROM orders`
)

var _ storage.Orders = (*ordersStorage)(nil)

type ordersStorage struct {
	*Connection
	timeout time.Duration
}

func (o *ordersStorage) Create(ctx context.Context, ordNew *order.Order) (ord *order.Order, ok bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(ordNew))
	logger.Info().Msg("storing new order")

	err = o.BeginTxFunc(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		if ord, err = orderByNumber(ctx, tx, ordNew.Number); err != nil {
			logger.Err(err).Msg("failed to check existing order with the same number")
			return err
		}
		if ord != nil {
			logger.Warn().Msg("order withe same number already exists")
			return nil
		}
		if _, err = tx.Exec(ctx, `INSERT INTO orders(number, user_id, status, uploaded_at) VALUES($1, $2, $3, $4)`,
			ordNew.Number, ordNew.UserID, ordNew.Status, ordNew.UploadedAt); err != nil {
			logger.Err(err).Msg("failed to persist new order")
			return err
		}
		ok = true
		ord = ordNew
		logger.Trace().Msg("success")
		return nil
	})
	return
}

func (o *ordersStorage) OrdersByUser(ctx context.Context, userID user.ID) (order.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(userID))
	logger.Info().Msg("querying client orders")

	rows, err := o.Query(ctx, ordersQuery+` WHERE user_id = $1 ORDER BY uploaded_at`, userID)
	if err != nil {
		logger.Err(err).Msg("failed to query client orders")
		return nil, err
	}
	return fetchOrders(rows)
}

func (o *ordersStorage) OrdersByStatus(ctx context.Context, statuses ...order.Status) (order.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName)
	logger.Info().Msg("querying orders by status")

	rows, err := o.Query(ctx, ordersQuery+` WHERE status = ANY($1) ORDER BY uploaded_at`, order.StatusesToStrings(statuses))
	if err != nil {
		logger.Err(err).Msg("failed to query orders")
		return nil, err
	}
	return fetchOrders(rows)
}

func (o *ordersStorage) Update(ctx context.Context, number order.Number, status order.Status, accrual *model.Sum) error {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ctx, logger := logging.ServiceLogger(ctx, ordersStorageName, logging.With(number))
	logger.Info().Msg("querying client orders")

	return o.BeginFunc(ctx, func(tx pgx.Tx) error {
		ord, err := orderByNumber(ctx, tx, number)
		if err != nil {
			logger.Err(err).Msg("failed to read existing order")
			return err
		}
		if ord == nil {
			err := fmt.Errorf("order with number doesn't exist: %v", number)
			logger.Err(err).Msg("failed to read existing order")
			return err
		}
		if change := ord.Status.Compare(status); change == 0 {
			logger.Warn().Msg("status is unchanged: skipping update")
			return nil
		} else if change < 0 {
			err := fmt.Errorf("can't descend order status: from %v to %v", ord.Status, status)
			logger.Err(err).Msg("failed to update order")
			return err
		}
		if _, err := o.Exec(ctx, "UPDATE orders SET status = $2, accrual = $3 WHERE number = $1", number, status, accrual); err != nil {
			logger.Err(err).Msg("failed to update order")
			return err
		}
		logger.Trace().Msg("success")
		return nil
	})
}

func orderByNumber(ctx context.Context, db queryExecutor, number order.Number) (*order.Order, error) {
	return fetchOrder(db.QueryRow(ctx, ordersQuery+` WHERE number = $1`, number))
}

func fetchOrder(s rowScanner) (*order.Order, error) {
	var ord order.Order
	var accrual sql.NullFloat64
	err := s.Scan(&ord.Number, &ord.Status, &accrual, &ord.UserID, &ord.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	sum := model.Sum(accrual.Float64)
	if accrual.Valid {
		ord.Accrual = &sum
	}
	return &ord, nil
}

func fetchOrders(rows pgx.Rows) (order.Orders, error) {
	defer rows.Close()

	orders := make(order.Orders, 0)
	for rows.Next() {
		ord, err := fetchOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, ord)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func Orders(conn *Connection) storage.Orders {
	return &ordersStorage{
		Connection: conn,
		timeout:    DefaultTimeout,
	}
}
