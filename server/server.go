/*
Package server implements all methods required to process messages.
*/
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

var (
	ErrServerShutdown = errors.New("server was shutdown or already started")
)

type WSConnection struct {
	sync.RWMutex

	wsConfig WSConfig
	wsConn   *websocket.Conn
	wsReconn atomic.Bool
}

type Server struct {
	config           ServerConfig
	quitCh           chan struct{}
	startupComplete  chan struct{}
	shutdownComplete chan struct{}

	mu sync.RWMutex

	running  atomic.Bool
	shutdown atomic.Bool

	// Monitor
	monitorServer *http.Server

	// GRPC
	grpcListener net.Listener
	grpcServer   *grpc.Server

	// NATS
	ncMu     sync.RWMutex
	ncConn   *nats.Conn
	ncReconn atomic.Bool

	// WebSocket connections.
	wsConnections map[string]*WSConnection
}

// NewServer will setup a new server instance struct.
func NewServer(config ServerConfig) (*Server, error) {
	s := &Server{}
	s.config = config
	s.quitCh = make(chan struct{})
	s.startupComplete = make(chan struct{})
	s.shutdownComplete = make(chan struct{})
	s.wsConnections = make(map[string]*WSConnection)

	// Create list of connections
	for _, cfg := range s.config.WebSocket {
		if cfg.Enabled {
			s.wsConnections[cfg.Name] = &WSConnection{wsConfig: cfg}
		}
	}
	return s, nil
}

// Start the server
func (s *Server) Start() error {
	// Prevent multiple Start() calls
	if s.IsRunning() || s.IsShutdown() {
		return ErrServerShutdown
	}

	// Set running to true to avoid race between multiple Start() calls
	s.Noticef("Starting StockMQ Server")
	s.running.Store(true)

	// Start signal handler
	s.HandleSignals()

	// Start monitor
	s.StartMonitor()

	// Start GRPC
	s.StartGRPC()

	// Start NATS client
	s.StartNATS()

	// Start WebSockets
	for _, conn := range s.wsConnections {
		go s.StartWS(conn)
	}

	// Notify that server startup completed
	close(s.startupComplete)

	return nil
}

// Shutdown will shutdown the server instance.
func (s *Server) Shutdown() {
	// Prevent multiple Shutdown() calls
	if s.IsShutdown() {
		return
	}

	// Set shutdown to true to avoid race between multiple Shutdown() calls
	s.shutdown.Store(true)

	// Kick NATS if its running
	s.Noticef("Shutting down the NATS connection...")
	s.CloseNATS()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Kick WebSocket if its running
	for k, conn := range s.wsConnections {
		conn.Lock()
		if conn.wsConn != nil {
			s.Noticef("Shutting down the %s websocket...", k)
			conn.wsConn.Close()
			conn.wsConn = nil
		}
		conn.Unlock()
	}

	// Kick off HTTP monitor
	if s.monitorServer != nil {
		s.Noticef("Shutting down the monitor...")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := s.monitorServer.Shutdown(ctx); err != nil {
			s.Errorf("error during graceful shutdown: %v", err)
		}
		s.monitorServer = nil
	}

	// Kick off GRPC server
	if s.grpcListener != nil {
		s.Noticef("Shutting down the GRPC server...")
		s.grpcServer.Stop()
		s.grpcListener.Close()
	}

	// Release go routines
	close(s.quitCh)

	// Notify the shutdown is complete
	close(s.shutdownComplete)
}

// WaitForShutdown will block until the server has been fully shutdown.
func (s *Server) WaitForShutdown() {
	<-s.shutdownComplete
}

// IsRunning returns whether service is running.
func (s *Server) IsRunning() bool {
	return s.running.Load()
}

// IsShutdown returns whether server is performing shutdown.
func (s *Server) IsShutdown() bool {
	return s.shutdown.Load()
}
