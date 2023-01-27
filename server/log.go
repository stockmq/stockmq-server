package server

import (
	"fmt"
	"log"
)

// Logger Configuration
type LoggerConfig struct {
	Debug bool `xml:"Debug"`
}

// DefaultLoggerConfig returns default Logger config.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Debug: false,
	}
}

// LoggerConfig returns Logger configuration.
func (s *Server) LoggerConfig() LoggerConfig {
	return s.ServerConfig().Logger
}

// Debugf prints the debug message to the log.
func (s *Server) Debugf(format string, v ...any) {
	if s.LoggerConfig().Debug {
		log.Printf("[DBG] %s", fmt.Sprintf(format, v...))
	}
}

// Noticef prints the notice message to the log.
func (s *Server) Noticef(format string, v ...any) {
	log.Printf("[INF] %s", fmt.Sprintf(format, v...))
}

// Warnf prints the warning message to the log.
func (s *Server) Warnf(format string, v ...any) {
	log.Printf("[WRN] %s", fmt.Sprintf(format, v...))
}

// Errorf prints the error message to the log.
func (s *Server) Errorf(format string, v ...any) {
	log.Printf("[ERR] %s", fmt.Sprintf(format, v...))
}
