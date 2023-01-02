# StockMQ Server

![build status](https://github.com/stockmq/stockmq-server/actions/workflows/build.yml/badge.svg)

High-Performance message broker for the market data.

This repository provides core functionality including WebSocket connector for Binance and Tinkoff OpenAPI. 

# Requirements

NATS Server

```
GO111MODULE=on go install github.com/nats-io/nats-server/v2@latest
$(GOPATH)/bin/nats-server
```

# Start the server

Configure all required feeds in stockmq-server.xml

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
