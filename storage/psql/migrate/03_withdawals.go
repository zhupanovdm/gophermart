package migrate

import "github.com/go-pg/migrations/v8"

const v03up = `
CREATE TABLE withdrawals (
	id BIGSERIAL NOT NULL CONSTRAINT withdrawals_pk PRIMARY KEY,
	order_id BIGINT NOT NULL CONSTRAINT withdrawals_orders_id_fk REFERENCES orders ON DELETE CASCADE,
	processed_at TIMESTAMPTZ NOT NULL,
	sum DOUBLE PRECISION NOT NULL
);`

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(v03up)
		return
	})
}
