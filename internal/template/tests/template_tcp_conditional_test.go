package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/template"
	"testing"
)

func TestTCPConditional(t *testing.T) {
	ctx := template.MergeContext(map[string]any{
		"cmd": "PING",
	}, nil, nil)
	r, _ := template.NewRuntime(ctx, extensions.NewRegistry())
	out, err := r.Render("response", `{{ if eq .input.cmd "PING" }}PONG{{ else }}UNKNOWN{{ end }}`)
	assert.NoError(t, err)
	assert.Equal(t, "PONG", out)
}
