package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/assets"
	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/handlers"
	"github.com/bloomyindev/time-tracker/internal/i18n"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
)

func main() {
	cfg := config.Load()

	dbPath := flag.String("db-path", cfg.DBPath, "path to the sqlite database file (env: DB_PATH)")
	flag.Parse()
	cfg.DBPath = *dbPath

	if err := i18n.Load(); err != nil {
		log.Fatalf("load locales: %v", err)
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	authSvc := auth.NewService(conn, cfg.JWTSecret)

	log.Printf("listening on port %d", cfg.Port)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.Home)
	mux.HandleFunc("GET /lang/{code}", handlers.SetLocale)
	mux.HandleFunc("GET /login", handlers.Login)
	mux.HandleFunc("POST /login", handlers.LoginSubmit(authSvc))
	mux.HandleFunc("GET /logout", handlers.Logout(authSvc))

	mux.Handle("GET /clients", authSvc.RequireAuth(handlers.ListClients(conn)))
	mux.Handle("POST /clients", authSvc.RequireAuth(handlers.CreateClient(conn)))
	mux.Handle("GET /clients/{id}", authSvc.RequireAuth(handlers.ClientDetail(conn)))
	mux.Handle("POST /clients/{id}/delete", authSvc.RequireAuth(handlers.DeleteClient(conn)))
	mux.Handle("POST /clients/{id}/task-types", authSvc.RequireAuth(handlers.SyncClientTaskTypes(conn)))

	mux.Handle("GET /task-types", authSvc.RequireAuth(handlers.ListTaskTypes(conn)))
	mux.Handle("POST /task-types", authSvc.RequireAuth(handlers.CreateTaskType(conn)))
	mux.Handle("POST /task-types/{id}/delete", authSvc.RequireAuth(handlers.DeleteTaskType(conn)))

	mux.Handle("GET /periods", authSvc.RequireAuth(handlers.ListPeriods(conn)))
	mux.Handle("POST /periods", authSvc.RequireAuth(handlers.CreatePeriod(conn)))
	mux.Handle("POST /periods/{id}/default", authSvc.RequireAuth(handlers.SetDefaultPeriod(conn)))
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
		log.Fatalf("mount static assets: %v", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), i18n.Middleware(mux)))
}
