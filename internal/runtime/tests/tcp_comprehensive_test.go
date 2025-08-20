package tests

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/schema"
)

func TestTCPBasicCommunication(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9001,
		Meta: schema.Meta{
			Name:        "TCP Echo Server",
			Description: "Simple TCP echo server with commands",
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<cmd>\\w+)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ eq .cmd "PING" }}`,
					Respond: "PONG",
				},
				{
					If:      `{{ eq .cmd "ECHO" }}`,
					Respond: "{{ .input }}",
				},
				{
					If:      `{{ eq .cmd "TIME" }}`,
					Respond: "Current time: {{ now }}",
				},
			},
			Else: "Unknown command: {{ .cmd }}",
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test PING command
	t.Run("PING Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9001")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("PING\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "PONG\n", response)
	})

	// Test ECHO command
	t.Run("ECHO Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9001")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("ECHO\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "ECHO\n", response)
	})

	// Test TIME command
	t.Run("TIME Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9001")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("TIME\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "Current time:")
	})

	// Test unknown command
	t.Run("Unknown Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9001")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("INVALID\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "Unknown command: INVALID\n", response)
	})
}

func TestTCPWithParameters(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9002,
		OnMessage: &schema.OnMessage{
			Match: "(?P<cmd>\\w+),(?P<arg>.*)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ eq .cmd "HELLO" }}`,
					Respond: "Hi {{ .arg }}!",
				},
				{
					If:      `{{ eq .cmd "CALC" }}`,
					Respond: "Result: {{ .arg }}",
				},
				{
					If:      `{{ eq .cmd "UPPER" }}`,
					Respond: "{{ upper .arg }}",
				},
				{
					If:      `{{ eq .cmd "LOWER" }}`,
					Respond: "{{ lower .arg }}",
				},
			},
			Else: "ERROR: Invalid format",
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test HELLO with parameter
	t.Run("HELLO with Name", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9002")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("HELLO,World\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "Hi World!\n", response)
	})

	// Test UPPER command
	t.Run("UPPER Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9002")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("UPPER,hello world\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "HELLO WORLD\n", response)
	})

	// Test LOWER command
	t.Run("LOWER Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9002")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("LOWER,HELLO WORLD\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "hello world\n", response)
	})
}

func TestTCPWithContext(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9003,
		Context: &schema.Context{
			Variables: map[string]any{
				"version":  "1.0.0",
				"server":   "UseKuro TCP Server",
				"commands": []string{"STATUS", "INFO", "HELP"},
			},
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<cmd>\\w+)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ eq .cmd "STATUS" }}`,
					Respond: "Server: {{ .context.server }} | Version: {{ .context.version }} | Status: OK",
				},
				{
					If:      `{{ eq .cmd "INFO" }}`,
					Respond: "{{ .context.server }} v{{ .context.version }}",
				},
				{
					If: `{{ eq .cmd "HELP" }}`,
					Respond: `Available commands: {{ range .context.commands }}{{ . }} {{ end }}`,
				},
			},
			Else: "Command not recognized. Type HELP for available commands.",
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test STATUS command
	t.Run("STATUS Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9003")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("STATUS\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "UseKuro TCP Server")
		assert.Contains(t, response, "1.0.0")
		assert.Contains(t, response, "Status: OK")
	})

	// Test HELP command
	t.Run("HELP Command", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9003")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("HELP\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "STATUS")
		assert.Contains(t, response, "INFO")
		assert.Contains(t, response, "HELP")
	})
}

func TestTCPMultipleConnections(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9004,
		Session: &schema.Session{
			Timeout: "5s",
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<msg>.*)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ contains .msg "hello" }}`,
					Respond: "Hello from server!",
				},
			},
			Else: "Message received: {{ .msg }}",
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test multiple concurrent connections
	t.Run("Multiple Concurrent Connections", func(t *testing.T) {
		connections := make([]net.Conn, 3)
		
		// Create multiple connections
		for i := 0; i < 3; i++ {
			conn, err := net.Dial("tcp", "localhost:9004")
			require.NoError(t, err)
			connections[i] = conn
			defer conn.Close()
		}

		// Send messages from each connection
		for i, conn := range connections {
			msg := fmt.Sprintf("hello from client %d\n", i+1)
			_, err := conn.Write([]byte(msg))
			require.NoError(t, err)
		}

		// Read responses
		for _, conn := range connections {
			reader := bufio.NewReader(conn)
			response, err := reader.ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, "Hello from server!\n", response)
		}
	})
}

func TestTCPComplexPatterns(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9005,
		OnMessage: &schema.OnMessage{
			Match: "(?P<action>\\w+)\\s+(?P<resource>\\w+)(?:\\s+(?P<params>.*))?",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ and (eq .action "GET") (eq .resource "user") }}`,
					Respond: `{"user": {"id": 1, "name": "Alice"}}`,
				},
				{
					If:      `{{ and (eq .action "GET") (eq .resource "product") }}`,
					Respond: `{"product": {"id": 101, "name": "Laptop", "price": 999.99}}`,
				},
				{
					If:      `{{ and (eq .action "SET") (eq .resource "config") }}`,
					Respond: `Configuration updated: {{ .params }}`,
				},
				{
					If:      `{{ eq .action "LIST" }}`,
					Respond: `Listing all {{ .resource }}s`,
				},
			},
			Else: `Invalid command: {{ .action }} {{ .resource }}`,
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test GET user
	t.Run("GET user", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9005")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("GET user\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "Alice")
	})

	// Test SET config
	t.Run("SET config", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9005")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("SET config debug=true\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "Configuration updated: debug=true")
	})

	// Test LIST command
	t.Run("LIST products", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9005")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("LIST product\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "Listing all products\n", response)
	})
}

func TestTCPBinaryProtocol(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "tcp",
		Port:     9006,
		OnMessage: &schema.OnMessage{
			Match: "(?P<header>\\w{2})(?P<payload>.*)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ eq .header "HB" }}`,
					Respond: "HEARTBEAT_ACK",
				},
				{
					If:      `{{ eq .header "DT" }}`,
					Respond: "DATA_RECEIVED: {{ len .payload }} bytes",
				},
				{
					If:      `{{ eq .header "ST" }}`,
					Respond: "STATUS: ONLINE",
				},
			},
			Else: "UNKNOWN_HEADER: {{ .header }}",
		},
	}

	handler := runtime.NewTCPHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test heartbeat
	t.Run("Heartbeat Protocol", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9006")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("HB\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Equal(t, "HEARTBEAT_ACK\n", response)
	})

	// Test data packet
	t.Run("Data Packet", func(t *testing.T) {
		conn, err := net.Dial("tcp", "localhost:9006")
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.Write([]byte("DTHello World\n"))
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		require.NoError(t, err)

		assert.Contains(t, response, "DATA_RECEIVED")
		assert.Contains(t, response, "11 bytes")
	})
}
