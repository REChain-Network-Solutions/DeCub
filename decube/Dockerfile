FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o decube ./cmd/decube

FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create decube user
RUN addgroup -g 1000 decube && \
    adduser -D -s /bin/sh -u 1000 -G decube decube

WORKDIR /var/lib/decube

# Copy the binary from builder stage
COPY --from=builder /app/decube /usr/local/bin/decube

# Copy configuration files
COPY --from=builder /app/config ./config

# Create necessary directories
RUN mkdir -p /var/lib/decube/etcd /var/lib/decube/etcd/wal /var/log/decube && \
    chown -R decube:decube /var/lib/decube /var/log/decube

USER decube

# Expose ports
EXPOSE 2379 2380 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["decube", "--config", "./config/config.yaml"]
