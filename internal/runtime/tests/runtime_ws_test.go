package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/usekuro/usekuro/internal/runtime"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/usekuro/usekuro/internal/schema"
)

func TestWSWithExternalFunction(t *testing.T) {
	// Crear archivo temporal con función toUpper
	kurofPath := getSamplePath("toUpper_test.kurof")

	// Definición del mock
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     9201,
		Import:   []string{kurofPath},
		Context: &schema.Context{
			Variables: map[string]any{},
		},
		OnMessage: &schema.OnMessage{
			Match: `(?P<cmd>\w+)`,
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ if eq .input.cmd "ping" }}true{{ else }}false{{ end }}`,
					Respond: "pong",
				},
			},
			Else: "unknown command",
		},
	}

	handler := runtime.NewWSHandler()
	assert.NoError(t, handler.Start(def))

	// Esperamos que el servidor se levante
	time.Sleep(300 * time.Millisecond)

	wsURL := "ws://localhost:9201/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	assert.NoError(t, err)

	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(msg))
}
