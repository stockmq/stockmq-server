package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WSMsgHandler func(s *Server, w *WSConnection, msg []byte) error

var (
	Handlers = map[string]WSMsgHandler{
		"Debug":   WSDebugHandler,
		"Binance": WSBinanceHandler,
		"Tinkoff": WSTinkoffHandler,
	}
)

// IsWSReconnecting returns whether websocket is scheduled to reconnect.
func (c *WSConnection) IsWSReconnecting() bool {
	return c.wsReconn.Load()
}

// WSKeepAlive enabled Ping-Pong with given timeout.
func (s *Server) WSKeepAlive(cfg WSConfig, c *websocket.Conn) {
	if cfg.PingTimeout < 1 {
		s.Debugf("WSS %s: PingTimeout is less than 1. KeepAlive disabled.", cfg.Name)
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
				s.Errorf("WSS %s: error sending ping message: %v", cfg.Name, err)
				return
			}
			select {
			case <-ticker.C:
				if time.Since(lastResponse) > timeout {
					s.Errorf("WSS %s: ping-pong timeout", cfg.Name)
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
	s.Noticef("Starting WebSocket connection %s (%s)", cfg.Name, cfg.URL)

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

	s.Errorf("WSS %s: %v", conn.wsConfig.Name, err)

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
		s.Noticef("WSS %s: Reconnecting in %d seconds", cfg.Name, cfg.RetryDelay)

		select {
		case <-s.quitCh:
			return
		case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			conn.wsReconn.Store(false)
			s.StartWS(conn)
		}
	}()
}
