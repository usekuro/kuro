package template

import (
	"bytes"
	"text/template"

	"github.com/usekuro/usekuro/internal/extensions"
)

type Runtime struct {
	templates *template.Template
	context   map[string]any
}

// Nuevo: ahora acepta un Registry de extensiones
func NewRuntime(ctx map[string]any, registry *extensions.Registry) (*Runtime, error) {
	t := template.New("base").Funcs(FuncMap())

	// Cargar extensiones kurof si existen
	for _, ext := range registry.Extensions {
		if _, err := t.New(ext.Name).Parse(ext.Content); err != nil {
			return nil, err
		}
	}

	return &Runtime{templates: t, context: ctx}, nil
}

func (r *Runtime) Render(name, raw string) (string, error) {
	tmpl, err := r.templates.New(name).Parse(raw)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, r.context)
	return out.String(), err
}
