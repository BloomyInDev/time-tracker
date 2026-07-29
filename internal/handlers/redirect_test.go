package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appi18n "github.com/bloomyindev/time-tracker/internal/i18n"
	"github.com/bloomyindev/time-tracker/internal/redirect"
	"github.com/invopop/ctxi18n"
)

func TestMain(m *testing.M) {
	if err := appi18n.Load(); err != nil {
		panic(err)
	}
	m.Run()
}

// localeReq builds a GET request carrying the context the templates expect:
// a locale and, like the live server, the current URI.
func localeReq(target string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, target, nil)
	ctx, err := ctxi18n.WithLocale(r.Context(), "en")
	if err != nil {
		panic(err)
	}
	r = r.WithContext(ctx)

	var out *http.Request
	redirect.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		out = req
	})).ServeHTTP(httptest.NewRecorder(), r)
	return out
}

func TestLoginPageCarriesRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	Login(w, localeReq("/login?redirect=%2Ftasks%3Fclient%3D3"))

	if body := w.Body.String(); !strings.Contains(body, `name="redirect" value="/tasks?client=3"`) {
		t.Errorf("hidden redirect field missing, body:\n%s", body)
	}
}

func TestLoginPageDropsUnsafeRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	Login(w, localeReq("/login?redirect=https%3A%2F%2Fevil.example"))

	if strings.Contains(w.Body.String(), `name="redirect"`) {
		t.Error("unsafe redirect kept as the login form destination")
	}
}

func TestLocaleLinksReturnToCurrentPage(t *testing.T) {
	w := httptest.NewRecorder()
	Login(w, localeReq("/login?redirect=%2Ftasks"))

	body := w.Body.String()
	for _, want := range []string{
		`href="/lang/en?redirect=%2Flogin%3Fredirect%3D%252Ftasks"`,
		`href="/lang/fr?redirect=%2Flogin%3Fredirect%3D%252Ftasks"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("locale link %s missing, body:\n%s", want, body)
		}
	}
}

func TestSetLocaleRedirect(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"returns to page", "?redirect=%2Ftime%3Fmonth%3D2026-07", "/time?month=2026-07"},
		{"rejects off-site", "?redirect=%2F%2Fevil.example", "/"},
		{"defaults to root", "", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/lang/fr"+tt.query, nil)
			r.SetPathValue("code", "fr")
			w := httptest.NewRecorder()
			SetLocale(w, r)

			if w.Code != http.StatusSeeOther {
				t.Errorf("status = %d, want %d", w.Code, http.StatusSeeOther)
			}
			if got := w.Header().Get("Location"); got != tt.want {
				t.Errorf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}
