package server

import "testing"

func TestGRPCConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.GRPCConfig(), cfg.GRPC)
}
