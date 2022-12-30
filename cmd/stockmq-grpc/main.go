package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stockmq/stockmq-server/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	addr   = flag.String("addr", "localhost:9101", "the address to connect to")
	ca     = flag.String("ca", "../../certs/root.pem", "CA certificate")
	useTLS = flag.Bool("useTLS", false, "Enable TLS")
)

// loadTLSCredentials loads certificate of the CA who signed server's certificate.
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(*ca)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	// Parse flags
	flag.Parse()

	opts := []grpc.DialOption{}
	if *useTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMonitorClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.IsRunning(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("could not perform call: %v", err)
	}
	log.Printf("IsRunning: %v", r.Value)
}
