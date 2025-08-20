package tests

import (
	"github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/schema"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPWithExternalFunction(t *testing.T) {
	kurofPath := getSamplePath("toUpper_test.kurof")

	// 2. Crear definición del mock
	def := &schema.MockDefinition{
		Protocol: "http",
		Port:     8088,
		Import:   []string{kurofPath},
		Context: &schema.Context{
			Variables: map[string]any{
				"nombre": "gatito",
			},
		},
		Routes: []schema.Route{
			{
				Path:   "/hola",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Body:   `{{ template "toUpper" .context.nombre }}`,
				},
			},
		},
	}

	handler := runtime.NewHTTPHandler()
	go handler.Start(def)
	time.Sleep(200 * time.Millisecond) // esperar a que el server esté listo

	resp, err := http.Get("http://localhost:8088/hola")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "GATITO", string(body))
}
