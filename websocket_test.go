package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDial(t *testing.T) {
	t.Run("connection refused", func(t *testing.T) {
		conn, err := Dial("127.0.0.2:12345", "password")
		if !assert.Error(t, err) {
			// Close connection if established.
			assert.NoError(t, conn.Close())
		}
		assert.EqualError(t, err, "dial tcp 127.0.0.2:12345: connect: connection refused")
	})
}
