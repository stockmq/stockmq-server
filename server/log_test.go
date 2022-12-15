package server

import (
	"bytes"
	"log"
	"testing"
)

func expectOutput(t *testing.T, f func(), expected string) {
	var buf bytes.Buffer
	writer := log.Writer()
	flags := log.Flags()

	log.SetOutput(&buf)
	log.SetFlags(flags &^ (log.Ldate | log.Ltime))

	defer func() {
		log.SetFlags(flags)
		log.SetOutput(writer)
	}()
	f()
	expectDeepEqual(t, buf.String(), expected)
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
