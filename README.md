# StockMQ Server
High-Performance message broker for the market data

# Installation

```
GO111MODULE=on go install github.com/nats-io/nats-server/v2@latest
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
