package migrate

import "github.com/go-pg/migrations/v8"

const v01up = `
CREATE TABLE users (
	id SERIAL NOT NULL CONSTRAINT users_pk PRIMARY KEY,
	login VARCHAR(255) NOT NULL,
	password VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX users_id_uindex ON users (id);
CREATE UNIQUE INDEX users_login_uindex ON users (login);`

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(v01up)
		return
	})
}
