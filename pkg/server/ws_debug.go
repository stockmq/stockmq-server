package server

// WSDebugHandler just prints received message.
func WSDebugHandler(s *Server, w *WSConnection, msg []byte) error {
	s.logger.Info("WSS", "name", w.wsConfig.Name, "message", msg)
	return nil
}

func init() {
	Handlers["Debug"] = WSDebugHandler
}
