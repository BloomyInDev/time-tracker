package auth

import (
	"context"
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/redirect"
)

const CookieName = "session"

func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, loginURL(r), http.StatusSeeOther)
			return
		}

		userID, ok := s.sessions.Get(cookie.Value)
		if !ok {
			http.Redirect(w, r, loginURL(r), http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loginURL points at the login page, carrying the page the user was denied so
// they land back on it once signed in. Only GET requests are worth resuming.
func loginURL(r *http.Request) string {
	if r.Method != http.MethodGet {
		return "/login"
	}
	return redirect.URLWith("/login", r.URL.RequestURI())
}
