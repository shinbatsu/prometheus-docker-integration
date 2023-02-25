FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go install -ldflags="-s -w" .

FROM alpine:3.18 AS final

COPY --from=builder /go/bin/prometheus-docker-integration /usr/bin/prometheus-docker-integration

RUN apk add --no-cache ca-certificates

EXPOSE 9476

CMD ["/usr/bin/prometheus-docker-integration"]
