package server

import (
	"bytes"
	"log"
	"reflect"
	"testing"
)

// expectDeepEqual compares two interfaces.
func expectDeepEqual(t *testing.T, i interface{}, expected interface{}) {
	if reflect.TypeOf(i) != reflect.TypeOf(expected) {
		t.Fatalf("Expected value to be %T, got %T", expected, i)
	}

	if !reflect.DeepEqual(i, expected) {
		t.Fatalf("Value is incorrect.\ngot: %+v\nexpected: %+v", i, expected)
	}
}

// expectOutput tests stdout output to match expected string.
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

// Unwrap ignores the error.
func Unwrap[T any](v T, err error) T {
	return v
}
