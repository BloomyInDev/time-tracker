package handlers

import (
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) PeriodsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listPeriods)
	r.Post("/", h.createPeriod)
	r.Post("/{id}/default", h.setDefaultPeriod)
	r.Get("/{id}/edit", h.editPeriodForm)
	r.Post("/{id}/rename", h.renamePeriod)
	r.Post("/{id}/delete", h.deletePeriod)
	return r
}

func (h *Handlers) listPeriods(w http.ResponseWriter, r *http.Request) {
	periods, err := db.ListPeriods(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	templates.Periods(periods).Render(r.Context(), w)
}

func (h *Handlers) createPeriod(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	if _, err := db.CreatePeriod(h.DB, userID(r), r.FormValue("name")); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/periods", http.StatusSeeOther)
}

func (h *Handlers) setDefaultPeriod(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := db.SetDefaultPeriod(h.DB, userID(r), id); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/periods", http.StatusSeeOther)
}

func (h *Handlers) editPeriodForm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	period, err := db.GetPeriod(h.DB, userID(r), id)
	if err != nil {
		http.Error(w, "period not found", http.StatusNotFound)
		return
	}
	templates.EditPeriod(period).Render(r.Context(), w)
}

func (h *Handlers) renamePeriod(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if !parseForm(w, r) {
		return
	}
	if err := db.UpdatePeriod(h.DB, userID(r), id, r.FormValue("name")); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/periods", http.StatusSeeOther)
}

func (h *Handlers) deletePeriod(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := db.DeletePeriod(h.DB, userID(r), id); err != nil {
		fail(w, err)
		return
	}
	http.Redirect(w, r, "/periods", http.StatusSeeOther)
}
