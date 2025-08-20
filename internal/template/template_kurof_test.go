package template_test

import (
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/template"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtensionTemplateExecution(t *testing.T) {
	// Simulamos carga de una función externa (como si viniera de .kurof)
	kurofCode := `{{ define "shout" }}{{ .message | upper }}{{ end }}`
	reg := extensions.NewRegistry()
	reg.Register("shout", kurofCode, "test.kurof")

	ctx := template.MergeContext(map[string]any{
		"message": "hola mundo",
	}, nil, nil)

	r, err := template.NewRuntime(ctx, reg)
	require.NoError(t, err)

	out, err := r.Render("test", `{{ template "shout" .input }}`)
	require.NoError(t, err)
	require.Equal(t, "HOLA MUNDO", out)
}

func TestDuplicateFunctionConflict(t *testing.T) {
	reg := extensions.NewRegistry()
	reg.Register("same", `{{ define "same" }}a{{ end }}`, "a.kurof")
	reg.Register("same", `{{ define "same" }}b{{ end }}`, "b.kurof") // sobrescribe

	ext, ok := reg.Get("same")
	require.True(t, ok)
	require.Contains(t, ext.Content, "b") // validamos sobrescritura
}

func TestLoadLocalKurof(t *testing.T) {
	// archivo temporal simulado
	path := "test_func.kurof"
	code := `{{ define "ok" }}✔️{{ end }}`
	err := os.WriteFile(path, []byte(code), 0644)
	require.NoError(t, err)
	defer os.Remove(path)

	content, err := extensions.LoadKurof(path)
	require.NoError(t, err)
	require.Contains(t, content, "define")
}
