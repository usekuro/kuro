package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/template"
	"testing"
)

func TestContextTemplate(t *testing.T) {
	ctx := template.MergeContext(nil, nil, map[string]any{
		"version": "1.0.0",
	})
	r, _ := template.NewRuntime(ctx, extensions.NewRegistry())
	out, err := r.Render("body", `API v{{ .context.version }}`)
	assert.NoError(t, err)
	assert.Equal(t, "API v1.0.0", out)
}
