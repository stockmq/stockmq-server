package server

import (
	"fmt"
)

type HealthStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// healthStatus returns the current status of the server.
func (s *Server) healthStatus() *HealthStatus {
	failures := []string{}

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
