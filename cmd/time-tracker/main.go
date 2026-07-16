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

			authSvc := auth.NewService(conn, cfg.JWTSecret)

			mux := http.NewServeMux()

			mux.HandleFunc("GET /", handlers.Home)
			mux.HandleFunc("GET /lang/{code}", handlers.SetLocale)
			mux.HandleFunc("GET /login", handlers.Login)
			mux.HandleFunc("POST /login", handlers.LoginSubmit(authSvc))
			mux.HandleFunc("GET /logout", handlers.Logout(authSvc))

			mux.Handle("GET /clients", authSvc.RequireAuth(handlers.ListClients(conn)))
			mux.Handle("POST /clients", authSvc.RequireAuth(handlers.CreateClient(conn)))
			mux.Handle("GET /clients/{id}", authSvc.RequireAuth(handlers.ClientDetail(conn)))
			mux.Handle("GET /clients/{id}/report", authSvc.RequireAuth(handlers.ClientReport(conn)))
			mux.Handle("GET /clients/{id}/edit", authSvc.RequireAuth(handlers.EditClientForm(conn)))
			mux.Handle("POST /clients/{id}/rename", authSvc.RequireAuth(handlers.RenameClient(conn)))
			mux.Handle("POST /clients/{id}/delete", authSvc.RequireAuth(handlers.DeleteClient(conn)))
			mux.Handle("POST /clients/{id}/task-types", authSvc.RequireAuth(handlers.SyncClientTaskTypes(conn)))

			mux.Handle("GET /task-types", authSvc.RequireAuth(handlers.ListTaskTypes(conn)))
			mux.Handle("POST /task-types", authSvc.RequireAuth(handlers.CreateTaskType(conn)))
			mux.Handle("GET /task-types/{id}/edit", authSvc.RequireAuth(handlers.EditTaskTypeForm(conn)))
			mux.Handle("POST /task-types/{id}/rename", authSvc.RequireAuth(handlers.RenameTaskType(conn)))
			mux.Handle("POST /task-types/{id}/delete", authSvc.RequireAuth(handlers.DeleteTaskType(conn)))

			mux.Handle("GET /periods", authSvc.RequireAuth(handlers.ListPeriods(conn)))
			mux.Handle("POST /periods", authSvc.RequireAuth(handlers.CreatePeriod(conn)))
			mux.Handle("POST /periods/{id}/default", authSvc.RequireAuth(handlers.SetDefaultPeriod(conn)))
			mux.Handle("GET /periods/{id}/edit", authSvc.RequireAuth(handlers.EditPeriodForm(conn)))
			mux.Handle("POST /periods/{id}/rename", authSvc.RequireAuth(handlers.RenamePeriod(conn)))
			mux.Handle("POST /periods/{id}/delete", authSvc.RequireAuth(handlers.DeletePeriod(conn)))

			mux.Handle("GET /tasks", authSvc.RequireAuth(handlers.ListTasks(conn)))
			mux.Handle("POST /tasks", authSvc.RequireAuth(handlers.CreateTask(conn)))
			mux.Handle("GET /tasks/{id}/edit", authSvc.RequireAuth(handlers.EditTaskForm(conn)))
			mux.Handle("POST /tasks/{id}/update", authSvc.RequireAuth(handlers.UpdateTask(conn)))
			mux.Handle("POST /tasks/{id}/delete", authSvc.RequireAuth(handlers.DeleteTask(conn)))

			mux.Handle("GET /time", authSvc.RequireAuth(handlers.ListTimeEntries(conn)))
			mux.Handle("GET /time/report", authSvc.RequireAuth(handlers.TimeReport(conn)))

			mux.Handle("GET /account", authSvc.RequireAuth(handlers.Account(conn)))
			mux.Handle("POST /account/hours", authSvc.RequireAuth(handlers.UpdateDailyHours(conn)))
			mux.Handle("POST /account/password", authSvc.RequireAuth(handlers.ChangePassword(conn, authSvc)))

			staticFS, err := fs.Sub(assets.Static, "static")
			if err != nil {
				return fmt.Errorf("mount static assets: %w", err)
			}
			mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

			log.Printf("listening on port %d", cfg.Port)
			return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), i18n.Middleware(mux))
		},
	}
}
