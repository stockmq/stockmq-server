package server

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// DB Configuration.
type DBConfig struct {
	URL          string `xml:"URL"`
	Token        string `xml:"Token"`
	Organization string `xml:"Organization"`
	Bucket       string `xml:"Bucket"`
}

// DefaultDBConfig returns default InfluxDB config.
func DefaultDBConfig() DBConfig {
	return DBConfig{
		URL:          "http://127.0.0.1:8086",
		Token:        "",
		Organization: "stockmq",
		Bucket:       "stockmq-data",
	}
}

// DBConfig returns DB configuration.
func (s *Server) DBConfig() DBConfig {
	return s.ServerConfig().DB
}

// StartDB starts the DB (InfluxDB) client.
func (s *Server) StartDB() {
	cfg := s.DBConfig()
	s.Noticef("Starting DB connection to %s", cfg.URL)

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
					s.Errorf("DB write error: %v", err)
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}
