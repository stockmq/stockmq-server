package server

// ProcessCandle processes the candle.
func (s *Server) ProcessCandle(c *Candle) error {
	s.NATSSend(c)
	s.MongoDBStore(c)
	s.InfluxDBStore(c)
	return nil
}

// ProcessQuote processes the quote.
func (s *Server) ProcessQuote(c *Quote) error {
	s.NATSSend(c)
	s.MongoDBStore(c)
	s.InfluxDBStore(c)
	return nil
}
