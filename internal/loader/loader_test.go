package loader

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLoadMockFromFile(t *testing.T) {
	content := `
protocol: http
port: 8081
routes:
  - path: /ping
    method: GET
    response:
      status: 200
      body: "pong"
`
	tmp := "test_ping.kuro"
	err := os.WriteFile(tmp, []byte(content), 0644)
	require.NoError(t, err)
	defer os.Remove(tmp)

	def, err := LoadMockFromFile(tmp)
	require.NoError(t, err)
	require.Equal(t, "http", def.Protocol)
	require.Equal(t, 8081, def.Port)
	require.Len(t, def.Routes, 1)
	require.Equal(t, "/ping", def.Routes[0].Path)
}
