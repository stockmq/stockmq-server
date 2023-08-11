package server

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
)

// Tinkoff configuration.
type TinkoffConfig struct {
	Name    string   `xml:"Name"`
	Enabled bool     `xml:"Enabled"`
	Config  string   `xml:"Config"`
	Candles []string `xml:"Candle"`
}

// Tinkoff connection.
type TinkoffConnection struct {
	sync.RWMutex

	tinkoffClient *investgo.Client
	tinkoffConfig TinkoffConfig
}

// Tinkoff Logger bridge to log/slog.
type TinkoffLogger struct {
	s *Server
	c *TinkoffConnection
}

func (l *TinkoffLogger) Infof(template string, args ...any) {
	l.s.logger.Info(fmt.Sprintf(template, args...), "source", l.c.tinkoffConfig.Name)
}

func (l *TinkoffLogger) Errorf(template string, args ...any) {
	l.s.logger.Error(fmt.Sprintf(template, args...), "source", l.c.tinkoffConfig.Name)
}

func (l *TinkoffLogger) Fatalf(template string, args ...any) {
	l.s.logger.Error(fmt.Sprintf(template, args...), "source", l.c.tinkoffConfig.Name)
}

// Start Tinkoff client.
func (s *Server) startTinkoff(conn *TinkoffConnection) {
	source := conn.tinkoffConfig.Name

	// Load configuration for SDK
	cfg, err := investgo.LoadConfig(conn.tinkoffConfig.Config)
	if err != nil {
		s.logger.Error("Tinkoff config error", "error", err)
		return
	}

	// Create Logger for Tinkoff
	logger := &TinkoffLogger{s: s, c: conn}

	// Create Tinkoff client
	ctx := context.Background()

	client, err := investgo.NewClient(ctx, cfg, logger)
	if err != nil {
		s.logger.Error("Tinkoff creating error", "error", err)
		return
	}

	conn.Lock()
	conn.tinkoffClient = client
	conn.Unlock()

	// Create Market Data client
	market := client.NewMarketDataStreamClient()

	// Create Stream
	stream, err := market.MarketDataStream()
	if err != nil {
		s.logger.Error("Tinkoff market data stream error", "error", err)
	}

	candleChan, err := stream.SubscribeCandle(conn.tinkoffConfig.Candles, pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE, true)

	if err != nil {
		s.logger.Error("Tinkoff candle subscribe error", "error", err)
	}

	// Listen to channel
	go func() {
		err := stream.Listen()
		if err != nil {
			s.logger.Error("Tinkoff stream", "error", err)
		}
	}()

	// Process candles
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case c, ok := <-candleChan:
				if !ok {
					return
				}

				rcv := time.Now()

				r := &Candle{
					MessageHeader: MessageHeader{
						Symbol:  c.Figi,
						Time:    c.Time.AsTime().UnixMicro(),
						TimeSrv: c.Time.AsTime().UnixMicro(),
						TimeRcv: rcv.UnixMicro(),
						Source:  source,
					},

					Interval: c.Interval.Enum().String(),
					Open:     fmt.Sprintf("%f", c.GetOpen().ToFloat()),
					High:     fmt.Sprintf("%f", c.GetHigh().ToFloat()),
					Low:      fmt.Sprintf("%f", c.GetLow().ToFloat()),
					Close:    fmt.Sprintf("%f", c.GetClose().ToFloat()),
					Volume:   strconv.Itoa(int(c.Volume)),
				}

				if err := s.ProcessCandle(r); err != nil {
					s.logger.Error("Error processing tinkoff candle", "conn", source)
				}
			}
		}
	}(ctx)
}
