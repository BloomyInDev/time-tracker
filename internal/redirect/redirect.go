// Package redirect carries the "where should the user land next" value
// between the auth middleware, the login flow and the locale switcher.
// Destinations are always sanitised so a crafted link can't bounce a user
// off-site (open redirect).
package redirect

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// Param is the query string / form field holding the destination.
const Param = "redirect"

type contextKey int

const currentURIKey contextKey = iota

// Sanitize returns dest when it is a safe same-site destination: a relative
// path anchored at "/" with no scheme, no host and no protocol-relative or
// backslash trickery. Anything else falls back to fallback.
func Sanitize(dest, fallback string) string {
	if dest == "" {
		return fallback
	}
	if strings.HasPrefix(dest, "//") || strings.HasPrefix(dest, `/\`) {
		return fallback
	}

	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || !strings.HasPrefix(u.Path, "/") {
		return fallback
	}
	return u.String()
}

// URLWith appends dest to base as the redirect parameter, skipping it when
// dest isn't a safe same-site destination.
func URLWith(base, dest string) string {
	dest = Sanitize(dest, "")
	if dest == "" {
		return base
	}

	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + Param + "=" + url.QueryEscape(dest)
}

// Middleware stores the current request URI in the context so templates can
// build links that come back to the page the user is on.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), currentURIKey, r.URL.RequestURI())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CurrentURI returns the URI of the request being rendered, or "" outside a
// GET request.
func CurrentURI(ctx context.Context) string {
	uri, _ := ctx.Value(currentURIKey).(string)
	return uri
}
