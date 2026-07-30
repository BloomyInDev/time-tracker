package handlers

import (
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) TaskTypesRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listTaskTypes)
	r.Post("/", h.createTaskType)
	r.Get("/{id}/edit", h.editTaskTypeForm)
	r.Post("/{id}/rename", h.renameTaskType)
	r.Post("/{id}/delete", h.deleteTaskType)
	return r
}

func (h *Handlers) listTaskTypes(w http.ResponseWriter, r *http.Request) {
	types, err := db.ListTaskTypes(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	templates.TaskTypes(types).Render(r.Context(), w)
}

func (h *Handlers) createTaskType(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	if _, err := db.CreateTaskType(h.DB, userID(r), r.FormValue("name")); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/task-types", http.StatusSeeOther)
}

func (h *Handlers) editTaskTypeForm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	taskType, err := db.GetTaskType(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "task type not found", http.StatusNotFound)
		return
	}
	templates.EditTaskType(taskType).Render(r.Context(), w)
}

func (h *Handlers) renameTaskType(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if !parseForm(w, r) {
		return
	}
	if err := db.UpdateTaskType(h.DB, userID(r), id, r.FormValue("name")); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/task-types", http.StatusSeeOther)
}

func (h *Handlers) deleteTaskType(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := db.DeleteTaskType(h.DB, userID(r), id); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/task-types", http.StatusSeeOther)
}
