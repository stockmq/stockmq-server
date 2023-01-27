package server

import "testing"

func TestStartDB(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(cfg)
	srv.StartInfluxDB()

	if srv.dbClient == nil {
		t.Fatalf("InfluxDB client expected to be not nil")
	}

	if srv.dbWriter == nil {
		t.Fatalf("InfluxDB writer expected to be not nil")
	}
}
