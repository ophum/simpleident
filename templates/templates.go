package templates

import "embed"

//go:embed **/*.tmpl *.tmpl
var FS embed.FS
