package internal

import "embed"

// TemplateFS is the embedded filesystem for templates
//
//go:embed templates/*.html templates/*.txt
var TemplateFS embed.FS
