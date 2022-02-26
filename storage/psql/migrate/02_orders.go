package migrate

import "github.com/go-pg/migrations/v8"

const v02up = `
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
	id BIGSERIAL NOT NULL CONSTRAINT orders_pk PRIMARY KEY,
    number VARCHAR(64) NOT NULL,
    user_id INT NOT NULL CONSTRAINT orders_users_id_fk REFERENCES users ON DELETE CASCADE,
	status order_status NOT NULL,
	uploaded_at TIMESTAMPTZ NOT NULL,
	accrual DOUBLE PRECISION
);

CREATE UNIQUE INDEX orders_id_uindex ON orders (id);
CREATE UNIQUE INDEX orders_number_uindex ON orders (number);`

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(v02up)
		return
	})
}
