package server

import (
	"crypto/tls"
	"net"

	"github.com/nusov/stockmq-server/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPC Configuration
type GRPCConfig struct {
	Bind           string `xml:"Bind"`
	TLS            bool   `xml:"TLS"`
	TLSCertificate string `xml:"TLSCertificate"`
	TLSKey         string `xml:"TLSKey"`
}

// StartNATS starts the NATS client.
func (s *Server) StartGRPC() {
	cfg := s.GRPCConfig()
	s.Noticef("Starting GRPC on %v tls: %v", cfg.Bind, cfg.TLS)

	grpcListener, err := net.Listen("tcp", cfg.Bind)
	if err != nil {
		s.Errorf("GRPC: cannot listen on %s: %v", cfg.Bind, err)
	}

	opts := []grpc.ServerOption{}

	if cfg.TLS {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSKey)
		if err != nil {
			s.Errorf("GRPC: cannot load TLS certificate: %v", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMonitorServer(grpcServer, &Backend{s: s})

	s.mu.Lock()
	s.grpcListener = grpcListener
	s.grpcServer = grpcServer
	s.mu.Unlock()

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			if !s.IsShutdown() {
				s.Errorf("GRPC: error serving: %v", err)
			}
		}
	}()
}
