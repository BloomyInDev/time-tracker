package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/bloomyindev/time-tracker/migrations"

	sqlite "modernc.org/sqlite"
)

// ErrInUse is returned when a row can't be deleted because another row still
// references it (foreign key constraint).
var ErrInUse = errors.New("resource is still referenced by other records")

// sqliteConstraintForeignKey is SQLite's extended result code
// SQLITE_CONSTRAINT_FOREIGNKEY, returned when a delete/update violates a
// foreign key constraint.
const sqliteConstraintForeignKey = 787

// isForeignKeyErr reports whether err is a foreign-key constraint violation.
func isForeignKeyErr(err error) bool {
	var se *sqlite.Error
	return errors.As(err, &se) && se.Code() == sqliteConstraintForeignKey
}

// Options tunes the SQLite connection.
type Options struct {
	// WAL switches on write-ahead logging, which stops readers and writers
	// from blocking each other. It costs two extra files next to the
	// database (-wal and -shm), so it stays opt-in: capping the pool at a
	// single connection already removes contention within this process.
	WAL bool
}

// Open connects to the SQLite database at path with foreign keys enforced,
// limits writes to a single connection (SQLite allows one writer), and applies
// all embedded migrations in alphabetical order.
//
// The pragmas live in the DSN rather than in a PRAGMA statement issued after
// connecting, because database/sql pools connections: a statement configures
// only whichever connection happened to serve it, leaving every other
// connection in the pool on the defaults.
func Open(path string, opts Options) (*sql.DB, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	if opts.WAL {
		dsn += "&_pragma=journal_mode(WAL)"
	}

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect sqlite: %w", err)
	}
	// SQLite permits only one writer at a time; a single connection avoids
	// "database is locked" errors under concurrent writes. Every
	// transaction must therefore run its statements on the tx, never on
	// the pool, or it would wait on a connection it is itself holding.
	conn.SetMaxOpenConns(1)

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// migrate applies embedded migration files in alphabetical order, skipping any
// already recorded in schema_migrations. The first migration is written to be
// idempotent, so it runs safely against databases that predate this ledger.
func migrate(conn *sql.DB) error {
	if _, err := conn.Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`,
	); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var applied int
		if err := conn.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied > 0 {
			continue
		}
		stmts, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := conn.Exec(string(stmts)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := conn.Exec(
			`INSERT INTO schema_migrations (name) VALUES (?)`, name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}
