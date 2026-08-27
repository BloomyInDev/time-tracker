package handlers

import (
	"io/fs"
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/i18n"
	"github.com/bloomyindev/time-tracker/internal/redirect"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router assembles the whole site. It is the only place that knows the
// URL prefixes; a resource file never repeats its own mount path.
func (h *Handlers) Router(static fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	// Canonicalise the path before routing, so "/clients/" and
	// "/clients//" reach the same handler as "/clients" instead of
	// falling through to the catch-all.
	r.Use(middleware.CleanPath, middleware.StripSlashes)
	r.Use(i18n.Middleware, redirect.Middleware)

	// Assets are read-only: bind GET and HEAD explicitly, because
	// http.FileServer ignores the method and would otherwise serve a
	// file body for POST, DELETE or TRACE too.
	assets := http.StripPrefix("/static/", http.FileServerFS(static))
	r.Method(http.MethodGet, "/static/*", assets)
	r.Method(http.MethodHead, "/static/*", assets)

	// Public pages: the landing redirect, the login flow and the locale
	// switch. Unknown paths land on the app rather than a bare 404.
	r.Get("/", h.home)
	r.NotFound(h.home)
	r.Get("/login", h.loginForm)
	r.Post("/login", h.loginSubmit)
	r.Get("/logout", h.logout)
	r.Get("/lang/{code}", h.setLocale)

	// Everything below needs a session.
	r.Group(func(r chi.Router) {
		r.Use(h.Auth.RequireAuth)
		r.Mount("/clients", h.ClientsRouter())
		r.Mount("/task-types", h.TaskTypesRouter())
		r.Mount("/periods", h.PeriodsRouter())
		r.Mount("/tasks", h.TasksRouter())
		r.Mount("/time", h.TimeRouter())
		r.Mount("/account", h.AccountRouter())
	})

	return r
}
