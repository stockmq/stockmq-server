package server

import (
	"reflect"
	"testing"
)

func expectDeepEqual(t *testing.T, i interface{}, expected interface{}) {
	if reflect.TypeOf(i) != reflect.TypeOf(expected) {
		t.Fatalf("Expected value to be %T, got %T", expected, i)
	}

	if !reflect.DeepEqual(i, expected) {
		t.Fatalf("Value is incorrect.\ngot: %+v\nexpected: %+v", i, expected)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	expectDeepEqual(t, cfg, cfg)
}

func TestLoggerConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.LoggerConfig(), cfg.Logger)
}

func TestMonitorConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.MonitorConfig(), cfg.Monitor)
}

func TestNATSConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.NATSConfig(), cfg.NATS)
}

func TestGRPCConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.GRPCConfig(), cfg.GRPC)
}
