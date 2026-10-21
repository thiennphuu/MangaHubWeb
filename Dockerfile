FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the API server by default (other binaries can be added similarly)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/api-server ./cmd/api-server
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/tcp-server ./cmd/tcp-server
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/udp-server ./cmd/udp-server
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/grpc-server ./cmd/grpc-server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates
RUN mkdir -p /app/data

COPY --from=builder /bin/api-server /bin/api-server
COPY --from=builder /bin/tcp-server /bin/tcp-server
COPY --from=builder /bin/udp-server /bin/udp-server
COPY --from=builder /bin/grpc-server /bin/grpc-server
COPY config.yaml ./config.yaml
# Default ports are configured in config.yaml; environment variables (e.g., DB_PATH, TCP_PORT, UDP_PORT, GRPC_PORT, WS_PORT) can override per container.

# Default command can be overridden per-service in docker-compose
CMD ["/bin/api-server"]



