package server

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents server health status.
type HealthStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// healthStatus returns the current status of the server.
func (s *Server) healthStatus() *HealthStatus {
	failures := []string{}

	// Check InfluxDB
	if s.InfluxDBConfig().Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if pong, err := s.dbClient.Ping(ctx); err != nil || !pong {
			failures = append(failures, "influxdb")
		}

	}

	// Check NATS
	s.ncMu.RLock()
	if s.ncConn == nil || !s.ncConn.IsConnected() {
		failures = append(failures, "nats")
	}
	s.ncMu.RUnlock()

	if len(failures) == 0 {
		return &HealthStatus{Status: "ok"}
	} else {
		return &HealthStatus{Status: "error", Error: fmt.Sprintf("not connected to %v", failures)}
	}
}
