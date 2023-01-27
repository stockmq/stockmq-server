package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"

	"github.com/stockmq/stockmq-server/pb"
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

// DefaultGRPCConfig returns default GRPC config
func DefaultGRPCConfig() GRPCConfig {
	return GRPCConfig{
		Bind:           "127.0.0.1:9101",
		TLS:            false,
		TLSCertificate: "",
		TLSKey:         "",
	}
}

// GRPCConfig returns GRPC configuration.
func (s *Server) GRPCConfig() GRPCConfig {
	return s.ServerConfig().GRPC
}

// StartGRPC starts the GRPC server.
func (s *Server) StartGRPC() error {
	cfg := s.GRPCConfig()
	s.Noticef("Starting GRPC on %v tls: %v", cfg.Bind, cfg.TLS)

	grpcListener, err := net.Listen("tcp", cfg.Bind)
	if err != nil {
		return fmt.Errorf("GRPC: cannot listen on %s: %v", cfg.Bind, err)
	}

	opts := []grpc.ServerOption{}

	if cfg.TLS {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSKey)
		if err != nil {
			return fmt.Errorf("GRPC: cannot load TLS certificate: %v", err)
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

				// TODO (nusov): cancel Start() and close all open connections before exit
				os.Exit(1)
			}
		}
	}()

	return nil
}
