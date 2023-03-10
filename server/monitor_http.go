package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Monitor Configuration
type MonitorConfig struct {
	Bind           string   `xml:"Bind"`
	TLS            bool     `xml:"TLS"`
	TLSCertificate string   `xml:"TLSCertificate"`
	TLSKey         string   `xml:"TLSKey"`
	Headers        []Header `xml:"Header"`
}

// DefaultMonitorConfig returns default Monitor config
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		Bind:           "127.0.0.1:9100",
		TLS:            false,
		TLSCertificate: "",
		TLSKey:         "",
	}
}

// MonitorConfig returns MonitorConfig.
func (s *Server) MonitorConfig() MonitorConfig {
	return s.ServerConfig().Monitor
}

const (
	LivezEndpoint  = "/livez"
	ReadyzEndpoint = "/readyz"
)

// ResponseHandler handles responses for monitor routes (JSONP and JSON).
func (s *Server) ResponseHandler(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	for _, header := range s.MonitorConfig().Headers {
		w.Header().Set(header.Name, header.Text)
	}

	b, err := json.Marshal(data)
	if err != nil {
		s.Errorf("Monitor: Error marshaling response to %s request: %v", r.URL, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)

	if callback := r.URL.Query().Get("callback"); callback != "" {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", callback, b)
	} else {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", b)
	}
}

// HandleLivez returns liveness check.
func (s *Server) HandleLivez(w http.ResponseWriter, r *http.Request) {
	if status := s.healthStatus(); status.Error == "" {
		s.ResponseHandler(w, r, http.StatusOK, status)
	} else {
		s.ResponseHandler(w, r, http.StatusServiceUnavailable, status)
	}
}

// HandleReadyz returns readiness check.
func (s *Server) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	v := &HealthStatus{Status: "ok"}
	s.ResponseHandler(w, r, http.StatusOK, v)
}

// StartMonitor starts the HTTP or HTTPs server if needed.
func (s *Server) StartMonitor() {
	cfg := s.MonitorConfig()
	s.Noticef("Starting Monitor on %v tls: %v", cfg.Bind, cfg.TLS)

	mux := http.NewServeMux()
	mux.HandleFunc(LivezEndpoint, s.HandleLivez)
	mux.HandleFunc(ReadyzEndpoint, s.HandleReadyz)

	srv := &http.Server{
		Addr:    cfg.Bind,
		Handler: mux,
	}

	s.mu.Lock()
	s.monitorServer = srv
	s.mu.Unlock()

	go func() {
		serve := func() error {
			if cfg.TLS {
				return srv.ListenAndServeTLS(cfg.TLSCertificate, cfg.TLSKey)
			} else {
				return srv.ListenAndServe()
			}
		}

		if err := serve(); err != nil {
			if !s.IsShutdown() {
				s.Errorf("Monitor: error starting monitor (FATAL): %v", err)

				// TODO (nusov): cancel Start() and close all open connections before exit
				os.Exit(1)
			}
		}
	}()
}
