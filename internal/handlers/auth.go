package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/bloomyindev/time-tracker/internal/redirect"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
	"github.com/invopop/ctxi18n/i18n"
)

func (h *Handlers) loginForm(w http.ResponseWriter, r *http.Request) {
	dest := redirect.Sanitize(r.URL.Query().Get(redirect.Param), "")
	templates.Login("", dest).Render(r.Context(), w)
}

func (h *Handlers) loginSubmit(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}

	dest := redirect.Sanitize(r.FormValue(redirect.Param), "/")

	token, err := h.Auth.Login(r.FormValue("email"), r.FormValue("password"))
	if errors.Is(err, auth.ErrInvalidCredentials) {
		templates.Login(i18n.T(r.Context(), "login.invalid_credentials"), dest).Render(r.Context(), w)
		return
	}
	if err != nil {
		templates.Login(i18n.T(r.Context(), "login.something_went_wrong"), dest).Render(r.Context(), w)
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
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		h.Auth.Logout(cookie.Value)
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
