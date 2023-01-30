package server

// ProcessCandle processes the candle.
func (s *Server) ProcessCandle(name string, c *Candle) error {
	s.NATSSend(c.NATSSubject(), c)
	s.MongoDBStore(c)
	s.InfluxDBStore(c.InfluxDBPoint())
	return nil
}

// ProcessQuote processes the quote.
func (s *Server) ProcessQuote(name string, c *Quote) error {
	s.NATSSend(c.NATSSubject(), c)
	s.MongoDBStore(c)
	s.InfluxDBStore(c.InfluxDBPoint())
	return nil
}
