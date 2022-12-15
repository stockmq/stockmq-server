package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
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

// InfluxDBPoint returns the Point
func (m *Candle) InfluxDBPoint() *write.Point {
	p := influxdb2.NewPoint(
		"candle",
		map[string]string{"symbol": m.Symbol, "source": m.Source, "interval": m.Interval},
		map[string]interface{}{
			"time":     m.Time,
			"time_srv": m.TimeSrv,
			"time_rcv": m.TimeRcv,
			"open":     Unwrap(strconv.ParseFloat(m.Open, 64)),
			"high":     Unwrap(strconv.ParseFloat(m.High, 64)),
			"low":      Unwrap(strconv.ParseFloat(m.Low, 64)),
			"close":    Unwrap(strconv.ParseFloat(m.Close, 64)),
			"volume":   Unwrap(strconv.ParseFloat(m.Volume, 64)),
		},
		time.UnixMicro(m.TimeSrv),
	)

	return p
}

// InfluxDBPoint returns the Point
func (m *Quote) InfluxDBPoint() *write.Point {
	p := influxdb2.NewPoint(
		"quote",
		map[string]string{"symbol": m.Symbol, "source": m.Source},
		map[string]interface{}{
			"time_srv":   m.TimeSrv,
			"time_rcv":   m.TimeRcv,
			"bids":       Unwrap(json.Marshal(m.Bids)),
			"bids_depth": m.BidsDepth,
			"asks":       Unwrap(json.Marshal(m.Asks)),
			"asks_depth": m.AsksDepth,
		},
		time.UnixMicro(m.Time),
	)

	return p
}

func Unwrap[T any](v T, err error) T {
	return v
}
