package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/template"
	"testing"
)

func TestUUIDAndJSON(t *testing.T) {
	ctx := template.MergeContext(map[string]any{
		"user": "kurocat",
	}, nil, nil)
	r, _ := template.NewRuntime(ctx, extensions.NewRegistry())
	out, err := r.Render("json", `{"id": "{{ uuid }}", "user": "{{ .input.user }}"}`)
	assert.NoError(t, err)
	assert.Contains(t, out, "kurocat")
	assert.Contains(t, out, "id")
}
