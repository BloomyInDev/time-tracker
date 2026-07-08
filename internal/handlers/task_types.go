package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
)

func ListTaskTypes(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		types, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templates.TaskTypes(types).Render(r.Context(), w)
	}
}

func CreateTaskType(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if _, err := db.CreateTaskType(conn, userID, r.FormValue("name")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/task-types", http.StatusSeeOther)
	}
}

func DeleteTaskType(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := db.DeleteTaskType(conn, userID, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/task-types", http.StatusSeeOther)
	}
}
