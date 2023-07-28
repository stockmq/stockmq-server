package server

// WSDebugHandler just prints received message.
func WSDebugHandler(s *Server, w *WSConnection, msg []byte) error {
	s.Noticef("WSS %s: %s", w.wsConfig.Name, msg)
	return nil
}

func init() {
	Handlers["Debug"] = WSDebugHandler
}
