package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/go-chi/chi/v5"
	"github.com/invopop/ctxi18n/i18n"
)

func (h *Handlers) AccountRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.account)
	r.Post("/hours", h.updateDailyHours)
	r.Post("/password", h.changePassword)
	return r
}

func (h *Handlers) account(w http.ResponseWriter, r *http.Request) {
	user, err := db.GetUser(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	templates.Account(user, "", "").Render(r.Context(), w)
}

func (h *Handlers) updateDailyHours(w http.ResponseWriter, r *http.Request) {
	user, err := db.GetUser(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if !parseForm(w, r) {
		return
	}

	var hours [7]float64
	for i := 0; i < 7; i++ {
		raw := r.FormValue("hours_" + strconv.Itoa(i))
		if raw == "" {
			continue
		}
		hrs, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			http.Error(w, "invalid hours", http.StatusBadRequest)
			return
		}
		hours[i] = hrs
	}

	if err := db.UpdateDailyHours(h.DB, userID(r), hours); err != nil {
		fail(w, err)
		return
	}
	user.DailyHours = hours
	templates.Account(user, "", i18n.T(r.Context(), "account.hours_saved")).Render(r.Context(), w)
}

func (h *Handlers) changePassword(w http.ResponseWriter, r *http.Request) {
	user, err := db.GetUser(h.DB, userID(r))
	if err != nil {
		fail(w, err)
		return
	}
	if !parseForm(w, r) {
		return
	}

	err = h.Auth.ChangePassword(userID(r), r.FormValue("current_password"), r.FormValue("new_password"))
	if errors.Is(err, auth.ErrInvalidCredentials) {
		templates.Account(user, i18n.T(r.Context(), "account.wrong_current_password"), "").Render(r.Context(), w)
		return
	}
	if err != nil {
		templates.Account(user, i18n.T(r.Context(), "login.something_went_wrong"), "").Render(r.Context(), w)
		return
	}

	templates.Account(user, "", i18n.T(r.Context(), "account.password_changed")).Render(r.Context(), w)
}
