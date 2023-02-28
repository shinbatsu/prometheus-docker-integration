FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /usr/bin/prometheus-docker-integration .

FROM alpine:3.18

COPY --from=builder /usr/bin/prometheus-docker-integration /usr/bin/prometheus-docker-integration

RUN apk add --no-cache ca-certificates

CMD ["/usr/bin/prometheus-docker-integration"]
