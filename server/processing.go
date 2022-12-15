package server

// ProcessCandle processes the candle.
func (s *Server) ProcessCandle(name string, c *Candle) error {
	s.NATSSend(c.NATSSubject(), c)
	return nil
}

// ProcessQuote processes the quote.
func (s *Server) ProcessQuote(name string, c *Quote) error {
	s.NATSSend(c.NATSSubject(), c)
	return nil
}
