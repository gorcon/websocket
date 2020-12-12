package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

const MockPassword = "password"

const MockCommandStatusResponseText = `status
hostname: Rust Server [DOCKER]
version : 2260 secure (secure mode enabled, connected to Steam3)
map     : Procedural Map
players : 0 (500 max) (0 queued) (0 joining)

id name ping connected addr owner violation kicks
`

func handlers() http.Handler {
	server := http.NewServeMux()

	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	server.HandleFunc("/"+MockPassword, func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v\n", err)
			return
		}

		defer ws.Close()

		var response Message

		// Receive message.
		_, p, err := ws.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure): unexpected EO") {
				log.Printf("read message error: %v\n", err)
			}
			return
		}

		var message Message
		if err := json.Unmarshal(p, &message); err != nil {
			// TODO: What Rust responses on read message fail?
			fmt.Println(string(p))
			log.Printf("unmarshal message error: %v\n", err)
			return
		}

		switch message.Message {
		case "status":
			response = Message{
				Message:    MockCommandStatusResponseText,
				Identifier: message.Identifier,
				Type:       "Generic",
			}
		case "deadline":
			time.Sleep(DefaultDeadline + 1*time.Second)
			response = Message{
				Message:    fmt.Sprintf("sleep for %d secends", DefaultDeadline+1*time.Second),
				Identifier: message.Identifier,
				Type:       "Generic",
			}
		default:
			response = Message{
				Message:    fmt.Sprintf("Command '%s' not found", message.Message),
				Identifier: message.Identifier,
				Type:       "Warning",
			}
		}

		js, err := json.Marshal(response)
		if err != nil {
			log.Printf("marshal response error: %v\n", err)
			return
		}

		if err := ws.WriteMessage(websocket.TextMessage, js); err != nil {
			log.Printf("write response error: %v\n", err)
			return
		}
	})

	return server
}

func TestDial(t *testing.T) {
	server := httptest.NewServer(handlers())
	defer server.Close()

	t.Run("connection refused", func(t *testing.T) {
		conn, err := Dial("127.0.0.2:12345", "password")
		if !assert.Error(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}

		assert.EqualError(t, err, "dial tcp 127.0.0.2:12345: connect: connection refused")
	})

	t.Run("authentication failed", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), "wrong")
		if !assert.Error(t, err) {
			assert.NoError(t, conn.Close())
		}

		assert.EqualError(t, err, "websocket: bad handshake")
	})

	t.Run("auth success", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword, SetDialTimeout(5*time.Second))
		if assert.NoError(t, err) {
			assert.NoError(t, conn.Close())
		}
	})
}

func TestConn_Execute(t *testing.T) {
	server := httptest.NewServer(handlers())
	defer server.Close()

	t.Run("incorrect command", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("")
		assert.Equal(t, err, ErrCommandEmpty)
		assert.Equal(t, 0, len(result))

		result, err = conn.Execute(string(make([]byte, 1001)))
		assert.Equal(t, err, ErrCommandTooLong)
		assert.Equal(t, 0, len(result))
	})

	t.Run("closed network connection", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, conn.Close())

		result, err := conn.Execute("status")
		assert.EqualError(t, err, fmt.Sprintf("write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("read deadline", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword, SetDeadline(1*time.Second))
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("deadline")
		assert.EqualError(t, err, fmt.Sprintf("read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr()))
		assert.Equal(t, 0, len(result))
	})

	t.Run("unknown command", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("random")
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Command '%s' not found", "random"), result)
	})

	t.Run("success command", func(t *testing.T) {
		conn, err := Dial(server.Listener.Addr().String(), MockPassword)
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, conn.Close())
		}()

		result, err := conn.Execute("status")
		assert.NoError(t, err)
		assert.Equal(t, MockCommandStatusResponseText, result)
	})

	// Environment variable TEST_RUST_SERVER allows to sends commands to real
	// Rust server.
	//
	// Some Rust commands:
	// console.tail 5
	// status
	// playerlist
	// serverinfo
	if run := getVar("TEST_RUST_SERVER", "false"); run == "true" {
		addr := getVar("TEST_RUST_SERVER_ADDR", "127.0.0.1:28016")
		password := getVar("TEST_RUST_SERVER_PASSWORD", "docker")

		t.Run("rust server", func(t *testing.T) {
			conn, err := Dial(addr, password)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				assert.NoError(t, conn.Close())
			}()

			result, err := conn.Execute("status")
			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			fmt.Println(result)
		})
	}
}

// getVar returns environment variable or default value.
func getVar(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
