package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

// staticFS is a stand-in for the embedded asset tree.
var staticFS = fstest.MapFS{"css/style.css": &fstest.MapFile{Data: []byte("body{}")}}

// TestStaticServesReadMethodsOnly guards the asset route: http.FileServer
// ignores the request method, so the route has to bind GET and HEAD itself
// or a POST/DELETE/TRACE would be answered with the file body.
func TestStaticServesReadMethodsOnly(t *testing.T) {
	tests := []struct {
		method   string
		wantCode int
		wantBody string
	}{
		{http.MethodGet, http.StatusOK, "body{}"},
		{http.MethodHead, http.StatusOK, ""},
		{http.MethodPost, http.StatusMethodNotAllowed, ""},
		{http.MethodDelete, http.StatusMethodNotAllowed, ""},
		{http.MethodPut, http.StatusMethodNotAllowed, ""},
		{http.MethodTrace, http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/static/css/style.css", nil)
			testHandlers().Router(staticFS).ServeHTTP(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
			if got := w.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

// TestTrailingSlashesReachTheSameRoute checks that a stray or doubled
// slash still lands inside the mounted resource router rather than falling
// through to the catch-all. Unauthenticated, the tell is the login
// redirect: the catch-all would send the user to /tasks instead.
func TestTrailingSlashesReachTheSameRoute(t *testing.T) {
	paths := []string{
		"/clients", "/clients/", "/clients//",
		"/clients/1/edit", "/clients/1/edit/",
		"/tasks/", "/time/report/", "/account/",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			testHandlers().Router(staticFS).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

			if w.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
			}
			if loc := w.Header().Get("Location"); loc == "/tasks" {
				t.Errorf("%s fell through to the catch-all instead of the mounted route", path)
			}
		})
	}
}

// TestRootIsNotStripped makes sure canonicalising slashes leaves "/"
// itself alone, so the landing redirect still works.
func TestRootIsNotStripped(t *testing.T) {
	w := httptest.NewRecorder()
	testHandlers().Router(staticFS).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/tasks" {
		t.Errorf("got %d -> %q, want 303 -> /tasks", w.Code, w.Header().Get("Location"))
	}
}
