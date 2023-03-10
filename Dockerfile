# syntax=docker/dockerfile:1

FROM golang:1.20 as build

WORKDIR /go/src/app
COPY main.go go.mod go.sum ./
COPY pb ./pb
COPY server ./server

RUN go mod download
RUN go vet -v ./...
RUN go test -v ./...

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static

COPY --from=build /go/bin/app /

EXPOSE 9100/tcp
EXPOSE 9101/tcp

ENTRYPOINT [ "/app" ]
CMD [ "-n", "nats://nats:4222", "-m", "0.0.0.0:9100", "-g", "0.0.0.0:9101" ]
