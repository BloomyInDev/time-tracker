// Package i18n loads locale dictionaries and detects the request locale
// from a "lang" cookie override or the Accept-Language header.
package i18n

import (
	"embed"
	"net/http"
	"slices"

	"github.com/invopop/ctxi18n"
)

//go:embed locales
var localesFS embed.FS

const (
	CookieName    = "lang"
	DefaultLocale = "en"
)

var SupportedLocales = []string{"en", "fr"}

func Load() error {
	return ctxi18n.LoadWithDefault(localesFS, DefaultLocale)
}

// Middleware detects the request locale and attaches it to the context.
// Priority: "lang" cookie override, then Accept-Language header, then DefaultLocale.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := DefaultLocale
		if cookie, err := r.Cookie(CookieName); err == nil && isSupported(cookie.Value) {
			locale = cookie.Value
		} else if al := r.Header.Get("Accept-Language"); al != "" {
			locale = al
		}

		ctx, err := ctxi18n.WithLocale(r.Context(), locale)
		if err != nil {
			ctx, _ = ctxi18n.WithLocale(r.Context(), DefaultLocale)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isSupported(code string) bool {
	return slices.Contains(SupportedLocales, code)
}
