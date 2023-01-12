# StockMQ Server

![build status](https://github.com/stockmq/stockmq-server/actions/workflows/build.yml/badge.svg)
![build status](https://github.com/stockmq/stockmq-server/actions/workflows/docker-build.yml/badge.svg)


High-Performance message broker for the market data.

This repository provides core functionality including WebSocket connector for Binance and Tinkoff OpenAPI. 

# Requirements

NATS Server

```
GO111MODULE=on go install github.com/nats-io/nats-server/v2@latest
$(GOPATH)/bin/nats-server
```

# Example configuration

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Config>
    <WebSocket>
        <Name>Binance-BTCUSD</Name>
        <Enabled>true</Enabled>
        <URL>wss://stream.binance.com:9443/ws</URL>
        <Handler>Binance</Handler>
        <DialTimeout>4</DialTimeout>
        <RetryDelay>3</RetryDelay>
        <PingTimeout>60</PingTimeout>
        <ReadLimit>655350</ReadLimit>
        <InitMessage>{"id": 0, "method": "SUBSCRIBE", "params": ["btcusdt@kline_1s", "btcusdt@depth"]}</InitMessage>
    </WebSocket>
</Config>
```

# Start the server

Configure all required feeds in stockmq-server.xml

```
go build
./stockmq-server -c stockmq-server.xml
```

# Listen for updates

```
cd cmd/stockmq-nats
go build
./stockmq-nats -debug
```
