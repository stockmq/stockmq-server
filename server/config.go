package server

// Header represents HTTP header
type Header struct {
	Name string `xml:"Name,attr"`
	Text string `xml:",chardata"`
}

// WebSocket Configuration
type WSConfig struct {
	Name         string   `xml:"Name"`
	Enabled      bool     `xml:"Enabled"`
	URL          string   `xml:"URL"`
	Handler      string   `xml:"Handler"`
	DialTimeout  int      `xml:"DialTimeout"`
	RetryDelay   int      `xml:"RetryDelay"`
	PingTimeout  int      `xml:"PingTimeout"`
	ReadLimit    int64    `xml:"ReadLimit"`
	Headers      []Header `xml:"Header"`
	InitMessages []string `xml:"InitMessage"`
}

// Server Configuration
type ServerConfig struct {
	Logger    LoggerConfig  `xml:"Logger"`
	Monitor   MonitorConfig `xml:"Monitor"`
	NATS      NATSConfig    `xml:"NATS"`
	WebSocket []WSConfig    `xml:"WebSocket"`
}

// DefaultConfig returns default ServerConfig.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Logger:  DefaultLoggerConfig(),
		Monitor: DefaultMonitorConfig(),
		NATS:    DefaultNATSConfig(),
	}
}

// Config returns a copy of Server configuration.
func (s *Server) ServerConfig() ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// LoggerConfig returns Logger configuration.
func (s *Server) LoggerConfig() LoggerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Logger
}

// NATSConfig returns NATS configuration.
func (s *Server) NATSConfig() NATSConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.NATS
}
