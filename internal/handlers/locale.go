package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/bloomyindev/time-tracker/internal/i18n"
)

func SetLocale(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

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

	dest := "/"
	if ref, err := url.Parse(r.Header.Get("Referer")); err == nil && ref.Host == r.Host {
		dest = ref.RequestURI()
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
