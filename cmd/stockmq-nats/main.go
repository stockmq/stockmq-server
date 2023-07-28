package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stockmq/stockmq-server/pkg/server"
)

var (
	url     = flag.String("url", "nats://127.0.0.1:4222", "NATS URL")
	subject = flag.String("subject", "*.>", "Subject")
	debug   = flag.Bool("debug", false, "Debug messages")
)

func main() {
	// Parse flags.
	flag.Parse()

	// Connect to NATS.
	nc, err := nats.Connect(*url)
	if err != nil {
		panic(err)
	}

	defer nc.Close()

	// Simple Async Subscriber
	nc.Subscribe(*subject, func(m *nats.Msg) {
		msg := &server.MessageHeader{}

		if err := json.Unmarshal(m.Data, msg); err != nil {
			panic(err)
		}

		fmt.Printf("%s: [Server -> Broker: %5dμs] [Broker -> NATS -> Client: %5dμs]\n",
			m.Subject,
			msg.TimeRcv-msg.TimeSrv,
			time.Now().UnixMicro()-msg.TimeRcv,
		)
		if *debug {
			fmt.Printf("%s: %s\n", m.Subject, (m.Data))
		}
	})

	select {}
}
