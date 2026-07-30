package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/bloomyindev/time-tracker/internal/assets"
	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/handlers"
	"github.com/bloomyindev/time-tracker/internal/i18n"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "time-tracker",
		Usage: "self-hosted time tracker: web server and admin tools",
		Commands: []*cli.Command{
			serveCommand(),
			registerCommand(),
			exportUsersCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// dbPathFlag overrides the database path (env: DB_PATH). A fresh flag is
// returned per command so each owns its own value.
func dbPathFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "db-path",
		Usage:   "path to the sqlite database file",
		Value:   config.Load().DBPath,
		Sources: cli.EnvVars("DB_PATH"),
	}
}

func openDB(cmd *cli.Command) (*sql.DB, error) {
	return db.Open(cmd.String("db-path"))
}

func serveCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "run the web server",
		Flags: []cli.Flag{dbPathFlag()},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := config.Load()

			if err := i18n.Load(); err != nil {
				return fmt.Errorf("load locales: %w", err)
			}

			conn, err := openDB(cmd)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}

			staticFS, err := fs.Sub(assets.Static, "static")
			if err != nil {
				return fmt.Errorf("mount static assets: %w", err)
			}

			h := handlers.New(conn, auth.NewService(conn, cfg.JWTSecret))

			log.Printf("listening on port %d", cfg.Port)
			return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), h.Router(staticFS))
		},
	}
}
