package server

import "testing"

func TestMonitorConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.MonitorConfig(), cfg.Monitor)
}
