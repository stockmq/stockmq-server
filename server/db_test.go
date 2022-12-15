package server

import "testing"

func TestStartDB(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(cfg)
	srv.StartDB()

	if srv.dbClient == nil {
		t.Fatalf("DB client expected to be not nil")
	}

	if srv.dbWriter == nil {
		t.Fatalf("DB writer expected to be not nil")
	}
}
