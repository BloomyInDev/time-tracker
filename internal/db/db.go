package db

import (
	"database/sql"
	"strings"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS clients (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	user_id INTEGER NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS task_types (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id),
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS task_types_for_client (
	client_id INTEGER NOT NULL REFERENCES clients(id),
	task_type_id INTEGER NOT NULL REFERENCES task_types(id),
	PRIMARY KEY (client_id, task_type_id)
);

CREATE TABLE IF NOT EXISTS periods (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id),
	name TEXT NOT NULL,
	is_default INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id),
	client_id INTEGER NOT NULL REFERENCES clients(id),
	task_type_id INTEGER NOT NULL REFERENCES task_types(id),
	title TEXT NOT NULL,
	hours_spent DOUBLE NOT NULL,
	date DATE NOT NULL
);
`

// migrations holds schema changes that ALTER existing tables, which
// CREATE TABLE IF NOT EXISTS can't express. Each is safe to run
// repeatedly: "duplicate column" errors from an already-applied
// migration are ignored.
var migrations = []string{
	`ALTER TABLE tasks ADD COLUMN period_id INTEGER REFERENCES periods(id)`,
	`ALTER TABLE users ADD COLUMN hours_mon DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_tue DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_wed DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_thu DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_fri DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_sat DOUBLE NOT NULL DEFAULT 0`,
	`ALTER TABLE users ADD COLUMN hours_sun DOUBLE NOT NULL DEFAULT 0`,
}

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	for _, m := range migrations {
		if _, err := conn.Exec(m); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}
