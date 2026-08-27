// Package migrations embeds the versioned SQL migration files so they ship
// inside the binary. It lives in migrations/ because //go:embed only reaches
// files at or below the embedding source file's directory.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
