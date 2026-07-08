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

	mux.Handle("GET /api/clients", authSvc.RequireAuth(handlers.ListClients(conn)))
	mux.Handle("POST /api/clients", authSvc.RequireAuth(handlers.CreateClient(conn)))
	mux.Handle("PUT /api/clients/{id}", authSvc.RequireAuth(handlers.UpdateClient(conn)))
	mux.Handle("DELETE /api/clients/{id}", authSvc.RequireAuth(handlers.DeleteClient(conn)))

	mux.Handle("GET /api/task-types", authSvc.RequireAuth(handlers.ListTaskTypes(conn)))
	mux.Handle("POST /api/task-types", authSvc.RequireAuth(handlers.CreateTaskType(conn)))
	mux.Handle("PUT /api/task-types/{id}", authSvc.RequireAuth(handlers.UpdateTaskType(conn)))
	mux.Handle("DELETE /api/task-types/{id}", authSvc.RequireAuth(handlers.DeleteTaskType(conn)))

	mux.Handle("GET /api/tasks", authSvc.RequireAuth(handlers.ListTasks(conn)))
	mux.Handle("POST /api/tasks", authSvc.RequireAuth(handlers.CreateTask(conn)))
	mux.Handle("PUT /api/tasks/{id}", authSvc.RequireAuth(handlers.UpdateTask(conn)))
	mux.Handle("DELETE /api/tasks/{id}", authSvc.RequireAuth(handlers.DeleteTask(conn)))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux))
}
