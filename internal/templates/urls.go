package templates

import (
	"context"

	"github.com/a-h/templ"
	"github.com/bloomyindev/time-tracker/internal/redirect"
)

// langURL builds a locale switch link that returns to the current page.
func langURL(ctx context.Context, code string) templ.SafeURL {
	return templ.SafeURL(redirect.URLWith("/lang/"+code, redirect.CurrentURI(ctx)))
}
