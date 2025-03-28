FROM golang:alpine AS builder

WORKDIR /app_parser

COPY main.go .

RUN go mod init mod

RUN go get github.com/prometheus/client_golang/prometheus
RUN go get github.com/prometheus/client_golang/prometheus/promhttp@v1.20.5

RUN apk add --no-cache sudo

RUN go build -o parser main.go

FROM alpine

WORKDIR /app_parser

COPY --from=builder /app_parser/parser .

RUN apk add --no-cache sudo

CMD ["./parser"]

