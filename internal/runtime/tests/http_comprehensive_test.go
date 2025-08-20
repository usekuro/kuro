package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/schema"
)

func TestHTTPBasicRoutes(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "http",
		Port:     8090,
		Meta: schema.Meta{
			Name:        "Test API",
			Description: "API de prueba para productos",
		},
		Routes: []schema.Route{
			{
				Path:   "/health",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: `{"status": "ok", "timestamp": "{{ now }}"}`,
				},
			},
			{
				Path:   "/products",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: `[
						{"id": 1, "name": "Laptop", "price": 999.99},
						{"id": 2, "name": "Mouse", "price": 29.99}
					]`,
				},
			},
			{
				Path:   "/products",
				Method: "POST",
				Response: schema.ResponseDefinition{
					Status: 201,
					Headers: map[string]string{
						"Content-Type": "application/json",
						"Location":     "/products/3",
					},
					Body: `{"id": 3, "name": "{{ .input.name }}", "price": {{ .input.price }}}`,
				},
			},
		},
	}

	handler := runtime.NewHTTPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test GET /health
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8090/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var body map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "ok", body["status"])
		assert.NotEmpty(t, body["timestamp"])
	})

	// Test GET /products
	t.Run("Get Products", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8090/products")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var products []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&products)
		require.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, "Laptop", products[0]["name"])
	})

	// Test POST /products
	t.Run("Create Product", func(t *testing.T) {
		newProduct := map[string]interface{}{
			"name":  "Keyboard",
			"price": 79.99,
		}
		jsonData, _ := json.Marshal(newProduct)

		resp, err := http.Post("http://localhost:8090/products", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 201, resp.StatusCode)
		assert.Equal(t, "/products/3", resp.Header.Get("Location"))

		var created map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&created)
		require.NoError(t, err)
		assert.Equal(t, float64(3), created["id"])
		assert.Equal(t, "Keyboard", created["name"])
		assert.Equal(t, 79.99, created["price"])
	})
}

func TestHTTPWithTemplates(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "http",
		Port:     8091,
		Context: &schema.Context{
			Variables: map[string]any{
				"users": []map[string]interface{}{
					{"id": 1, "name": "Alice", "role": "admin"},
					{"id": 2, "name": "Bob", "role": "user"},
					{"id": 3, "name": "Charlie", "role": "user"},
				},
			},
		},
		Routes: []schema.Route{
			{
				Path:   "/users",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Headers: map[string]string{
						"Content-Type":  "application/json",
						"X-Total-Count": "{{ len .context.users }}",
					},
					Body: `[
{{- $first := true -}}
{{- range $u := .context.users -}}
    {{- if eq (index $u "role") "user" -}}
        {{- if not $first }},{{ end -}}
        {"id": {{ index $u "id" }}, "name": "{{ index $u "name" }}", "role": "{{ index $u "role" }}"}
        {{- $first = false -}}
    {{- end -}}
{{- end -}}
]
`,
				},
			},
			{
				Path:   "/users/admin",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Body: `[
{{ range .context.users }}{{ if eq .role "admin" }}
  {"id": {{ .id }}, "name": "{{ .name }}"}
{{ end }}{{ end }}
]`,
				},
			},
		},
	}

	handler := runtime.NewHTTPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test template rendering with context
	t.Run("Users List with Templates", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8091/users")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "3", resp.Header.Get("X-Total-Count"))

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var users []map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &users))

		assert.Len(t, users, 2)
		assert.Equal(t, "Bob", users[0]["name"])
	})

	// Test conditional template
	t.Run("Admin Users Filter", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8091/users/admin")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Alice")
		assert.NotContains(t, string(body), "Bob")
	})
}

func TestHTTPDynamicHeaders(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "http",
		Port:     8092,
		Session: &schema.Session{
			Timeout: "5s",
		},
		Routes: []schema.Route{
			{
				Path:   "/api/data",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Headers: map[string]string{
						"X-Request-ID":    "{{ uuid }}",
						"X-Timestamp":     "{{ now }}",
						"X-Debug-Enabled": "{{ if .context.debug }}true{{ else }}false{{ end }}",
						"Cache-Control":   "max-age=3600",
					},
					Body: `{"message": "Dynamic headers test"}`,
				},
			},
		},
	}

	handler := runtime.NewHTTPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://localhost:8092/api/data")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
	assert.NotEmpty(t, resp.Header.Get("X-Timestamp"))
	assert.Equal(t, "false", resp.Header.Get("X-Debug-Enabled"))
	assert.Equal(t, "max-age=3600", resp.Header.Get("Cache-Control"))
}

func TestHTTPErrorResponses(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "http",
		Port:     8093,
		Routes: []schema.Route{
			{
				Path:   "/api/protected",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 401,
					Headers: map[string]string{
						"WWW-Authenticate": "Bearer realm=\"api\"",
					},
					Body: `{"error": "Unauthorized", "message": "Token required"}`,
				},
			},
			{
				Path:   "/api/notfound",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 404,
					Body:   `{"error": "Not Found", "path": "{{ .input.path }}"}`,
				},
			},
			{
				Path:   "/api/error",
				Method: "POST",
				Response: schema.ResponseDefinition{
					Status: 500,
					Body:   `{"error": "Internal Server Error", "timestamp": "{{ now }}"}`,
				},
			},
		},
	}

	handler := runtime.NewHTTPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test 401 Unauthorized
	t.Run("Unauthorized Response", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8093/api/protected")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode)
		assert.Equal(t, "Bearer realm=\"api\"", resp.Header.Get("WWW-Authenticate"))
	})

	// Test 404 Not Found
	t.Run("Not Found Response", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8093/api/notfound")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	// Test 500 Internal Server Error
	t.Run("Server Error Response", func(t *testing.T) {
		resp, err := http.Post("http://localhost:8093/api/error", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 500, resp.StatusCode)
	})
}
