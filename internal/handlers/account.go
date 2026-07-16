package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/invopop/ctxi18n/i18n"
)

func Account(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		user, err := db.GetUser(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templates.Account(user, "", "").Render(r.Context(), w)
	}
}

func UpdateDailyHours(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		user, err := db.GetUser(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var hours [7]float64
		for i := 0; i < 7; i++ {
			raw := r.FormValue("hours_" + strconv.Itoa(i))
			if raw == "" {
				continue
			}
			h, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				http.Error(w, "invalid hours", http.StatusBadRequest)
				return
			}
			hours[i] = h
		}

		if err := db.UpdateDailyHours(conn, userID, hours); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user.DailyHours = hours
		templates.Account(user, "", i18n.T(r.Context(), "account.hours_saved")).Render(r.Context(), w)
	}
}

func ChangePassword(conn *sql.DB, svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		user, err := db.GetUser(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		err = svc.ChangePassword(userID, r.FormValue("current_password"), r.FormValue("new_password"))
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
}
