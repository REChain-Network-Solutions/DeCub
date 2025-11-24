.PHONY: all build test clean docker-build docker-run docker-stop deps proto

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main binary
BINARY_NAME=rechain
BINARY_UNIX=$(BINARY_NAME)_unix

# CLI binary
CLI_BINARY_NAME=rechainctl
CLI_BINARY_UNIX=$(CLI_BINARY_NAME)_unix

# Build directories
BUILD_DIR=./build
DIST_DIR=./dist

all: test build

deps:
	$(GOMOD) download
	$(GOMOD) tidy

build: deps
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/rechain
	$(GOBUILD) -o $(BUILD_DIR)/$(CLI_BINARY_NAME) -v ./cmd/rechainctl

build-linux: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY_UNIX) -v ./cmd/rechain
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(CLI_BINARY_UNIX) -v ./cmd/rechainctl

test:
	$(GOTEST) -v ./...

test-integration:
	$(GOTEST) -v -tags=integration ./tests/...

test-coverage:
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

bench:
	$(GOTEST) -bench=. -benchmem ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) --config ./config/config.yaml

run-dev: build
	./$(BUILD_DIR)/$(BINARY_NAME) --config ./config/config.dev.yaml

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./api/proto/rechain.proto

docker-build:
	docker build -t rechain:latest .

docker-run: docker-build
	docker run -p 1317:1317 -p 9090:9090 -p 26656:26656 -p 9091:9091 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/logs:/app/logs \
		-v $(PWD)/config:/app/config \
		rechain:latest

docker-stop:
	docker stop $(shell docker ps -q --filter ancestor=rechain:latest)

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-compose-logs:
	docker-compose logs -f

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@mkdir -p data logs certs
	@cp config/config.yaml config/config.dev.yaml
	@echo "Development environment ready!"

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	# Add mock generation commands here

# Linting
lint:
	golangci-lint run

# Formatting
fmt:
	$(GOCMD) fmt ./...

# Security scanning
security:
	gosec ./...

# Dependency updates
update-deps:
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Release build
release: clean test build-linux
	@echo "Release build complete!"
	@ls -la $(DIST_DIR)/

# Install development tools
install-tools:
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOCMD) install github.com/cosmtrek/air@latest

# Help
help:
	@echo "Available targets:"
	@echo "  all              - Run tests and build"
	@echo "  build            - Build binaries"
	@echo "  build-linux      - Build for Linux"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  bench            - Run benchmarks"
	@echo "  clean            - Clean build artifacts"
	@echo "  run              - Run the application"
	@echo "  run-dev          - Run in development mode"
	@echo "  proto            - Generate protobuf files"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  docker-stop      - Stop Docker container"
	@echo "  docker-compose-up    - Start with docker-compose"
	@echo "  docker-compose-down  - Stop docker-compose"
	@echo "  docker-compose-logs  - Show docker-compose logs"
	@echo "  dev-setup        - Setup development environment"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  security         - Run security scanner"
	@echo "  update-deps      - Update dependencies"
	@echo "  release          - Create release build"
	@echo "  install-tools    - Install development tools"
	@echo "  help             - Show this help"
