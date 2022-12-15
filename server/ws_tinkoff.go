package server

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	tinkoffEventOrderbook = "orderbook"
	tinkoffEventCandle    = "candle"
	tinkoffEventError     = "error"
)

var tinkoffIntervals = map[string]string{
	"1min":  "1m",
	"2min":  "2m",
	"3min":  "3m",
	"5min":  "5m",
	"10min": "10m",
	"15min": "15m",
	"30min": "30m",
	"hour":  "1h",
	"2hour": "2h",
	"4hour": "4h",
	"day":   "1d",
	"week":  "1w",
	"month": "1M",
}

type TinkoffEvent struct {
	Event   string          `json:"event"`
	Time    time.Time       `json:"time"`
	Payload json.RawMessage `json:"payload"`
}

type TinkoffError struct {
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
}

type TinkoffCandle struct {
	Open     float64   `json:"o"`
	Close    float64   `json:"c"`
	High     float64   `json:"h"`
	Low      float64   `json:"l"`
	Volume   int       `json:"v"`
	Time     time.Time `json:"time"`
	Interval string    `json:"interval"`
	Figi     string    `json:"figi"`
}

type TinkoffOrderBook struct {
	Figi  string      `json:"figi"`
	Depth int         `json:"depth"`
	Bids  [][]float64 `json:"bids"`
	Asks  [][]float64 `json:"asks"`
}

// WSTinkoffHandler handles messages from tinkoff.
func WSTinkoffHandler(s *Server, w *WSConnection, msg []byte) error {
	rcv := time.Now()

	message := &TinkoffEvent{}
	if err := json.Unmarshal(msg, message); err != nil {
		return err
	}

	switch message.Event {
	case tinkoffEventOrderbook:
		c := &TinkoffOrderBook{}
		if err := json.Unmarshal(message.Payload, c); err != nil {
			return err
		}

		header := MessageHeader{
			Symbol:  c.Figi,
			Time:    message.Time.UnixMicro(),
			TimeSrv: message.Time.UnixMicro(),
			TimeRcv: rcv.UnixMicro(),
			Source:  w.wsConfig.Name,
		}

		r := &Quote{
			MessageHeader: header,

			AsksDepth: c.Depth,
			BidsDepth: c.Depth,
			Asks:      make([][]string, c.Depth),
			Bids:      make([][]string, c.Depth),
		}

		for i := 0; i < c.Depth; i++ {
			r.Asks[i] = []string{fmt.Sprintf("%g", c.Asks[i][0]), fmt.Sprintf("%g", c.Asks[i][1])}
			r.Bids[i] = []string{fmt.Sprintf("%g", c.Bids[i][0]), fmt.Sprintf("%g", c.Bids[i][1])}
		}

		return s.ProcessQuote(w.wsConfig.Name, r)
	case tinkoffEventCandle:
		c := &TinkoffCandle{}
		if err := json.Unmarshal(message.Payload, c); err != nil {
			return err
		}

		header := MessageHeader{
			Symbol:  c.Figi,
			Time:    c.Time.UnixMicro(),
			TimeSrv: message.Time.UnixMicro(),
			TimeRcv: rcv.UnixMicro(),
			Source:  w.wsConfig.Name,
		}

		r := &Candle{
			MessageHeader: header,

			Interval: tinkoffIntervals[c.Interval],
			Open:     fmt.Sprintf("%g", c.Open),
			High:     fmt.Sprintf("%g", c.High),
			Low:      fmt.Sprintf("%g", c.Low),
			Close:    fmt.Sprintf("%g", c.Close),
			Volume:   fmt.Sprintf("%d", c.Volume),
		}

		return s.ProcessCandle(w.wsConfig.Name, r)
	case tinkoffEventError:
		c := &TinkoffError{}
		if err := json.Unmarshal(message.Payload, c); err != nil {
			return err
		}
		s.Errorf("WSS %s: %+v", w.wsConfig.Name, c)
	}
	return nil
}
