package server

// Server Configuration.
type ServerConfig struct {
	Logger    LoggerConfig   `xml:"Logger"`
	Monitor   MonitorConfig  `xml:"Monitor"`
	MongoDB   MongoDBConfig  `xml:"MongoDB"`
	InfluxDB  InfluxDBConfig `xml:"InfluxDB"`
	NATS      NATSConfig     `xml:"NATS"`
	GRPC      GRPCConfig     `xml:"GRPC"`
	WebSocket []WSConfig     `xml:"WebSocket"`
}

// DefaultConfig returns default ServerConfig.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Logger:   DefaultLoggerConfig(),
		Monitor:  DefaultMonitorConfig(),
		MongoDB:  DefaultMongoDBConfig(),
		InfluxDB: DefaultInfluxDBConfig(),
		NATS:     DefaultNATSConfig(),
		GRPC:     DefaultGRPCConfig(),
	}
}

// Config returns a copy of Server configuration.
func (s *Server) ServerConfig() ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}
