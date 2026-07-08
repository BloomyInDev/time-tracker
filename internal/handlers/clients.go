package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
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
		assignedIDs := make(map[int64]bool, len(assigned))
		for _, t := range assigned {
			assignedIDs[t.ID] = true
		}

		allTypes, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		taskTypeChoices := make([]templates.TaskTypeChoice, len(allTypes))
		for i, t := range allTypes {
			taskTypeChoices[i] = templates.TaskTypeChoice{TaskType: t, Assigned: assignedIDs[t.ID]}
		}

		tasks, err := db.ListTasksByClient(conn, userID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var totalHours float64
		for _, t := range tasks {
			totalHours += t.HoursSpent
		}

		templates.ClientDetail(client, taskTypeChoices, tasks, totalHours).Render(r.Context(), w)
	}
}

func SyncClientTaskTypes(conn *sql.DB) http.HandlerFunc {
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

		checked := make(map[int64]bool)
		for _, raw := range r.Form["task_type_id"] {
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "invalid task_type_id", http.StatusBadRequest)
				return
			}
			checked[id] = true
		}

		allTypes, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, t := range allTypes {
			if checked[t.ID] {
				err = db.AssignTaskTypeToClient(conn, clientID, t.ID)
			} else {
				err = db.UnassignTaskTypeFromClient(conn, clientID, t.ID)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/clients/"+strconv.FormatInt(clientID, 10), http.StatusSeeOther)
	}
}
