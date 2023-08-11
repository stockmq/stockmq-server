package server

import (
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

// Unwrap ignores the error.
func Unwrap[T any](v T, err error) T {
	return v
}
