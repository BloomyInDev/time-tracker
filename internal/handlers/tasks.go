package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) TasksRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listTasks)
	r.Post("/", h.createTask)
	r.Get("/{id}/edit", h.editTaskForm)
	r.Post("/{id}/update", h.updateTask)
	r.Post("/{id}/delete", h.deleteTask)
	return r
}

func (h *Handlers) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.ListTasks(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	clients, err := db.ListClientsOrderedByName(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	types, err := db.ListTaskTypes(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	periods, err := db.ListPeriods(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	byClient, err := db.ListTaskTypesByClient(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}

	var defaultPeriodID int64
	if p, err := db.GetDefaultPeriod(h.DB, userID(r)); err == nil {
		defaultPeriodID = p.ID
	}

	// The dropdown starts on the first client the form offers, so render
	// that client's task types straight away; task-filter.js takes over
	// from there.
	initial := selectableTaskTypes(types, byClient[firstUnarchivedClientID(clients)])

	templates.Tasks(clients, types, initial, periods, byClient, groupByDay(tasks), time.Now().Format("2006-01-02"), defaultPeriodID).Render(r.Context(), w)
}

// firstUnarchivedClientID returns the client the task form preselects, or
// 0 when the user has no client to log against.
func firstUnarchivedClientID(clients []models.Client) int64 {
	for _, c := range clients {
		if !c.IsArchived {
			return c.ID
		}
	}
	return 0
}

// selectableTaskTypes narrows the user's task types to those assigned to a
// client. A client with none assigned accepts them all, so the full list
// comes back unchanged.
func selectableTaskTypes(types []models.TaskType, allowedIDs []int64) []models.TaskType {
	if len(allowedIDs) == 0 {
		return types
	}
	allowed := make(map[int64]bool, len(allowedIDs))
	for _, id := range allowedIDs {
		allowed[id] = true
	}
	selectable := make([]models.TaskType, 0, len(allowedIDs))
	for _, t := range types {
		if allowed[t.ID] {
			selectable = append(selectable, t)
		}
	}
	return selectable
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

// clientAcceptsTasks rejects archived clients: they keep their history
// but must not receive new tasks.
func clientAcceptsTasks(conn *sql.DB, userID, clientID int64) (bool, error) {
	client, err := db.GetClient(conn, userID, clientID)
	if err != nil {
		return false, err
	}
	return !client.IsArchived, nil
}

// taskTypeAllowedForClient enforces that a task type belongs to the user
// and, when the client has task types configured, that it is one of them.
// A client with none configured accepts any of the user's task types.
func taskTypeAllowedForClient(conn *sql.DB, userID, clientID, taskTypeID int64) (bool, error) {
	if _, err := db.GetTaskType(conn, userID, taskTypeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	allowed, err := db.ListTaskTypesForClient(conn, userID, clientID)
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

// taskForm holds the fields shared by the create and update forms.
type taskForm struct {
	clientID   int64
	taskTypeID int64
	periodID   int64
	title      string
	hoursSpent float64
	date       time.Time
}

// parseTaskForm reads and validates the task form, writing its own 400
// and reporting false on the first bad field.
func parseTaskForm(w http.ResponseWriter, r *http.Request) (taskForm, bool) {
	if !parseForm(w, r) {
		return taskForm{}, false
	}

	var f taskForm
	var err error
	if f.clientID, err = strconv.ParseInt(r.FormValue("client_id"), 10, 64); err != nil {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return taskForm{}, false
	}
	if f.taskTypeID, err = strconv.ParseInt(r.FormValue("task_type_id"), 10, 64); err != nil {
		http.Error(w, "invalid task_type_id", http.StatusBadRequest)
		return taskForm{}, false
	}
	if f.hoursSpent, err = strconv.ParseFloat(r.FormValue("hours_spent"), 64); err != nil {
		http.Error(w, "invalid hours_spent", http.StatusBadRequest)
		return taskForm{}, false
	}
	if f.date, err = time.Parse("2006-01-02", r.FormValue("date")); err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return taskForm{}, false
	}
	if f.periodID, err = parsePeriodID(r); err != nil {
		http.Error(w, "invalid period_id", http.StatusBadRequest)
		return taskForm{}, false
	}
	f.title = r.FormValue("title")
	return f, true
}

func (h *Handlers) createTask(w http.ResponseWriter, r *http.Request) {
	f, ok := parseTaskForm(w, r)
	if !ok {
		return
	}

	active, err := clientAcceptsTasks(h.DB, userID(r), f.clientID)
	if err != nil {
		http.Error(w, "client not found", http.StatusBadRequest)
		return
	}
	if !active {
		http.Error(w, "client is archived", http.StatusBadRequest)
		return
	}

	allowed, err := taskTypeAllowedForClient(h.DB, userID(r), f.clientID, f.taskTypeID)
	if err != nil {
		fail(w, err)
		return
	}
	if !allowed {
		http.Error(w, "task type not allowed for this client", http.StatusBadRequest)
		return
	}

	_, err = db.CreateTask(h.DB, models.Task{
		UserID:     userID(r),
		ClientID:   f.clientID,
		TaskTypeID: f.taskTypeID,
		PeriodID:   f.periodID,
		Title:      f.title,
		HoursSpent: f.hoursSpent,
		Date:       f.date,
	})
	if err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handlers) editTaskForm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	task, err := db.GetTask(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	clients, err := db.ListClients(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	types, err := db.ListTaskTypes(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	periods, err := db.ListPeriods(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	byClient, err := db.ListTaskTypesByClient(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}

	templates.EditTask(task, clients, types, periods, byClient).Render(r.Context(), w)
}

func (h *Handlers) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	existing, err := db.GetTask(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	f, ok := parseTaskForm(w, r)
	if !ok {
		return
	}

	// Moving a task onto an archived client is a new assignment, so
	// it's refused; a task already on an archived client stays
	// editable.
	if f.clientID != existing.ClientID {
		active, err := clientAcceptsTasks(h.DB, userID(r), f.clientID)
		if err != nil {
			http.Error(w, "client not found", http.StatusBadRequest)
			return
		}
		if !active {
			http.Error(w, "client is archived", http.StatusBadRequest)
			return
		}
	}

	// Same idea for the task type: a pair that predates the client's
	// current configuration stays editable, so long as the edit doesn't
	// change either side of it.
	if f.clientID != existing.ClientID || f.taskTypeID != existing.TaskTypeID {
		allowed, err := taskTypeAllowedForClient(h.DB, userID(r), f.clientID, f.taskTypeID)
		if err != nil {
			fail(w, err)
			return
		}
		if !allowed {
			http.Error(w, "task type not allowed for this client", http.StatusBadRequest)
			return
		}
	}

	err = db.UpdateTask(h.DB, models.Task{
		ID:         id,
		UserID:     userID(r),
		ClientID:   f.clientID,
		TaskTypeID: f.taskTypeID,
		PeriodID:   f.periodID,
		Title:      f.title,
		HoursSpent: f.hoursSpent,
		Date:       f.date,
	})
	if err != nil {
		fail(w, err)
		return
	}
	// Anchor the reload on the edited row so the browser restores the
	// scroll position instead of jumping to the top of the list.
	http.Redirect(w, r, "/tasks#task-"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handlers) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := db.DeleteTask(h.DB, userID(r), id); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}
