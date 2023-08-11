package server

import (
	"testing"
)

func TestLoggerConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.LoggerConfig(), cfg.Logger)
}
