# syntax=docker/dockerfile:1

FROM golang:1.21 as build
ARG CGO_ENABLED=0

WORKDIR /go/src/app
COPY main.go go.mod go.sum ./
COPY pb ./pb
COPY server ./server

RUN go mod download
RUN go vet -v ./...
RUN go test -v ./...

RUN CGO_ENABLED=$CGO_ENABLED go build -o /go/bin/app

FROM gcr.io/distroless/static

COPY --from=build /go/bin/app /

EXPOSE 9100/tcp
EXPOSE 9101/tcp

ENTRYPOINT [ "/app" ]
CMD [ "-h" ]
