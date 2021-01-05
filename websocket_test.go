package websocket_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorcon/websocket"
	gorilla "github.com/gorilla/websocket"
)

const StatusMessage = `status
hostname: Rust Server [DOCKER]
version : 2260 secure (secure mode enabled, connected to Steam3)
map     : Procedural Map
players : 0 (500 max) (0 queued) (0 joining)

id name ping connected addr owner violation kicks
`

func handlers() http.Handler {
	server := http.NewServeMux()

	var upgrader = gorilla.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	server.HandleFunc("/password", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v\n", err)
			return
		}

		defer ws.Close()

		var response websocket.Message

		// Receive message.
		_, p, err := ws.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(), "websocket: close 1006 (abnormal closure): unexpected EO") {
				log.Printf("read message error: %v\n", err)
			}
			return
		}

		var message websocket.Message
		if err := json.Unmarshal(p, &message); err != nil {
			// TODO: What Rust responses on read message fail?
			fmt.Println(string(p))
			log.Printf("unmarshal message error: %v\n", err)
			return
		}

		switch message.Message {
		case "status":
			response = websocket.Message{
				Message:    StatusMessage,
				Identifier: message.Identifier,
				Type:       "Generic",
			}
		case "deadline":
			time.Sleep(websocket.DefaultDeadline + 1*time.Second)
			response = websocket.Message{
				Message:    fmt.Sprintf("sleep for %d secends", websocket.DefaultDeadline+1*time.Second),
				Identifier: message.Identifier,
				Type:       "Generic",
			}
		default:
			response = websocket.Message{
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

		if err := ws.WriteMessage(gorilla.TextMessage, js); err != nil {
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
		wantErrContains := "connect: connection refused"

		_, err := websocket.Dial("127.0.0.2:12345", "password")
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		wantErrContains := "websocket: bad handshake"

		_, err := websocket.Dial(server.Listener.Addr().String(), "wrong")
		if err == nil || !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("got err %q, want to contain %q", err, wantErrContains)
		}
	})

	t.Run("auth success", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password", websocket.SetDialTimeout(5*time.Second))
		if err != nil {
			t.Errorf("got err %q, want %v", err, nil)
			return
		}

		conn.Close()
	})
}

func TestConn_Execute(t *testing.T) {
	server := httptest.NewServer(handlers())
	defer server.Close()

	t.Run("incorrect command", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("")
		if !errors.Is(err, websocket.ErrCommandEmpty) {
			t.Errorf("got err %q, want %q", err, websocket.ErrCommandEmpty)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}

		result, err = conn.Execute(string(make([]byte, 1001)))
		if !errors.Is(err, websocket.ErrCommandTooLong) {
			t.Errorf("got err %q, want %q", err, websocket.ErrCommandTooLong)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("closed network connection", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		conn.Close()

		result, err := conn.Execute("status")
		wantErrMsg := fmt.Sprintf("webrcon: write tcp %s->%s: use of closed network connection", conn.LocalAddr(), conn.RemoteAddr())
		if err == nil || err.Error() != wantErrMsg {
			t.Errorf("got err %q, want to contain %q", err, wantErrMsg)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("read deadline", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password", websocket.SetDeadline(1*time.Second))
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("deadline")
		wantErrMsg := fmt.Sprintf("webrcon: read tcp %s->%s: i/o timeout", conn.LocalAddr(), conn.RemoteAddr())
		if err == nil || err.Error() != wantErrMsg {
			t.Errorf("got err %q, want to contain %q", err, wantErrMsg)
		}

		if len(result) != 0 {
			t.Fatalf("got result len %d, want %d", len(result), 0)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("random")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}

		resultWant := fmt.Sprintf("Command '%s' not found", "random")
		if result != resultWant {
			t.Fatalf("got result %q, want %q", result, resultWant)
		}
	})

	t.Run("success command", func(t *testing.T) {
		conn, err := websocket.Dial(server.Listener.Addr().String(), "password")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}
		defer conn.Close()

		result, err := conn.Execute("status")
		if err != nil {
			t.Fatalf("got err %q, want %v", err, nil)
		}

		resultWant := StatusMessage
		if result != resultWant {
			t.Fatalf("got result %q, want %q", result, resultWant)
		}
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
			conn, err := websocket.Dial(addr, password)
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}
			defer conn.Close()

			result, err := conn.Execute("status")
			if err != nil {
				t.Fatalf("got err %q, want %v", err, nil)
			}

			if len(result) == 0 {
				t.Fatal("got result len 0, want not 0")
			}

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
