package db

import (
	"database/sql"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/bloomyindev/time-tracker/migrations"
	_ "modernc.org/sqlite"
)

// assertAllMigrationsRecorded checks the ledger holds one row per embedded
// migration file, so adding a migration never means editing this file.
func assertAllMigrationsRecorded(t *testing.T, conn *sql.DB) {
	t.Helper()
	files, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	if n := count(t, conn, `SELECT count(*) FROM schema_migrations`); n != len(files) {
		t.Errorf("schema_migrations = %d rows, want %d (one per migration file)", n, len(files))
	}
}

// open builds a migrated database in a temp dir.
func open(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"), Options{})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// seed creates one user with a client, a task type assigned to it, a period
// and a task tying them together.
func seed(t *testing.T, conn *sql.DB) {
	t.Helper()
	stmts := []string{
		`INSERT INTO users (id, email, password_hash) VALUES (1, 'a@b.c', 'x')`,
		`INSERT INTO clients (id, name, user_id) VALUES (1, 'Acme', 1)`,
		`INSERT INTO task_types (id, user_id, name) VALUES (1, 1, 'Dev')`,
		`INSERT INTO periods (id, user_id, name, is_default) VALUES (1, 1, 'Q1', 0)`,
		`INSERT INTO task_types_for_client (client_id, task_type_id) VALUES (1, 1)`,
		`INSERT INTO tasks (id, user_id, client_id, task_type_id, period_id, title, hours_spent, date)
		 VALUES (1, 1, 1, 1, 1, 'work', 2, '2026-01-01')`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(s); err != nil {
			t.Fatalf("seed %q: %v", s, err)
		}
	}
}

func count(t *testing.T, conn *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

// TestForeignKeysOnEveryConnection is the point of putting the pragma in the
// DSN. Setting it with a PRAGMA statement configures only the connection that
// served the statement, and database/sql hands out others freely.
func TestForeignKeysOnEveryConnection(t *testing.T) {
	conn := open(t)
	conn.SetMaxIdleConns(0) // force a fresh connection each round

	for i := range 5 {
		var on int
		if err := conn.QueryRow(`PRAGMA foreign_keys`).Scan(&on); err != nil {
			t.Fatal(err)
		}
		if on != 1 {
			t.Errorf("connection %d: foreign_keys = %d, want 1", i, on)
		}
	}
}

// TestWALIsOptIn guards the default: WAL leaves -wal and -shm files next to
// the database, so it only happens when asked for.
func TestWALIsOptIn(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{"default", Options{}, "delete"},
		{"opt in", Options{WAL: true}, "wal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := Open(filepath.Join(t.TempDir(), "test.db"), tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			var mode string
			if err := conn.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
				t.Fatal(err)
			}
			if mode != tt.want {
				t.Errorf("journal_mode = %q, want %q", mode, tt.want)
			}
		})
	}
}

// TestBadReferenceRejected covers what foreign keys buy beyond tidy deletes:
// a task can no longer point at a task type that doesn't exist, or at one
// belonging to somebody else.
func TestBadReferenceRejected(t *testing.T) {
	conn := open(t)
	seed(t, conn)

	_, err := conn.Exec(
		`INSERT INTO tasks (user_id, client_id, task_type_id, title, hours_spent, date)
		 VALUES (1, 1, 999, 'work', 1, '2026-01-01')`)
	if !isForeignKeyErr(err) {
		t.Errorf("insert with an unknown task_type_id = %v, want a foreign key error", err)
	}
}

// TestDeleteRefusedWhileReferenced covers the ErrInUse paths. Tasks are the
// records the app exists to keep, so nothing that still has hours logged
// against it can be deleted out from under them.
func TestDeleteRefusedWhileReferenced(t *testing.T) {
	tests := []struct {
		name   string
		remove func(*sql.DB) error
	}{
		{"client", func(c *sql.DB) error { return DeleteClient(c, 1, 1) }},
		{"task type", func(c *sql.DB) error { return DeleteTaskType(c, 1, 1) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := open(t)
			seed(t, conn)

			if err := tt.remove(conn); !errors.Is(err, ErrInUse) {
				t.Fatalf("delete = %v, want ErrInUse", err)
			}
			if n := count(t, conn, `SELECT count(*) FROM tasks`); n != 1 {
				t.Errorf("tasks = %d, want the task left untouched", n)
			}
		})
	}
}

// TestDeleteCascadesAssignments checks migration 0002: the join table is
// bookkeeping, so a delete that is otherwise allowed shouldn't be blocked by
// a checkbox someone ticked on an edit page.
func TestDeleteCascadesAssignments(t *testing.T) {
	conn := open(t)
	seed(t, conn)

	if _, err := conn.Exec(`DELETE FROM tasks`); err != nil {
		t.Fatal(err)
	}
	if err := DeleteClient(conn, 1, 1); err != nil {
		t.Fatalf("DeleteClient after removing its tasks: %v", err)
	}
	if n := count(t, conn, `SELECT count(*) FROM task_types_for_client`); n != 0 {
		t.Errorf("task_types_for_client = %d rows, want the assignment cascaded away", n)
	}
}

// TestDeletePeriodDetachesTasks covers the one delete that is allowed to
// proceed while tasks reference it: a period only groups hours, so removing it
// must not remove them.
func TestDeletePeriodDetachesTasks(t *testing.T) {
	conn := open(t)
	seed(t, conn)

	if err := DeletePeriod(conn, 1, 1); err != nil {
		t.Fatalf("DeletePeriod: %v", err)
	}
	if n := count(t, conn, `SELECT count(*) FROM periods`); n != 0 {
		t.Errorf("periods = %d, want 0", n)
	}
	if n := count(t, conn, `SELECT count(*) FROM tasks WHERE period_id IS NULL`); n != 1 {
		t.Errorf("detached tasks = %d, want the task kept with no period", n)
	}
}

// TestMigrationsAreIdempotent reopens the same file: every migration is
// already recorded, so the second open must be a no-op rather than an error.
func TestMigrationsAreIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	conn, err := Open(path, Options{})
	if err != nil {
		t.Fatal(err)
	}
	seed(t, conn)
	conn.Close()

	conn, err = Open(path, Options{})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer conn.Close()

	if n := count(t, conn, `SELECT count(*) FROM tasks`); n != 1 {
		t.Errorf("tasks = %d, want the seeded row intact", n)
	}
	assertAllMigrationsRecorded(t, conn)
}

// legacySchema is the schema exactly as the pre-ledger startup code left it:
// a CREATE TABLE block plus a list of ALTER TABLE ADD COLUMN statements, run
// on every open. Databases in the wild are in this shape, and migration 0001
// has to recognise them as already migrated.
const legacySchema = `
CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL);
CREATE TABLE clients (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, user_id INTEGER NOT NULL REFERENCES users(id));
CREATE TABLE task_types (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL REFERENCES users(id), name TEXT NOT NULL);
CREATE TABLE task_types_for_client (client_id INTEGER NOT NULL REFERENCES clients(id), task_type_id INTEGER NOT NULL REFERENCES task_types(id), PRIMARY KEY (client_id, task_type_id));
CREATE TABLE periods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL REFERENCES users(id), name TEXT NOT NULL, is_default INTEGER NOT NULL DEFAULT 0);
CREATE TABLE tasks (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL REFERENCES users(id), client_id INTEGER NOT NULL REFERENCES clients(id), task_type_id INTEGER NOT NULL REFERENCES task_types(id), title TEXT NOT NULL, hours_spent DOUBLE NOT NULL, date DATE NOT NULL);
ALTER TABLE tasks ADD COLUMN period_id INTEGER REFERENCES periods(id);
ALTER TABLE users ADD COLUMN hours_mon DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_tue DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_wed DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_thu DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_fri DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_sat DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN hours_sun DOUBLE NOT NULL DEFAULT 0;
ALTER TABLE clients ADD COLUMN is_archived INTEGER NOT NULL DEFAULT 0;
`

// TestMigratesLegacyDatabase is the upgrade path for existing installs: an
// untracked database keeps its rows, gains the ledger, and comes out with the
// cascade from 0002 applied.
func TestMigratesLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	seed(t, legacy)
	legacy.Close()

	conn, err := Open(path, Options{})
	if err != nil {
		t.Fatalf("migrating a legacy database: %v", err)
	}
	defer conn.Close()

	if n := count(t, conn, `SELECT count(*) FROM tasks`); n != 1 {
		t.Errorf("tasks = %d, want the existing row preserved", n)
	}
	if n := count(t, conn, `SELECT count(*) FROM task_types_for_client`); n != 1 {
		t.Errorf("assignments = %d, want the rebuild to have copied it", n)
	}
	// The user's daily-hours columns came from the old ALTERs; 0001 must
	// not have replaced the table and dropped them.
	if n := count(t, conn, `SELECT count(*) FROM pragma_table_info('users') WHERE name LIKE 'hours_%'`); n != 7 {
		t.Errorf("hours_* columns = %d, want 7", n)
	}
	assertAllMigrationsRecorded(t, conn)

	// The cascade only exists if 0002 actually rebuilt the table.
	if _, err := conn.Exec(`DELETE FROM tasks`); err != nil {
		t.Fatal(err)
	}
	if err := DeleteClient(conn, 1, 1); err != nil {
		t.Fatalf("DeleteClient: %v", err)
	}
	if n := count(t, conn, `SELECT count(*) FROM task_types_for_client`); n != 0 {
		t.Errorf("task_types_for_client = %d rows, want the assignment cascaded away", n)
	}
}

func TestTimeSettingsRoundTrip(t *testing.T) {
	conn, err := Open(filepath.Join(t.TempDir(), "settings.db"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO users (id, email, password_hash) VALUES (1, 'a@b.c', 'x')`); err != nil {
		t.Fatal(err)
	}

	u, err := GetUser(conn, 1)
	if err != nil || u.TimeStartDate != "" {
		t.Fatalf("new user: start = %q, err = %v, want empty", u.TimeStartDate, err)
	}

	hours := [7]float64{7, 7, 7, 7, 7, 0, 0}
	if err := UpdateTimeSettings(conn, 1, hours, "2026-09-01"); err != nil {
		t.Fatal(err)
	}
	u, _ = GetUser(conn, 1)
	if u.TimeStartDate != "2026-09-01" || u.DailyHours != hours {
		t.Errorf("after save: start = %q, hours = %v", u.TimeStartDate, u.DailyHours)
	}

	if err := UpdateTimeSettings(conn, 1, hours, ""); err != nil {
		t.Fatal(err)
	}
	u, _ = GetUser(conn, 1)
	if u.TimeStartDate != "" {
		t.Errorf("after clearing: start = %q, want empty", u.TimeStartDate)
	}
}

func TestVocabularyDefaultsAndRoundTrip(t *testing.T) {
	conn := open(t)
	seed(t, conn)

	u, err := GetUser(conn, 1)
	if err != nil {
		t.Fatal(err)
	}
	if u.Vocabulary != "default" {
		t.Errorf("new user vocabulary = %q, want default", u.Vocabulary)
	}

	if err := UpdateVocabulary(conn, 1, "projects"); err != nil {
		t.Fatal(err)
	}
	if u, _ = GetUser(conn, 1); u.Vocabulary != "projects" {
		t.Errorf("vocabulary = %q, want projects", u.Vocabulary)
	}
}
