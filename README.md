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
- **Bilingual:** English and French (`/lang/en`, `/lang/fr`).
- **Authentication:** email and password login backed by opaque, server-side
  session tokens stored in a cookie. Sessions live in memory, so they are lost
  on restart and can be revoked instantly.

## Install

Choose one of the options below, then follow [First run](#first-run) to create
a user and log in.

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
the `/data` volume (`DB_PATH=/data/time-tracker.db`). Run admin commands inside
the container with `docker exec <container> /app/time-tracker <command>` (see
[CLI](#cli)).

### Prebuilt binary

Download a tarball for your OS and architecture from the
[Releases](https://github.com/bloomyindev/time-tracker/releases) page (Linux,
macOS and Windows, on x86 and ARM). Each archive contains the single
`time-tracker` binary.

```sh
tar -xzf time-tracker_*_linux_amd64.tar.gz
cd time-tracker_*_linux_amd64
JWT_SECRET=change-me ./time-tracker serve
```

### From source

Building from source requires Go 1.26 or later. See
[DEVELOPMENT.md](DEVELOPMENT.md) for the full setup.

## First run

1. **Start the server.** It listens on <http://localhost:8080>. In production,
   always set `JWT_SECRET` (see [Configuration](#configuration)).
2. **Create a user.** There is no public sign-up:
   ```sh
   ./time-tracker register --email you@example.com --password secret
   # from source: ./bin/time-tracker register ...
   # docker:      docker exec <container> /app/time-tracker register ...
   ```
3. **Log in** at <http://localhost:8080/login> with those credentials.

## Configuration

The app is configured through environment variables:

| Variable     | Default                | Description                                                                                 |
|--------------|------------------------|---------------------------------------------------------------------------------------------|
| `DB_PATH`    | `time-tracker.db`      | Path to the SQLite database file.                                                           |
| `JWT_SECRET` | `dev-secret-change-me` | Secret for signing JWTs used by the (currently unused) bearer-token API flow. Set this in production. |

The server always listens on port `8080`. The database path can also be passed
with `--db-path`. The schema is created and migrated automatically on startup.

## CLI

The admin commands live in the same binary as the server:

```sh
time-tracker serve                                           # run the web server
time-tracker register --email <email> --password <password>  # create a user
time-tracker export-users                                    # dump users as JSON (no password hashes)
```

Every command accepts `--db-path` (or the `DB_PATH` environment variable).

## Contributing

Development setup, build instructions and the project layout are documented in
[DEVELOPMENT.md](DEVELOPMENT.md).

## License

GNU AGPL v3. See [LICENSE](LICENSE).
