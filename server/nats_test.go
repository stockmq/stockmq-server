package server

import (
	"testing"
)

func TestCandleSubject(t *testing.T) {
	r := &Candle{MessageHeader: MessageHeader{Symbol: "foo", Source: "bar"}, Interval: "1m"}
	expectDeepEqual(t, r.NATSSubject(), "C.1m.foo.bar")
}

func TestQuoteSubject(t *testing.T) {
	r := &Quote{MessageHeader: MessageHeader{Symbol: "foo", Source: "bar"}}
	expectDeepEqual(t, r.NATSSubject(), "Q.foo.bar")
}
