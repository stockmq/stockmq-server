package server

import (
	"encoding/json"
	"time"
)

const (
	binanceEventKline       = "kline"
	binenceEventDepthUpdate = "depthUpdate"
)

type BinanceMessage struct {
	ID        *int    `json:"id"`
	EventType *string `json:"e"`
	EventTime *int64  `json:"E"`
	ErrorCode *int    `json:"code"`
}

type BinanceResult struct {
	Result *interface{} `json:"result"`
	ID     int          `json:"id"`
}

type BinanceError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type BinanceKline struct {
	StartTime            int64  `json:"t"`
	EndTime              int64  `json:"T"`
	Symbol               string `json:"s"`
	Interval             string `json:"i"`
	FirstTradeID         int64  `json:"f"`
	LastTradeID          int64  `json:"L"`
	Open                 string `json:"o"`
	Close                string `json:"c"`
	High                 string `json:"h"`
	Low                  string `json:"l"`
	Volume               string `json:"v"`
	TradeNum             int64  `json:"n"`
	IsFinal              bool   `json:"x"`
	QuoteVolume          string `json:"q"`
	ActiveBuyVolume      string `json:"V"`
	ActiveBuyQuoteVolume string `json:"Q"`
	Ignore               string `json:"B"`
}

type BinanceCandle struct {
	EventType string       `json:"e"`
	EventTime int64        `json:"E"`
	Symbol    string       `json:"s"`
	Kline     BinanceKline `json:"k"`
}

type BinanceOrderBook struct {
	EventName     string     `json:"e"`
	EventType     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	LastUpdateID  int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

// WSBinanceHandler process message from the binance stream.
func WSBinanceHandler(s *Server, w *WSConnection, msg []byte) error {
	rcv := time.Now()

	message := &BinanceMessage{}
	if err := json.Unmarshal(msg, message); err != nil {
		return err
	}

	switch {
	case message.EventType != nil:
		switch *message.EventType {
		case binanceEventKline:
			c := &BinanceCandle{}
			if err := json.Unmarshal(msg, c); err != nil {
				return err
			}

			r := &Candle{
				MessageHeader: MessageHeader{
					Symbol:  c.Symbol,
					Time:    c.Kline.StartTime * 1000,
					TimeSrv: c.EventTime * 1000,
					TimeRcv: rcv.UnixMicro(),
					Source:  w.wsConfig.Name,
				},

				Interval: c.Kline.Interval,
				Open:     c.Kline.Open,
				High:     c.Kline.High,
				Low:      c.Kline.Low,
				Close:    c.Kline.Close,
				Volume:   c.Kline.Volume,
			}

			return s.ProcessCandle(w.wsConfig.Name, r)
		case binenceEventDepthUpdate:
			c := &BinanceOrderBook{}
			if err := json.Unmarshal(msg, c); err != nil {
				return err
			}

			r := &Quote{
				MessageHeader: MessageHeader{
					Symbol:  c.Symbol,
					Time:    *message.EventTime * 1000,
					TimeSrv: *message.EventTime * 1000,
					TimeRcv: rcv.UnixMicro(),
					Source:  w.wsConfig.Name,
				},

				AsksDepth: len(c.Asks),
				Asks:      c.Asks,
				BidsDepth: len(c.Bids),
				Bids:      c.Bids,
			}

			return s.ProcessQuote(w.wsConfig.Name, r)
		default:
			s.Errorf("WSS %s: unknown event '%s'", w.wsConfig.Name, *message.EventType)
		}
	case message.ErrorCode != nil:
		m := &BinanceError{}
		if err := json.Unmarshal(msg, m); err != nil {
			return err
		}
		s.Errorf("WSS %s: %+v", w.wsConfig.Name, m)
	case message.ID != nil:
		m := &BinanceError{}
		if err := json.Unmarshal(msg, m); err != nil {
			return err
		}
		s.Debugf("WSS %s: result %+v", w.wsConfig.Name, m)
	default:
		s.Debugf("WSS %s: unknown message %s", w.wsConfig.Name, msg)
	}
	return nil
}
