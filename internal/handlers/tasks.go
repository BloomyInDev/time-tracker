package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
)

func ListTasks(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		tasks, err := db.ListTasks(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, tasks)
	}
}

type taskBody struct {
	ClientID   int64     `json:"client_id"`
	TaskTypeID int64     `json:"task_type_id"`
	Title      string    `json:"title"`
	HoursSpent float64   `json:"hours_spent"`
	Date       time.Time `json:"date"`
}

func CreateTask(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())

		var body taskBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		task, err := db.CreateTask(conn, models.Task{
			UserID:     userID,
			ClientID:   body.ClientID,
			TaskTypeID: body.TaskTypeID,
			Title:      body.Title,
			HoursSpent: body.HoursSpent,
			Date:       body.Date,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, task)
	}
}

func UpdateTask(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var body taskBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		err = db.UpdateTask(conn, models.Task{
			ID:         id,
			UserID:     userID,
			ClientID:   body.ClientID,
			TaskTypeID: body.TaskTypeID,
			Title:      body.Title,
			HoursSpent: body.HoursSpent,
			Date:       body.Date,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
		w.WriteHeader(http.StatusNoContent)
	}
}
