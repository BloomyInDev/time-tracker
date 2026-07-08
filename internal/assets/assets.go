// Package assets embeds the static frontend files (CSS/JS) into the
// binary so the server doesn't depend on a "static" directory being
// present on disk at runtime.
package assets

import "embed"

//go:embed static
var Static embed.FS
