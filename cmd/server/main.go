package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/handlers"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	authSvc := auth.NewService(conn, cfg.JWTSecret)

	log.Printf("listening on port %d", cfg.Port)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.Home)
	mux.HandleFunc("GET /login", handlers.Login)
	mux.HandleFunc("POST /login", handlers.LoginSubmit(authSvc))
	mux.HandleFunc("GET /logout", handlers.Logout)

	mux.Handle("GET /clients", authSvc.RequireAuth(handlers.ListClients(conn)))
	mux.Handle("POST /clients", authSvc.RequireAuth(handlers.CreateClient(conn)))
	mux.Handle("GET /clients/{id}", authSvc.RequireAuth(handlers.ClientDetail(conn)))
	mux.Handle("POST /clients/{id}/delete", authSvc.RequireAuth(handlers.DeleteClient(conn)))
	mux.Handle("POST /clients/{id}/task-types", authSvc.RequireAuth(handlers.AssignTaskType(conn)))
	mux.Handle("POST /clients/{id}/task-types/{taskTypeID}/delete", authSvc.RequireAuth(handlers.UnassignTaskType(conn)))

	mux.Handle("GET /task-types", authSvc.RequireAuth(handlers.ListTaskTypes(conn)))
	mux.Handle("POST /task-types", authSvc.RequireAuth(handlers.CreateTaskType(conn)))
	mux.Handle("POST /task-types/{id}/delete", authSvc.RequireAuth(handlers.DeleteTaskType(conn)))

	mux.Handle("GET /tasks", authSvc.RequireAuth(handlers.ListTasks(conn)))
	mux.Handle("POST /tasks", authSvc.RequireAuth(handlers.CreateTask(conn)))
	mux.Handle("POST /tasks/{id}/delete", authSvc.RequireAuth(handlers.DeleteTask(conn)))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux))
}
