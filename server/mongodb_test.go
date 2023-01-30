package server

import "testing"

func TestMongoDBConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.MongoDBConfig(), cfg.MongoDB)
}
