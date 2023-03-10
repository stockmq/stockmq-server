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
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2_api "github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	ErrServerShutdown = errors.New("server was shutdown or already started")
)

// Server Configuration.
type ServerConfig struct {
	Logger    LoggerConfig   `xml:"Logger"`
	Monitor   MonitorConfig  `xml:"Monitor"`
	MongoDB   MongoDBConfig  `xml:"MongoDB"`
	InfluxDB  InfluxDBConfig `xml:"InfluxDB"`
	NATS      NATSConfig     `xml:"NATS"`
	GRPC      GRPCConfig     `xml:"GRPC"`
	WebSocket []WSConfig     `xml:"WebSocket"`
}

// DefaultConfig returns default ServerConfig.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Logger:   DefaultLoggerConfig(),
		Monitor:  DefaultMonitorConfig(),
		MongoDB:  DefaultMongoDBConfig(),
		InfluxDB: DefaultInfluxDBConfig(),
		NATS:     DefaultNATSConfig(),
		GRPC:     DefaultGRPCConfig(),
	}
}

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

	// MongoDB
	mongoMu     sync.RWMutex
	mongoReconn atomic.Bool
	mongoClient *mongo.Client

	// InfluxDB
	dbClient influxdb2.Client
	dbWriter influxdb2_api.WriteAPI

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

// Config returns a copy of Server configuration.
func (s *Server) ServerConfig() ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
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

	// Start MongoDB client
	if s.MongoDBConfig().Enabled {
		s.StartMongoDB()
	}

	// Start InfluxDB client
	if s.InfluxDBConfig().Enabled {
		s.StartInfluxDB()
	}

	// Start GRPC
	if err := s.StartGRPC(); err != nil {
		return err
	}

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

	// Kick off MongoDB
	if s.mongoClient != nil {
		s.Noticef("Shutting down the MongoDB connection...")
		s.CloseMongoDB()
	}

	// Kick off InfluxDB
	if s.dbWriter != nil {
		s.Noticef("Shutting down the InfluxDB connection...")
		s.dbWriter.Flush()

		if s.dbClient != nil {
			s.dbClient.Close()
		}
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
