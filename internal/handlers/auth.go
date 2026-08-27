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

	h.setSessionCookie(w, token, int((24 * time.Hour).Seconds()))
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// setSessionCookie writes the session cookie. Login and logout both go through
// it so their attributes can't drift: a browser replaces a cookie only when the
// name, path and domain all match, so a clear that disagrees with the set
// leaves the original in place.
func (h *Handlers) setSessionCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.Config.SecureCookies,
		MaxAge:   maxAge,
	})
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		h.Auth.Logout(cookie.Value)
	}

	h.setSessionCookie(w, "", -1)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
