package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATS Configuration
type NATSConfig struct {
	Name        string `xml:"Name"`
	URL         string `xml:"URL"`
	RetryDelay  int    `xml:"RetryDelay"`
	NoReconnect bool   `xml:"NoReconnect"`
}

// NATSSubjecter provided methods to generate subjects from entities.
type NATSSubjecter interface {
	NATSSubject() string
}

// DefaultNATSConfig returns default NATS config
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		Name:        "StockMQ",
		URL:         "nats://127.0.0.1:4222",
		RetryDelay:  5,
		NoReconnect: false,
	}
}

// NATSConfig returns NATS configuration.
func (s *Server) NATSConfig() NATSConfig {
	return s.ServerConfig().NATS
}

// NATSSubject returns the subject for the candle message
func (m *Candle) NATSSubject() string {
	return fmt.Sprintf("C.%s.%s.%s", m.Interval, m.Symbol, m.Source)
}

// NATSSubject returns the subject for the candle message
func (m *Quote) NATSSubject() string {
	return fmt.Sprintf("Q.%s.%s", m.Symbol, m.Source)
}

// StartNATS starts the NATS client.
func (s *Server) StartNATS() {
	cfg := s.NATSConfig()
	s.logger.Info("Starting NATS connection", "url", cfg.URL)

	nc, err := nats.Connect(cfg.URL, cfg.NATSOptions()...)
	if err != nil {
		s.HandleNATSError(err)
		return
	}

	s.ncMu.Lock()
	s.ncConn = nc
	s.ncMu.Unlock()
}

// NATSOptions returns a list of NATS connection options.
func (c *NATSConfig) NATSOptions() []nats.Option {
	options := []nats.Option{}
	if c.Name != "" {
		options = append(options, nats.Name(c.Name))
	}
	if c.NoReconnect {
		options = append(options, nats.NoReconnect())
	}
	return options
}

// IsNATSReconnecting returns whether NATS is scheduled to reconnect.
func (s *Server) IsNATSReconnecting() bool {
	return s.ncReconn.Load()
}

// CloseNATS closes the NATS connection.
func (s *Server) CloseNATS() {
	s.ncMu.Lock()
	defer s.ncMu.Unlock()
	if s.ncConn != nil {
		s.ncConn.Close()
		s.ncConn = nil
	}
}

// HandleNATSError handles NATS errors.
func (s *Server) HandleNATSError(err error) {
	// Do nothing if the server is shutting down or NATS is reconnecting
	if s.IsShutdown() || s.IsNATSReconnecting() {
		return
	}

	// Close NATS connection
	s.logger.Error("NATS Error", "error", err)
	s.CloseNATS()

	// Runs goroutine to restart NATS connection after RetryDelay
	go func() {
		cfg := s.NATSConfig()
		s.ncReconn.Store(true)
		s.logger.Info("NATS: Reconnecting", "url", cfg.URL, "delay", cfg.RetryDelay)

		select {
		case <-s.quitCh:
			return
		case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
			s.ncReconn.Store(false)
			s.StartNATS()
		}
	}()
}

// NATSSend sends message to the NATS.
func (s *Server) NATSSend(object NATSSubjecter) {
	s.ncMu.Lock()
	nc := s.ncConn
	s.ncMu.Unlock()

	if nc != nil {
		if b, err := json.Marshal(object); err == nil {
			if err := nc.Publish(object.NATSSubject(), b); err != nil {
				s.HandleNATSError(err)
			}
		}
	}
}
