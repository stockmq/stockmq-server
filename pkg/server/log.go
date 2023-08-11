package server

import (
	"log/slog"
	"os"
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

// NewLogger returns log/slog logger with configured options
func NewLogger(config ServerConfig) *slog.Logger {
	opts := &slog.HandlerOptions{}
	if config.Logger.Debug {
		opts.Level = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
