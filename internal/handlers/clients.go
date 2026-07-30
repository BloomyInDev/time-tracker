package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) ClientsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listClients)
	r.Post("/", h.createClient)
	r.Get("/{id}", h.clientDetail)
	r.Get("/{id}/report", h.clientReport)
	r.Get("/{id}/edit", h.editClientForm)
	r.Post("/{id}/edit", h.updateClient)
	r.Post("/{id}/delete", h.deleteClient)
	return r
}

func (h *Handlers) listClients(w http.ResponseWriter, r *http.Request) {
	clients, err := db.ListClientsOrderedByName(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	templates.Clients(clients).Render(r.Context(), w)
}

func (h *Handlers) createClient(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	if _, err := db.CreateClient(h.DB, userID(r), r.FormValue("name")); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

// taskTypeChoices lists the user's task types, each flagged with whether
// it is currently assigned to the given client.
func taskTypeChoices(conn *sql.DB, userID, clientID int64) ([]templates.TaskTypeChoice, error) {
	assigned, err := db.ListTaskTypesForClient(conn, clientID)
	if err != nil {
		return nil, err
	}
	assignedIDs := make(map[int64]bool, len(assigned))
	for _, t := range assigned {
		assignedIDs[t.ID] = true
	}

	allTypes, err := db.ListTaskTypes(conn, userID)
	if err != nil {
		return nil, err
	}
	choices := make([]templates.TaskTypeChoice, len(allTypes))
	for i, t := range allTypes {
		choices[i] = templates.TaskTypeChoice{TaskType: t, Assigned: assignedIDs[t.ID]}
	}
	return choices, nil
}

// editClientForm renders the single page that edits everything about a
// client: its name, its archived state and its allowed task types.
func (h *Handlers) editClientForm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	client, err := db.GetClient(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	choices, err := taskTypeChoices(h.DB, userID(r), id)
	if err != nil {
		fail(w, err)
		return
	}
	templates.EditClient(client, choices).Render(r.Context(), w)
}

// updateClient saves the whole edit form: name, archived flag and the
// client's allowed task types.
func (h *Handlers) updateClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := db.GetClient(h.DB, userID(r), id); err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	if !parseForm(w, r) {
		return
	}

	if err := db.UpdateClient(h.DB, userID(r), id, r.FormValue("name"), r.FormValue("is_archived") == "1"); err != nil {
		fail(w, err)
		return
	}
	if err := syncClientTaskTypes(h.DB, userID(r), id, r.Form["task_type_id"]); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

// syncClientTaskTypes assigns exactly the checked task types to the
// client and unassigns the others.
func syncClientTaskTypes(conn *sql.DB, userID, clientID int64, checkedIDs []string) error {
	checked := make(map[int64]bool, len(checkedIDs))
	for _, raw := range checkedIDs {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		checked[id] = true
	}

	allTypes, err := db.ListTaskTypes(conn, userID)
	if err != nil {
		return err
	}
	for _, t := range allTypes {
		if checked[t.ID] {
			err = db.AssignTaskTypeToClient(conn, clientID, t.ID)
		} else {
			err = db.UnassignTaskTypeFromClient(conn, clientID, t.ID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handlers) deleteClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := db.DeleteClient(h.DB, userID(r), id); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

func (h *Handlers) clientDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	client, err := db.GetClient(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	allTypes, err := db.ListTaskTypes(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}

	periods, err := db.ListPeriods(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}

	selectedPeriodID, err := clientFilterID(r, "period_id")
	if err != nil {
		http.Error(w, "invalid period_id", http.StatusBadRequest)
		return
	}
	selectedTaskTypeID, err := clientFilterID(r, "task_type_id")
	if err != nil {
		http.Error(w, "invalid task_type_id", http.StatusBadRequest)
		return
	}

	tasks, err := db.ListTasksByClientFiltered(h.DB, userID(r), id, selectedPeriodID, selectedTaskTypeID)
	if err != nil {
		fail(w, err)
		return
	}
	var totalHours float64
	hoursByType := make(map[int64]float64)
	for _, t := range tasks {
		totalHours += t.HoursSpent
		hoursByType[t.TaskTypeID] += t.HoursSpent
	}

	templates.ClientDetail(client, tasks, totalHours, hoursByType, allTypes, periods, selectedPeriodID, selectedTaskTypeID).Render(r.Context(), w)
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

// clientReport renders a print-friendly page for a client: the total hours on
// top, then one table per task type ("project"), honoring the active
// period/task-type filters.
func (h *Handlers) clientReport(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	client, err := db.GetClient(h.DB, userID(r), id)
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

	tasks, err := db.ListTasksByClientFiltered(h.DB, userID(r), id, periodID, taskTypeID)
	if err != nil {
		fail(w, err)
		return
	}
	allTypes, err := db.ListTaskTypes(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	periods, err := db.ListPeriods(h.DB, userID(r))
	if err != nil {
		fail(w, err)
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
		var hrs float64
		for _, t := range ts {
			hrs += t.HoursSpent
		}
		groups = append(groups, templates.ClientTypeGroup{Name: tt.Name, Tasks: ts, Hours: hrs})
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
