# time-tracker

A small, self-hosted time-tracking web app. Log hours per client, task type
and period; track daily-hours targets and overtime; and produce printable
per-client and per-period reports.

Built as a single Go binary with server-side rendered [templ](https://templ.guide)
views, a [Bulma](https://bulma.io) UI, and an embedded SQLite database. Static
assets and templates are compiled into the binary, so deployment is one file.
The same binary runs the web server (`time-tracker serve`) and the admin
commands (`register`, `export-users`).

## Features

- **Tasks** — log time entries with a title, client, task type, period, date
  and hours; grouped by day with per-day totals.
- **Clients / task types / periods** — manage each from its own page, with
  dedicated edit pages to rename them. A client can restrict which task types
  it allows. One period can be marked default.
- **Daily targets & overtime** — set target hours per weekday (in *My account*);
  the *Time* page shows each day's hours against its target, grouped by month
  with monthly subtotals.
- **Date-range filter** — sum hours, target and overtime between two dates
  (an open-ended range runs up to today).
- **Printable reports** — standalone print-friendly pages that open the browser
  print dialog for save-as-PDF:
  - `/time/report` — day-by-day time report over the selected range.
  - `/clients/{id}/report` — per-client report: a summary table of hours per
    task type plus one detail table per task type, honoring the active filters.
- **Bilingual** — English and French (`/lang/en`, `/lang/fr`).
- **Auth** — email/password login with opaque, server-side session tokens set
  as a cookie. Sessions live in memory (lost on restart, instantly revocable).

## Install

Pick one of the three options below, then follow [First run](#first-run) to
create a user and log in.

### Docker (recommended)

A multi-arch image (`linux/amd64`, `linux/arm64`) is published to GHCR on every
push to `master` and on version tags.

```sh
docker run -d -p 8080:8080 \
  -e JWT_SECRET=change-me \
  -v time-tracker-data:/data \
  ghcr.io/bloomyindev/time-tracker:latest
```

The container runs `time-tracker serve` by default. The database is stored in
the `/data` volume (`DB_PATH=/data/time-tracker.db`). Run admin commands in the
container with `docker exec <container> /app/time-tracker <command>` (see
[CLI](#cli)).

### Prebuilt binary

Download a tarball for your OS/arch from the
[Releases](https://github.com/bloomyindev/time-tracker/releases) page (Linux,
macOS and Windows, on x86 and ARM). Each archive contains the single
`time-tracker` binary.

```sh
tar -xzf time-tracker_*_linux_amd64.tar.gz
cd time-tracker_*_linux_amd64
JWT_SECRET=change-me ./time-tracker serve
```

### From source

Requires **Go 1.26+**. No CGO — the app uses the pure-Go
[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver, and `templ`
is pinned in `go.mod` (run via `go tool`), so nothing else to install.

```sh
make run          # generate templ code, build, and start the server
make build        # just build -> bin/time-tracker
./bin/time-tracker serve
```

## First run

1. **Start the server.** It listens on <http://localhost:8080>. In production,
   always set `JWT_SECRET` (see [Configuration](#configuration)).
2. **Create a user** (there is no public sign-up):
   ```sh
   ./time-tracker register --email you@example.com --password secret
   # from source: ./bin/time-tracker register ...
   # docker:      docker exec <container> /app/time-tracker register ...
   ```
3. **Log in** at <http://localhost:8080/login> with those credentials.

## Development

```sh
make dev   # templ --watch + air live reload
```

## Configuration

Configured through environment variables:

| Variable     | Default              | Description                          |
|--------------|----------------------|--------------------------------------|
| `DB_PATH`    | `time-tracker.db`    | Path to the SQLite database file.    |
| `JWT_SECRET` | `dev-secret-change-me` | Secret for signing JWTs used by the (currently unused) bearer-token API flow. **Set this in production.** |

The server always listens on port `8080`. The DB path can also be passed with
`-db-path`. The schema is created and migrated automatically on startup.

## CLI

The admin commands live in the same binary as the server:

```sh
time-tracker serve                                             # run the web server
time-tracker register --email <email> --password <password>    # create a user
time-tracker export-users                                      # dump users as JSON (no password hashes)
```

All commands accept `--db-path` (or the `DB_PATH` env var).

## Project layout

```
cmd/time-tracker  single binary: serve, register, export-users
internal/db       SQLite access + schema/migrations
internal/handlers HTTP handlers
internal/templates templ views (compiled to _templ.go)
internal/models   data types
internal/i18n     locale loading; locales/*.yml
internal/assets   embedded static CSS/JS
internal/config   env-based configuration
internal/service/auth  login, sessions, middleware
```

## License

GNU AGPL v3 — see [LICENSE](LICENSE).
