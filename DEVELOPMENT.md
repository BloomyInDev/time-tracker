# Development

## Requirements

- Go 1.26 or later.

No CGO is required. The app uses the pure-Go
[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver, and
[`templ`](https://templ.guide) is pinned in `go.mod` and run through `go tool`,
so there is nothing else to install.

## Build and run

```sh
make run     # generate templ code, build, then start the server
make build   # build only, producing bin/time-tracker
./bin/time-tracker serve
```

`make build` regenerates the templ views before compiling. To generate them on
their own, run `go tool templ generate`.

## Live reload

```sh
make dev
```

This runs `templ generate --watch` alongside [air](https://github.com/air-verse/air)
so both the templates and the server rebuild on change.

## Tests

```sh
make test    # go test ./...
```

## Project layout

```
cmd/time-tracker       single binary: serve, register, export-users
internal/db            SQLite access, schema and migrations
internal/handlers      HTTP handlers
internal/templates     templ views (compiled to _templ.go)
internal/models        data types
internal/i18n          locale loading; locales/*.yml
internal/assets        embedded static CSS and JS
internal/config        environment-based configuration
internal/service/auth  login, sessions and middleware
```

## Continuous integration

- `.github/workflows/build.yml` builds cross-compiled binaries for Linux, macOS
  and Windows on x86 and ARM, and publishes a multi-arch Docker image to GHCR.
- `.github/workflows/release.yml` builds the same target matrix and attaches the
  tarballs to a GitHub Release when a `v*` tag is pushed.
