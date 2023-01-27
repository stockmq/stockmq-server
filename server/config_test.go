package server

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	expectDeepEqual(t, cfg, cfg)
}
