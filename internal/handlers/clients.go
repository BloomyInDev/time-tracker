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

		periods, err := db.ListPeriods(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var selectedPeriodID int64
		if raw := r.URL.Query().Get("period_id"); raw != "" {
			selectedPeriodID, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "invalid period_id", http.StatusBadRequest)
				return
			}
		}
		var selectedTaskTypeID int64
		if raw := r.URL.Query().Get("task_type_id"); raw != "" {
			selectedTaskTypeID, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "invalid task_type_id", http.StatusBadRequest)
				return
			}
		}

		tasks, err := db.ListTasksByClientFiltered(conn, userID, id, selectedPeriodID, selectedTaskTypeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var totalHours float64
		hoursByType := make(map[int64]float64)
		for _, t := range tasks {
			totalHours += t.HoursSpent
			hoursByType[t.TaskTypeID] += t.HoursSpent
		}

		templates.ClientDetail(client, taskTypeChoices, tasks, totalHours, hoursByType, allTypes, periods, selectedPeriodID, selectedTaskTypeID).Render(r.Context(), w)
	}
}

// clientFilterID reads an optional int64 query param; a blank value means
// "no filter" rather than an error.
func clientFilterID(r *http.Request, key string) (int64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

// ClientReport renders a print-friendly page for a client: the total hours on
// top, then one table per task type ("project"), honoring the active
// period/task-type filters.
func ClientReport(conn *sql.DB) http.HandlerFunc {
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

		periodID, err := clientFilterID(r, "period_id")
		if err != nil {
			http.Error(w, "invalid period_id", http.StatusBadRequest)
			return
		}
		taskTypeID, err := clientFilterID(r, "task_type_id")
		if err != nil {
			http.Error(w, "invalid task_type_id", http.StatusBadRequest)
			return
		}

		tasks, err := db.ListTasksByClientFiltered(conn, userID, id, periodID, taskTypeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		allTypes, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		periods, err := db.ListPeriods(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// One table per task type, in the app's task-type order, keeping
		// only types that actually have tasks in the filtered set.
		var total float64
		tasksByType := make(map[int64][]models.Task)
		for _, t := range tasks {
			total += t.HoursSpent
			tasksByType[t.TaskTypeID] = append(tasksByType[t.TaskTypeID], t)
		}
		var groups []templates.ClientTypeGroup
		for _, tt := range allTypes {
			ts, ok := tasksByType[tt.ID]
			if !ok {
				continue
			}
			var h float64
			for _, t := range ts {
				h += t.HoursSpent
			}
			groups = append(groups, templates.ClientTypeGroup{Name: tt.Name, Tasks: ts, Hours: h})
		}

		var periodLabel string
		for _, p := range periods {
			if p.ID == periodID {
				periodLabel = p.Name
			}
		}
		var taskTypeLabel string
		for _, tt := range allTypes {
			if tt.ID == taskTypeID {
				taskTypeLabel = tt.Name
			}
		}

		view := templates.ClientReportView{
			ClientName:   client.Name,
			PeriodName:   periodLabel,
			TaskTypeName: taskTypeLabel,
			Total:        total,
			Groups:       groups,
			Periods:      periods,
		}
		templates.ClientReport(view).Render(r.Context(), w)
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
