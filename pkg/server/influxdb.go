package server

import (
	"encoding/json"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// InfluxDB Configuration.
type InfluxDBConfig struct {
	Enabled      bool   `xml:"Enabled"`
	URL          string `xml:"URL"`
	Token        string `xml:"Token"`
	Organization string `xml:"Organization"`
	Bucket       string `xml:"Bucket"`
}

// InfluxDBPointer provides a method to construct write.Point.
type InfluxDBPointer interface {
	InfluxDBPoint() *write.Point
}

// DefaultInfluxDBConfig returns default InfluxDB config.
func DefaultInfluxDBConfig() InfluxDBConfig {
	return InfluxDBConfig{
		Enabled:      false,
		URL:          "http://127.0.0.1:8086",
		Token:        "",
		Organization: "stockmq",
		Bucket:       "stockmq-data",
	}
}

// InfluxDBConfig returns InfluxDB configuration.
func (s *Server) InfluxDBConfig() InfluxDBConfig {
	return s.ServerConfig().InfluxDB
}

// StartDB starts the DB (InfluxDB) client.
func (s *Server) StartInfluxDB() {
	cfg := s.InfluxDBConfig()
	s.logger.Info("Starting InfluxDB connection", "url", cfg.URL)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.dbClient = influxdb2.NewClient(cfg.URL, cfg.Token)
	s.dbWriter = s.dbClient.WriteAPI(cfg.Organization, cfg.Bucket)

	errorsCh := s.dbWriter.Errors()

	go func() {
		for {
			select {
			case err := <-errorsCh:
				if err != nil {
					s.logger.Error("InfluxDB write error", "error", err)
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}

// InfluxDBPoint returns the Point.
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

// InfluxDBPoint returns the Point.
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

// InfluxDBStore stores the data point.
func (s *Server) InfluxDBStore(object InfluxDBPointer) {
	if s.dbWriter != nil {
		s.dbWriter.WritePoint(object.InfluxDBPoint())
	}
}
