package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/invopop/ctxi18n/i18n"
)

func Login(w http.ResponseWriter, r *http.Request) {
	templates.Login("").Render(r.Context(), w)
}

func LoginSubmit(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		token, err := svc.Login(r.FormValue("email"), r.FormValue("password"))
		if errors.Is(err, auth.ErrInvalidCredentials) {
			templates.Login(i18n.T(r.Context(), "login.invalid_credentials")).Render(r.Context(), w)
			return
		}
		if err != nil {
			templates.Login(i18n.T(r.Context(), "login.something_went_wrong")).Render(r.Context(), w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     auth.CookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int((24 * time.Hour).Seconds()),
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func Logout(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(auth.CookieName); err == nil {
			svc.Logout(cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     auth.CookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
