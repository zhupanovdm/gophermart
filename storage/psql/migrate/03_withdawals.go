package migrate

import "github.com/go-pg/migrations/v8"

const v03up = `
CREATE TABLE withdrawals (
	id BIGSERIAL NOT NULL CONSTRAINT withdrawals_pk PRIMARY KEY,
	order_number VARCHAR(64) NOT NULL,
    user_id INT NOT NULL CONSTRAINT withdrawals_users_id_fk REFERENCES users ON DELETE CASCADE,
	processed_at TIMESTAMPTZ NOT NULL,
	sum DOUBLE PRECISION NOT NULL
);`

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(v03up)
		return
	})
}
