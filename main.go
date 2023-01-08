package main

import (
	"encoding/xml"
	"flag"
	"os"

	"github.com/stockmq/stockmq-server/server"
)

func main() {
	// Get default config.
	cfg := server.DefaultConfig()
	cfn := ""

	// Parse flags.
	flag.StringVar(&cfn, "c", "", "Configuration file (XML)")
	flag.StringVar(&cfg.NATS.URL, "n", "nats://127.0.0.1:4222", "NATS URL")
	flag.StringVar(&cfg.Monitor.Bind, "m", "127.0.0.1:9100", "Monitor bind address")
	flag.StringVar(&cfg.GRPC.Bind, "g", "127.0.0.1:9101", "gRPC bind address")
	flag.BoolVar(&cfg.Logger.Debug, "d", false, "Enable Debug messages")
	flag.Parse()

	// Read the configuration file and override defaults.
	if cfn != "" {
		b, err := os.ReadFile(cfn)
		if err != nil {
			panic(err)
		}
		if err := xml.Unmarshal(b, &cfg); err != nil {
			panic(err)
		}
	}

	// Parse flags again to preserve the precedence.
	flag.Parse()

	// Create the server.
	s, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	// Start the server.
	if err := s.Start(); err != nil {
		panic(err)
	}

	// Wait until shutdown channel will be closed.
	s.WaitForShutdown()
}
