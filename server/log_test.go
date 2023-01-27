package server

import (
	"testing"
)

func TestLoggerConfig(t *testing.T) {
	cfg := DefaultConfig()
	srv, _ := NewServer(DefaultConfig())
	expectDeepEqual(t, srv.LoggerConfig(), cfg.Logger)
}

func TestLogDebugf(t *testing.T) {
	expectOutput(t, func() {
		s, _ := NewServer(DefaultConfig())
		s.Debugf("foo")
	}, "")

	expectOutput(t, func() {
		s, _ := NewServer(DefaultConfig())
		s.config.Logger.Debug = true
		s.Debugf("foo")
	}, "[DBG] foo\n")
}

func TestLogNoticef(t *testing.T) {
	expectOutput(t, func() {
		s, _ := NewServer(DefaultConfig())
		s.Noticef("foo")
	}, "[INF] foo\n")
}

func TestLogWarnf(t *testing.T) {
	expectOutput(t, func() {
		s, _ := NewServer(DefaultConfig())
		s.Warnf("foo")
	}, "[WRN] foo\n")
}

func TestLogErrorf(t *testing.T) {
	expectOutput(t, func() {
		s, _ := NewServer(DefaultConfig())
		s.Errorf("foo")
	}, "[ERR] foo\n")
}
