package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
)

func ListClients(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		clients, err := db.ListClients(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templates.Clients(clients).Render(r.Context(), w)
	}
}

func CreateClient(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if _, err := db.CreateClient(conn, userID, r.FormValue("name")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
	}
}

func DeleteClient(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := db.DeleteClient(conn, userID, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
	}
}

func ClientDetail(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		client, err := db.GetClient(conn, userID, id)
		if err != nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		assigned, err := db.ListTaskTypesForClient(conn, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		allTypes, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		assignedIDs := make(map[int64]bool, len(assigned))
		for _, t := range assigned {
			assignedIDs[t.ID] = true
		}
		var available []models.TaskType
		for _, t := range allTypes {
			if !assignedIDs[t.ID] {
				available = append(available, t)
			}
		}

		templates.ClientDetail(client, assigned, available).Render(r.Context(), w)
	}
}

func AssignTaskType(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		clientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if _, err := db.GetClient(conn, userID, clientID); err != nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		taskTypeID, err := strconv.ParseInt(r.FormValue("task_type_id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid task_type_id", http.StatusBadRequest)
			return
		}
		if _, err := db.GetTaskType(conn, userID, taskTypeID); err != nil {
			http.Error(w, "task type not found", http.StatusNotFound)
			return
		}

		if err := db.AssignTaskTypeToClient(conn, clientID, taskTypeID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/clients/"+strconv.FormatInt(clientID, 10), http.StatusSeeOther)
	}
}

func UnassignTaskType(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		clientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if _, err := db.GetClient(conn, userID, clientID); err != nil {
			http.Error(w, "client not found", http.StatusNotFound)
			return
		}
		taskTypeID, err := strconv.ParseInt(r.PathValue("taskTypeID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid task_type_id", http.StatusBadRequest)
			return
		}

		if err := db.UnassignTaskTypeFromClient(conn, clientID, taskTypeID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/clients/"+strconv.FormatInt(clientID, 10), http.StatusSeeOther)
	}
}
