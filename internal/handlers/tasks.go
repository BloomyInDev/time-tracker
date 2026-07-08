package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
)

func ListTasks(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())

		tasks, err := db.ListTasks(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		clients, err := db.ListClients(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		types, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.Tasks(tasks, clients, types).Render(r.Context(), w)
	}
}

func CreateTask(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		clientID, err := strconv.ParseInt(r.FormValue("client_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid client_id", http.StatusBadRequest)
			return
		}
		taskTypeID, err := strconv.ParseInt(r.FormValue("task_type_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid task_type_id", http.StatusBadRequest)
			return
		}
		hoursSpent, err := strconv.ParseFloat(r.FormValue("hours_spent"), 64)
		if err != nil {
			http.Error(w, "invalid hours_spent", http.StatusBadRequest)
			return
		}
		date, err := time.Parse("2006-01-02", r.FormValue("date"))
		if err != nil {
			http.Error(w, "invalid date", http.StatusBadRequest)
			return
		}

		_, err = db.CreateTask(conn, models.Task{
			UserID:     userID,
			ClientID:   clientID,
			TaskTypeID: taskTypeID,
			Title:      r.FormValue("title"),
			HoursSpent: hoursSpent,
			Date:       date,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
	}
}

func DeleteTask(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := db.DeleteTask(conn, userID, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
	}
}
