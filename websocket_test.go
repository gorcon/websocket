package websocket

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDial(t *testing.T) {
	server := httptest.NewServer(mockHandlers())
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
	server := httptest.NewServer(mockHandlers())
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
}
