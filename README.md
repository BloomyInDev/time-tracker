# time-tracker

A small, self-hosted time-tracking web app. Log hours per client, task type
and period, track daily-hours targets and overtime, and produce printable
per-client and per-period reports.

It ships as a single Go binary with server-side rendered
[templ](https://templ.guide) views, a [Bulma](https://bulma.io) UI, and an
embedded SQLite database. Static assets and templates are compiled into the
binary, so a deployment is one file. That same binary runs the web server
(`time-tracker serve`) and the admin commands (`register`, `export-users`).

## Features

- **Tasks:** log time entries with a title, client, task type, period, date
  and hours. Entries are grouped by day with per-day totals.
- **Clients, task types and periods:** manage each from its own page, with
  dedicated edit pages to rename them. A client can restrict which task types
  it allows, and one period can be marked as the default.
- **Daily targets and overtime:** set target hours per weekday under
  *My account*. The *Time* page shows each day's hours against its target,
  grouped by month with monthly subtotals.
- **Date-range filter:** sum hours, target and overtime between two dates. An
  open-ended range runs up to today.
- **Printable reports:** standalone, print-friendly pages that open the browser
  print dialog for save-as-PDF.
  - `/time/report` produces a day-by-day time report over the selected range.
  - `/clients/{id}/report` produces a per-client report: a summary table of
    hours per task type, followed by one detail table per task type, honoring
    the active filters.
- **CSV export:** `/time/report.csv` and `/clients/{id}/report.csv` return the
  same data as the reports, `;`-separated with decimal commas so they open
  directly in Excel or LibreOffice with a French locale.
- **Bilingual:** English and French (`/lang/en`, `/lang/fr`).
- **Vocabulary presets:** each user picks, under *My account*, the wording of
  the UI. `Projects` renames Clients as Projects, `Clients` is the default.
  A preset is a `presets.<name>` block in `internal/i18n/locales/*.yml` that
  lists only the keys it renames, plus its name in `i18n.Presets`.
- **Authentication:** email and password login backed by opaque, server-side
  session tokens stored in a cookie. Sessions live in memory, so they are lost
  on restart and can be revoked instantly.

## Install

Choose one of the options below, then follow [First run](#first-run) to create
a user and log in.

### Docker (recommended)

A multi-arch image is published to GHCR on every push to `master` and on
version tags, for `linux/amd64`, `linux/arm64` and `linux/arm/v7`.

```sh
docker run -d -p 8080:8080 \
  -e TRACKER_JWT_SECRET=change-me \
  -e TRACKER_SECURE_COOKIES=false \
  -v time-tracker-data:/data \
  ghcr.io/bloomyindev/time-tracker:latest
```

The container runs `time-tracker serve` by default. The database is stored in
the `/data` volume (`TRACKER_DB_PATH=/data/time-tracker.db`). Run admin commands inside
the container with `docker exec <container> /app/time-tracker <command>` (see
[CLI](#cli)).

An example [`compose.yml`](compose.yml) is included:

```sh
docker compose up -d
docker compose exec time-tracker /app/time-tracker register \
  --email you@example.com --password secret
```

### Prebuilt binary

Download a tarball for your OS and architecture from the
[Releases](https://github.com/bloomyindev/time-tracker/releases) page. Each
archive contains the single `time-tracker` binary. Prebuilt targets:

| OS      | Architectures |
|---------|---------------|
| Linux   | amd64, arm64, armv7 |
| macOS   | amd64, arm64  |
| Windows | amd64, arm64  |

Other targets Go supports (for example 386, riscv64, ppc64le, s390x, loong64 or
the BSDs) should work too: the binary is pure Go with no CGO. We just don't
publish binaries or Docker images for them, so build from source instead.

```sh
tar -xzf time-tracker_*_linux_amd64.tar.gz
cd time-tracker_*_linux_amd64
TRACKER_JWT_SECRET=change-me TRACKER_SECURE_COOKIES=false ./time-tracker serve
```

### From source

Building from source requires Go 1.26 or later. See
[DEVELOPMENT.md](DEVELOPMENT.md) for the full setup.

## First run

1. **Start the server.** It listens on <http://localhost:8080>. In production,
   always set `TRACKER_JWT_SECRET` (see [Configuration](#configuration)).
2. **Create a user.** There is no public sign-up:
   ```sh
   ./time-tracker register --email you@example.com --password secret
   # from source: ./bin/time-tracker register ...
   # docker:      docker exec <container> /app/time-tracker register ...
   ```
3. **Log in** at <http://localhost:8080/login> with those credentials.

## Configuration

The app is configured through environment variables:

| Variable                  | Default                | Description                                                                                          |
|---------------------------|------------------------|------------------------------------------------------------------------------------------------------|
| `TRACKER_PORT`            | `8080`                 | TCP port the server listens on.                                                                      |
| `TRACKER_DB_PATH`         | `time-tracker.db`      | Path to the SQLite database file.                                                                    |
| `TRACKER_JWT_SECRET`      | `dev-secret-change-me` | Secret for signing JWTs used by the (currently unused) bearer-token API flow. Set this in production. |
| `TRACKER_SECURE_COOKIES`  | `true`                 | Mark cookies `Secure`, so browsers only send them over HTTPS. Set to `false` to serve plain HTTP.     |
| `TRACKER_SQLITE_WAL`      | `false`                | Enable SQLite write-ahead logging. Adds `-wal` and `-shm` files next to the database.                 |

**Serving over plain HTTP?** Set `TRACKER_SECURE_COOKIES=false`. The default
assumes a TLS-terminating reverse proxy in front. Left on over plain HTTP, the
browser accepts the session cookie and then never sends it back, so logging in
appears to do nothing.

The database path can also be passed with `--db-path`, which takes precedence
over the environment. Migrations are embedded in the binary and applied on startup;
each one is recorded in a `schema_migrations` table so it runs exactly once.

## CLI

The admin commands live in the same binary as the server:

```sh
time-tracker serve                                           # run the web server
time-tracker register --email <email> --password <password>  # create a user
time-tracker export-users                                    # dump users as JSON (no password hashes)
```

Every command accepts `--db-path` (or the `TRACKER_DB_PATH` environment variable).

## Contributing

Development setup, build instructions and the project layout are documented in
[DEVELOPMENT.md](DEVELOPMENT.md).

## License

GNU AGPL v3. See [LICENSE](LICENSE).
