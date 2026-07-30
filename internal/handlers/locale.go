package handlers

import (
	"net/http"
	"time"

	"github.com/bloomyindev/time-tracker/internal/i18n"
	"github.com/bloomyindev/time-tracker/internal/redirect"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) setLocale(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	valid := false
	for _, l := range i18n.SupportedLocales {
		if l == code {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "unsupported locale", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     i18n.CookieName,
		Value:    code,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	dest := redirect.Sanitize(r.URL.Query().Get(redirect.Param), "/")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
