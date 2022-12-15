package server

import (
	"fmt"
)

// MessageHeader represents common fields for each message
type MessageHeader struct {
	Symbol  string `json:"symbol"`
	Source  string `json:"source"`
	Time    int64  `json:"time"`
	TimeSrv int64  `json:"time_srv"`
	TimeRcv int64  `json:"time_rcv"`
}

type Candle struct {
	MessageHeader

	Interval string `json:"interval"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
}

type Quote struct {
	MessageHeader

	BidsDepth int        `json:"bids_depth"`
	Bids      [][]string `json:"bids"`
	AsksDepth int        `json:"asks_depth"`
	Asks      [][]string `json:"asks"`
}

// NATSSubject returns the subject for the candle message
func (m *Candle) NATSSubject() string {
	return fmt.Sprintf("C.%s.%s.%s", m.Interval, m.Symbol, m.Source)
}

// NATSSubject returns the subject for the candle message
func (m *Quote) NATSSubject() string {
	return fmt.Sprintf("Q.%s.%s", m.Symbol, m.Source)
}
