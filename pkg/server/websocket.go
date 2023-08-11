package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket Handler callback.
type WSMsgHandler func(s *Server, w *WSConnection, msg []byte) error

// WebSocket Configuration.
type WSConfig struct {
	Name         string   `xml:"Name"`
	Enabled      bool     `xml:"Enabled"`
	URL          string   `xml:"URL"`
	Handler      string   `xml:"Handler"`
	DialTimeout  int      `xml:"DialTimeout"`
	RetryDelay   int      `xml:"RetryDelay"`
	PingTimeout  int      `xml:"PingTimeout"`
	ReadLimit    int64    `xml:"ReadLimit"`
	Headers      []Header `xml:"Header"`
	InitMessages []string `xml:"InitMessage"`
}

var (
	Handlers = map[string]WSMsgHandler{}
)

// IsWSReconnecting returns whether websocket is scheduled to reconnect.
func (c *WSConnection) IsWSReconnecting() bool {
	return c.wsReconn.Load()
}

// WSKeepAlive enabled Ping-Pong with given timeout.
func (s *Server) WSKeepAlive(cfg WSConfig, c *websocket.Conn) {
	if cfg.PingTimeout < 1 {
		return
	}

	timeout := time.Duration(cfg.PingTimeout) * time.Second
	ticker := time.NewTicker(timeout)
	lastResponse := time.Now()

	c.SetPongHandler(func(appData string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		defer c.Close()

		for {
			deadline := time.Now().Add(timeout / 2)

			if err := c.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
				s.logger.Error("WSS error sending ping message", "name", cfg.Name, "error", err)
				return
			}
			select {
			case <-ticker.C:
				if time.Since(lastResponse) > timeout {
					s.logger.Error("WSS ping-pong timeout", "name", cfg.Name)
					return
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}

// StartWS establishes websocket connection.
func (s *Server) StartWS(conn *WSConnection) {
	// Get config
	cfg := conn.wsConfig

	// Log the message
	s.logger.Info("Starting WebSocket connection", "name", cfg.Name, "url", cfg.URL)

	// Find the message handler for the given Type
	handler := Handlers[cfg.Handler]
	if handler == nil {
		s.WSHandleError(conn, fmt.Errorf("WSS: Cannot find handler '%v'", cfg.Handler))
		return
	}

	// Headers
	headers := make(http.Header)
	for _, header := range cfg.Headers {
		headers.Add(header.Name, header.Text)
	}

	// Dial WebSocket
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DialTimeout)*time.Second)
	defer cancel()
	c, _, err := websocket.DefaultDialer.DialContext(ctx, cfg.URL, headers)
	if err != nil {
		s.WSHandleError(conn, err)
		return
	}

	conn.Lock()
	conn.wsConn = c
	conn.wsConn.SetReadLimit(cfg.ReadLimit)
	conn.Unlock()

	s.WSKeepAlive(cfg, c)

	// Send init messages
	for _, msg := range cfg.InitMessages {
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			s.WSHandleError(conn, err)
			return
		}
	}

	// Run goroutine to process incoming messages
	go func() {
		defer c.Close()

		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				s.WSHandleError(conn, err)
				return
			}

			if err := handler(s, conn, raw); err != nil {
				s.WSHandleError(conn, err)
			}
		}
	}()
}

// WSHandleError handles the error.
func (s *Server) WSHandleError(conn *WSConnection, err error) {
	// Do nothing if the server is shutting down or WebSocket is reconnecting.
	if s.IsShutdown() || conn.IsWSReconnecting() {
		return
	}

	s.logger.Error("WSS Error", "name", conn.wsConfig.Name, "error", err)

	conn.Lock()
	if conn.wsConn != nil {
		conn.wsConn.Close()
		conn.wsConn = nil
	}
	cfg := conn.wsConfig
	conn.Unlock()

	// Runs goroutine to restart WebSocket connection after RetryDelay
	go func() {
		conn.wsReconn.Store(true)
		s.logger.Info("WSS Reconnecting", "name", cfg.Name, "delay", cfg.RetryDelay)

		select {
		case <-s.quitCh:
			return
		case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			conn.wsReconn.Store(false)
			s.StartWS(conn)
		}
	}()
}
