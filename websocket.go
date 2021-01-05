package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// DefaultDialTimeout provides default auth timeout to remote server.
	DefaultDialTimeout = 5 * time.Second

	// DefaultDeadline provides default deadline to tcp read/write operations.
	DefaultDeadline = 5 * time.Second

	// MaxCommandLen is an artificial restriction, but it will help in case of random
	// large queries.
	MaxCommandLen = 1000
)

var (
	// ErrAuthFailed is returned when the package id from authentication
	// response is -1.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrCommandTooLong is returned when executed command length is bigger
	// than MaxCommandLen characters.
	ErrCommandTooLong = errors.New("command too long")

	// ErrCommandEmpty is returned when executed command length equal 0.
	ErrCommandEmpty = errors.New("command too small")
)

// Conn represents a WebSocket connection.
type Conn struct {
	conn     *websocket.Conn
	settings Settings
}

// Dial creates a new authorized WebSocket dialer connection.
func Dial(address string, password string, options ...Option) (*Conn, error) {
	settings := DefaultSettings

	for _, option := range options {
		option(&settings)
	}

	u := url.URL{Scheme: "ws", Host: address, Path: password}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if err.Error() == `malformed HTTP response "\x88\x02\x03\xe8"` {
			return nil, ErrAuthFailed
		}

		return nil, fmt.Errorf("webrcon: %w", err)
	}

	rand.Seed(time.Now().UnixNano())

	client := Conn{conn: conn, settings: settings}

	return &client, nil
}

// Execute sends command string to execute to the remote server.
func (c *Conn) Execute(command string) (string, error) {
	if command == "" {
		return "", ErrCommandEmpty
	}

	if len(command) > MaxCommandLen {
		return "", ErrCommandTooLong
	}

	request := newMessage(command)

	data, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("webrcon: %w", err)
	}

	if err := c.write(data); err != nil {
		return "", err
	}

	for {
		p, err := c.read()
		if err != nil {
			return "", err
		}

		var response Message
		if err := json.Unmarshal(p, &response); err != nil {
			return "", fmt.Errorf("webrcon: %w", err)
		}

		if response.Identifier == request.Identifier {
			return response.Message, nil
		}
	}
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) write(data []byte) error {
	if c.settings.deadline != 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.settings.deadline)); err != nil {
			return fmt.Errorf("webrcon: %w", err)
		}
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("webrcon: %w", err)
	}

	return nil
}

func (c *Conn) read() ([]byte, error) {
	if c.settings.deadline != 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.settings.deadline)); err != nil {
			return nil, fmt.Errorf("webrcon: %w", err)
		}
	}

	_, p, err := c.conn.ReadMessage()
	if err != nil {
		return p, fmt.Errorf("webrcon: %w", err)
	}

	return p, nil
}
