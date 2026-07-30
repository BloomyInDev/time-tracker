-- 0001_init: the schema as it stood before migrations were tracked in a
-- ledger. Every statement is idempotent, because databases created by the
-- older startup code already have all of this: they ran a CREATE TABLE IF NOT
-- EXISTS block plus a list of ALTER TABLE ADD COLUMN statements on every open.
-- The columns those ALTERs added are inlined here, so a fresh database ends up
-- with exactly the same shape an existing one already has.

CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    hours_mon     DOUBLE NOT NULL DEFAULT 0,
    hours_tue     DOUBLE NOT NULL DEFAULT 0,
    hours_wed     DOUBLE NOT NULL DEFAULT 0,
    hours_thu     DOUBLE NOT NULL DEFAULT 0,
    hours_fri     DOUBLE NOT NULL DEFAULT 0,
    hours_sat     DOUBLE NOT NULL DEFAULT 0,
    hours_sun     DOUBLE NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS clients (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    is_archived INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS task_types (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS task_types_for_client (
    client_id    INTEGER NOT NULL REFERENCES clients(id),
    task_type_id INTEGER NOT NULL REFERENCES task_types(id),
    PRIMARY KEY (client_id, task_type_id)
);

CREATE TABLE IF NOT EXISTS periods (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    name       TEXT NOT NULL,
    is_default INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id      INTEGER NOT NULL REFERENCES users(id),
    client_id    INTEGER NOT NULL REFERENCES clients(id),
    task_type_id INTEGER NOT NULL REFERENCES task_types(id),
    title        TEXT NOT NULL,
    hours_spent  DOUBLE NOT NULL,
    date         DATE NOT NULL,
    period_id    INTEGER REFERENCES periods(id)
);
