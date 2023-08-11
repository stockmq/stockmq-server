package server

import (
	"os"
	"os/signal"
	"syscall"
)

// HandleSignals runs a goroutine to handle signals.
func (s *Server) HandleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigs:
				s.logger.Info("Signal received", "signal", sig)

				switch sig {
				case syscall.SIGINT:
					s.Shutdown()
				case syscall.SIGTERM:
					s.Shutdown()
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}
