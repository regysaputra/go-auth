package internal

import "embed"

//go:embed templates/*.html templates/*.txt
var TemplateFS embed.FS
