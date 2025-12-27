# Multi-stage build for DeCube
FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
COPY */go.mod */go.sum ./
COPY decub-*/go.mod decub-*/go.mod ./
COPY rechain/go.mod rechain/go.mod ./
COPY decube/go.mod decube/go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 decube && \
    adduser -D -u 1000 -G decube decube

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bin/* /usr/local/bin/

# Copy configuration
COPY config/config.example.yaml /etc/decube/config.yaml

# Create data directory
RUN mkdir -p /var/lib/decube && \
    chown -R decube:decube /var/lib/decube /etc/decube

# Switch to non-root user
USER decube

# Expose ports
EXPOSE 8080 9090 7000 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["decube", "--config", "/etc/decube/config.yaml"]

