package tests

import (
	"github.com/usekuro/usekuro/internal/runtime"
	"net"
	"path/filepath"
	runtime2 "runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/schema"
)

func getSamplePath(filename string) string {
	_, currentFile, _, _ := runtime2.Caller(0)
	baseDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "samples")
	return filepath.Join(baseDir, filename)
}

func TestTCPWithExternalFunction(t *testing.T) {
	// Crear archivo temporal .kurof con función toUpper
	kurofPath := getSamplePath("toUpper_test.kurof")
	// Definir mock TCP
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9101,
		Import:   []string{kurofPath},
		Context: &schema.Context{
			Variables: map[string]any{},
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<cmd>.+)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ if contains .input.cmd "ping" }}true{{ else }}false{{ end }}`,
					Respond: `{{ template "toUpper" .input.cmd }}`,
				},
			},
			Else: "NO MATCH",
		},
	}

	handler := runtime.NewTCPHandler()
	assert.NoError(t, handler.Start(def))
	defer handler.Stop()

	// Esperar que el server levante
	time.Sleep(200 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9101")
	assert.NoError(t, err)

	conn.Write([]byte("ping test"))
	resp := make([]byte, 1024)
	n, _ := conn.Read(resp)
	assert.Contains(t, string(resp[:n]), "PING TEST")

	conn.Close()
}
