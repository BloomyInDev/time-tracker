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
		clients, err := db.ListClientsOrderedByName(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		types, err := db.ListTaskTypes(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		periods, err := db.ListPeriods(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		byClient, err := db.ListTaskTypesByClient(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var defaultPeriodID int64
		if p, err := db.GetDefaultPeriod(conn, userID); err == nil {
			defaultPeriodID = p.ID
		}

		templates.Tasks(clients, types, periods, byClient, groupByDay(tasks), time.Now().Format("2006-01-02"), defaultPeriodID).Render(r.Context(), w)
	}
}

// parsePeriodID reads an optional period_id form value; a blank or
// missing value means "no period" rather than a validation error.
func parsePeriodID(r *http.Request) (int64, error) {
	raw := r.FormValue("period_id")
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

// groupByDay buckets tasks (already ordered by date desc) into per-day
// groups with a running total, for a single table showing each day's
// total followed by that day's tasks.
func groupByDay(tasks []models.Task) []templates.DayGroup {
	today := time.Now().Format("2006-01-02")
	var groups []templates.DayGroup
	for _, t := range tasks {
		date := t.Date.Format("02/01/2006")
		if len(groups) > 0 && groups[len(groups)-1].Date == date {
			last := &groups[len(groups)-1]
			last.Hours += t.HoursSpent
			last.Tasks = append(last.Tasks, t)
			continue
		}
		groups = append(groups, templates.DayGroup{
			Date:   date,
			Hours:  t.HoursSpent,
			Tasks:  []models.Task{t},
			Future: t.Date.Format("2006-01-02") > today,
		})
	}
	return groups
}

// taskTypeAllowedForClient enforces that a task type is one of the
// client's configured task types, when the client has any configured.
func taskTypeAllowedForClient(conn *sql.DB, clientID, taskTypeID int64) (bool, error) {
	allowed, err := db.ListTaskTypesForClient(conn, clientID)
	if err != nil {
		return false, err
	}
	if len(allowed) == 0 {
		return true, nil
	}
	for _, t := range allowed {
		if t.ID == taskTypeID {
			return true, nil
		}
	}
	return false, nil
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
		periodID, err := parsePeriodID(r)
		if err != nil {
			http.Error(w, "invalid period_id", http.StatusBadRequest)
			return
		}

		ok, err := taskTypeAllowedForClient(conn, clientID, taskTypeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "task type not allowed for this client", http.StatusBadRequest)
			return
		}

		_, err = db.CreateTask(conn, models.Task{
			UserID:     userID,
			ClientID:   clientID,
			TaskTypeID: taskTypeID,
			PeriodID:   periodID,
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

func EditTaskForm(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		task, err := db.GetTask(conn, userID, id)
		if err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
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
		periods, err := db.ListPeriods(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.EditTask(task, clients, types, periods).Render(r.Context(), w)
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
		if _, err := db.GetTask(conn, userID, id); err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

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
		periodID, err := parsePeriodID(r)
		if err != nil {
			http.Error(w, "invalid period_id", http.StatusBadRequest)
			return
		}

		ok, err := taskTypeAllowedForClient(conn, clientID, taskTypeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "task type not allowed for this client", http.StatusBadRequest)
			return
		}

		err = db.UpdateTask(conn, models.Task{
			ID:         id,
			UserID:     userID,
			ClientID:   clientID,
			TaskTypeID: taskTypeID,
			PeriodID:   periodID,
			Title:      r.FormValue("title"),
			HoursSpent: hoursSpent,
			Date:       date,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Anchor the reload on the edited row so the browser restores the
		// scroll position instead of jumping to the top of the list.
		http.Redirect(w, r, "/tasks#task-"+strconv.FormatInt(id, 10), http.StatusSeeOther)
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
