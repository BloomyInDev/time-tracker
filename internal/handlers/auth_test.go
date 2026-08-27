package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
)

// loggedIn builds a router backed by a real database holding one user, and
// returns it alongside that user's session token.
func loggedIn(t *testing.T, cfg config.Config) (http.Handler, string) {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"), db.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	svc := auth.NewService(conn, "test-secret")
	if err := svc.Register("a@b.c", "hunter2"); err != nil {
		t.Fatal(err)
	}
	token, err := svc.Login("a@b.c", "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	return New(conn, svc, cfg).Router(fstest.MapFS{}), token
}

// findCookie returns the named cookie from the response, failing if absent.
func findCookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %q cookie in the response", name)
	return nil
}

// TestSessionCookieSecureFollowsConfig covers both directions. The default is
// on, for a deployment behind a TLS-terminating proxy; plain-HTTP setups have
// to turn it off, or the browser accepts the cookie and never sends it back.
func TestSessionCookieSecureFollowsConfig(t *testing.T) {
	for _, secure := range []bool{true, false} {
		t.Run(map[bool]string{true: "secure", false: "insecure"}[secure], func(t *testing.T) {
			router, _ := loggedIn(t, config.Config{SecureCookies: secure})

			body := url.Values{"email": {"a@b.c"}, "password": {"hunter2"}}
			r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != http.StatusSeeOther {
				t.Fatalf("login status = %d, want %d", w.Code, http.StatusSeeOther)
			}
			c := findCookie(t, w, auth.CookieName)
			if c.Secure != secure {
				t.Errorf("Secure = %t, want %t", c.Secure, secure)
			}
			if !c.HttpOnly {
				t.Error("HttpOnly = false, want true")
			}
			if c.SameSite != http.SameSiteLaxMode {
				t.Errorf("SameSite = %v, want Lax", c.SameSite)
			}
		})
	}
}

// TestLogoutCookieMatchesLogin is why both go through one helper: a browser
// only replaces a cookie whose name, path and domain match, so a clear that
// disagrees with the set would leave the session cookie in place.
func TestLogoutCookieMatchesLogin(t *testing.T) {
	cfg := config.Config{SecureCookies: true}
	router, token := loggedIn(t, cfg)

	r := httptest.NewRequest(http.MethodGet, "/logout", nil)
	r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	c := findCookie(t, w, auth.CookieName)
	if c.Value != "" {
		t.Errorf("value = %q, want empty", c.Value)
	}
	if c.MaxAge >= 0 {
		t.Errorf("MaxAge = %d, want negative so the browser drops it", c.MaxAge)
	}
	if !c.Secure || !c.HttpOnly || c.Path != "/" {
		t.Errorf("attributes drifted from the login cookie: %+v", c)
	}
}

// TestLocaleCookieSecureFollowsConfig keeps the language cookie consistent
// with the session one; there is no reason for them to differ.
func TestLocaleCookieSecureFollowsConfig(t *testing.T) {
	router, _ := loggedIn(t, config.Config{SecureCookies: true})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lang/fr", nil))

	if c := findCookie(t, w, "lang"); !c.Secure {
		t.Error("Secure = false, want true")
	}
}
