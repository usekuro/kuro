package tests

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/schema"
)

func TestWebSocketBasicCommunication(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     8095,
		Meta: schema.Meta{
			Name:        "WebSocket Chat Server",
			Description: "Interactive WebSocket server for chat",
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<type>\\w+):(?P<content>.*)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ eq .type "chat" }}`,
					Respond: `{"type": "message", "content": "{{ .content }}", "timestamp": "{{ now }}"}`,
				},
				{
					If:      `{{ eq .type "ping" }}`,
					Respond: `{"type": "pong", "timestamp": "{{ now }}"}`,
				},
				{
					If:      `{{ eq .type "status" }}`,
					Respond: `{"type": "status", "online": true, "users": 1}`,
				},
			},
			Else: `{"type": "error", "message": "Unknown message type: {{ .type }}"}`,
		},
	}

	handler := runtime.NewWSHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test chat message
	t.Run("Chat Message", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8095", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send chat message
		err = conn.WriteMessage(websocket.TextMessage, []byte("chat:Hello World"))
		require.NoError(t, err)

		// Read response
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "message", response["type"])
		assert.Equal(t, "Hello World", response["content"])
		assert.NotEmpty(t, response["timestamp"])
	})

	// Test ping/pong
	t.Run("Ping Pong", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8095", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage, []byte("ping:"))
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "pong", response["type"])
	})

	// Test status
	t.Run("Status Request", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8095", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage, []byte("status:"))
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "status", response["type"])
		assert.Equal(t, true, response["online"])
		assert.Equal(t, float64(1), response["users"])
	})
}

func TestWebSocketJSONProtocol(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     8096,
		Context: &schema.Context{
			Variables: map[string]any{
				"rooms": []string{"general", "tech", "random"},
				"users": map[string]interface{}{
					"alice": map[string]interface{}{"role": "admin", "status": "online"},
					"bob":   map[string]interface{}{"role": "user", "status": "away"},
				},
			},
		},
		OnMessage: &schema.OnMessage{
			Match: ".*",
			Conditions: []schema.OnMessageRule{
				{
					If: `{{ contains .input "\"action\":\"join\"" }}`,
					Respond: `{
						"event": "joined",
						"room": "general",
						"available_rooms": {{ toJson .context.rooms }},
						"timestamp": "{{ now }}"
					}`,
				},
				{
					If: `{{ contains .input "\"action\":\"list_users\"" }}`,
					Respond: `{
						"event": "users_list",
						"users": {{ toJson .context.users }},
						"total": {{ len .context.users }}
					}`,
				},
				{
					If: `{{ contains .input "\"action\":\"send_message\"" }}`,
					Respond: `{
						"event": "message_sent",
						"id": "{{ uuid }}",
						"timestamp": "{{ now }}",
						"status": "delivered"
					}`,
				},
			},
			Else: `{"event": "error", "message": "Invalid action"}`,
		},
	}

	handler := runtime.NewWSHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test join action
	t.Run("Join Room", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8096", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		joinMsg := map[string]string{"action": "join", "user": "charlie"}
		msg, _ := json.Marshal(joinMsg)
		err = conn.WriteMessage(websocket.TextMessage, msg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "joined", response["event"])
		assert.Equal(t, "general", response["room"])
		
		rooms := response["available_rooms"].([]interface{})
		assert.Len(t, rooms, 3)
		assert.Contains(t, rooms, "general")
		assert.Contains(t, rooms, "tech")
	})

	// Test list users
	t.Run("List Users", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8096", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		listMsg := map[string]string{"action": "list_users"}
		msg, _ := json.Marshal(listMsg)
		err = conn.WriteMessage(websocket.TextMessage, msg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "users_list", response["event"])
		assert.Equal(t, float64(2), response["total"])
		
		users := response["users"].(map[string]interface{})
		assert.Contains(t, users, "alice")
		assert.Contains(t, users, "bob")
	})

	// Test send message
	t.Run("Send Message", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8096", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		sendMsg := map[string]string{
			"action":  "send_message",
			"content": "Hello everyone!",
			"room":    "general",
		}
		msg, _ := json.Marshal(sendMsg)
		err = conn.WriteMessage(websocket.TextMessage, msg)
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "message_sent", response["event"])
		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "delivered", response["status"])
	})
}

func TestWebSocketBroadcast(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     8097,
		OnMessage: &schema.OnMessage{
			Match: "(?P<cmd>\\w+)(?::(?P<data>.*))?",
			Conditions: []schema.OnMessageRule{
				{
					If: `{{ eq .cmd "broadcast" }}`,
					Respond: `{
						"type": "broadcast",
						"message": "{{ .data }}",
						"from": "server",
						"timestamp": "{{ now }}"
					}`,
				},
				{
					If: `{{ eq .cmd "echo" }}`,
					Respond: `{
						"type": "echo",
						"original": "{{ .data }}",
						"reversed": "{{ reverse .data }}"
					}`,
				},
			},
			Else: `{"type": "unknown", "command": "{{ .cmd }}"}`,
		},
	}

	handler := runtime.NewWSHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test broadcast
	t.Run("Broadcast Message", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8097", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage, []byte("broadcast:Important announcement"))
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "broadcast", response["type"])
		assert.Equal(t, "Important announcement", response["message"])
		assert.Equal(t, "server", response["from"])
	})
}

func TestWebSocketRealtimeEvents(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     8098,
		Context: &schema.Context{
			Variables: map[string]any{
				"events": []string{"user_joined", "user_left", "message", "typing"},
			},
		},
		OnMessage: &schema.OnMessage{
			Match: ".*",
			Conditions: []schema.OnMessageRule{
				{
					If: `{{ contains .input "subscribe" }}`,
					Respond: `{
						"type": "subscribed",
						"events": {{ toJson .context.events }},
						"subscription_id": "{{ uuid }}"
					}`,
				},
				{
					If: `{{ contains .input "typing" }}`,
					Respond: `{
						"type": "typing_indicator",
						"user": "anonymous",
						"is_typing": true,
						"timestamp": "{{ now }}"
					}`,
				},
				{
					If: `{{ contains .input "stop_typing" }}`,
					Respond: `{
						"type": "typing_indicator",
						"user": "anonymous",
						"is_typing": false,
						"timestamp": "{{ now }}"
					}`,
				},
			},
			Else: `{"type": "event", "name": "generic", "data": {}}`,
		},
	}

	handler := runtime.NewWSHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test event subscription
	t.Run("Subscribe to Events", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8098", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"action": "subscribe"}`))
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "subscribed", response["type"])
		events := response["events"].([]interface{})
		assert.Len(t, events, 4)
		assert.NotEmpty(t, response["subscription_id"])
	})

	// Test typing indicator
	t.Run("Typing Indicator", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8098", Path: "/"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		// Start typing
		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"action": "typing"}`))
		require.NoError(t, err)

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "typing_indicator", response["type"])
		assert.Equal(t, true, response["is_typing"])

		// Stop typing
		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"action": "stop_typing"}`))
		require.NoError(t, err)

		_, message, err = conn.ReadMessage()
		require.NoError(t, err)

		err = json.Unmarshal(message, &response)
		require.NoError(t, err)

		assert.Equal(t, "typing_indicator", response["type"])
		assert.Equal(t, false, response["is_typing"])
	})
}

func TestWebSocketMultipleClients(t *testing.T) {
	def := &schema.MockDefinition{
		Protocol: "ws",
		Port:     8099,
		Session: &schema.Session{
			Timeout: "10s",
		},
		OnMessage: &schema.OnMessage{
			Match: "(?P<user>\\w+):(?P<msg>.*)",
			Conditions: []schema.OnMessageRule{
				{
					If:      `{{ .user }}`,
					Respond: `{"from": "{{ .user }}", "message": "{{ .msg }}", "echo": true}`,
				},
			},
			Else: `{"error": "Invalid format"}`,
		},
	}

	handler := runtime.NewWSHandler()
	go handler.Start(def)
	defer handler.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test multiple simultaneous connections
	t.Run("Multiple Clients", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8099", Path: "/"}
		
		// Create multiple connections
		conn1, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn1.Close()

		conn2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn2.Close()

		// Send from client 1
		err = conn1.WriteMessage(websocket.TextMessage, []byte("alice:Hello from Alice"))
		require.NoError(t, err)

		// Send from client 2
		err = conn2.WriteMessage(websocket.TextMessage, []byte("bob:Hello from Bob"))
		require.NoError(t, err)

		// Read response for client 1
		_, message1, err := conn1.ReadMessage()
		require.NoError(t, err)
		var response1 map[string]interface{}
		json.Unmarshal(message1, &response1)
		assert.Equal(t, "alice", response1["from"])

		// Read response for client 2
		_, message2, err := conn2.ReadMessage()
		require.NoError(t, err)
		var response2 map[string]interface{}
		json.Unmarshal(message2, &response2)
		assert.Equal(t, "bob", response2["from"])
	})
}
