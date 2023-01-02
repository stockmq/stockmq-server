# StockMQ Server
High-Performance message broker for the market data.

This repository provides core functionality including WebSocket connector for Binance and Tinkoff OpenAPI. 

# Requirements

NATS Server

```
GO111MODULE=on go install github.com/nats-io/nats-server/v2@latest
$(GOPATH)/bin/nats-server
```

# Start the server

```
go build
./stockmq-server -c stockmq-server.xml
```

# Listen to NATS

```
cd cmd/stockmq-nats
go build
./stockmq-nats -debug
```
