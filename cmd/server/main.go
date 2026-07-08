package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/handlers"
)

func main() {
	cfg := config.Load()

	_, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	log.Printf("listening on port %d", cfg.Port)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.Home)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux))
}
