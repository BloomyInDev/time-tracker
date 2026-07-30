package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/go-chi/chi/v5"
)

// Handlers holds the dependencies every handler shares. Handlers are
// methods on it, which is what lets each resource file expose its own
// sub-router.
type Handlers struct {
	DB   *sql.DB
	Auth *auth.Service
}

func New(conn *sql.DB, authSvc *auth.Service) *Handlers {
	return &Handlers{DB: conn, Auth: authSvc}
}

// userID returns the authenticated user for the request. It panics if the
// route was mounted outside the auth group: that is a wiring mistake, and
// failing loudly beats silently querying user 0.
func userID(r *http.Request) int64 {
	id, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		panic("userID called on a route without RequireAuth")
	}
	return id
}

// pathID reads the {id} path param. It writes its own 400 and reports
// false when the value isn't an integer.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// parseForm parses the request body, writing its own 400 on failure.
func parseForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return false
	}
	return true
}

// fail writes an internal error; handlers use it for anything the user
// can't act on.
func fail(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
