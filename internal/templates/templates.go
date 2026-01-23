// Package templates provides embedded script templates for tap new.
package templates

import (
	"embed"
)

//go:embed bash.tmpl python.tmpl
var FS embed.FS
