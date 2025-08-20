package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/template"
	"testing"
)

func TestHTTPBodyRendering(t *testing.T) {
	ctx := template.MergeContext(nil, nil, map[string]any{
		"env": "testing",
	})
	r, _ := template.NewRuntime(ctx, extensions.NewRegistry())
	out, err := r.Render("body", `{"msg": "hello from {{ .context.env }}"}`)
	assert.NoError(t, err)
	assert.Contains(t, out, "testing")
}
