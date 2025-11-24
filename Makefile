.PHONY: build run test clean docker-build docker-run

# Build all components
build:
	cd decub-control-plane && go mod tidy && go build -o ../bin/control-plane
	cd decub-gcl/go && go mod tidy && go build -o ../../bin/gcl-go
	cd decub-gossip && go mod tidy && go build -o ../bin/gossip
	cd decub-cas && go mod tidy && go build -o ../bin/cas
	cd decub-catalog && go mod tidy && go build -o ../bin/catalog

# Run all services with docker-compose
run:
	docker-compose up -d

# Run tests
test:
	cd decub-control-plane && go test ./...
	cd decub-gcl/go && go test ./...
	cd decub-gossip && go test ./...
	cd decub-cas && go test ./...
	cd decub-catalog && go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	docker-compose down -v

# Build Docker images
docker-build:
	docker-compose build

# Run with Docker
docker-run:
	docker-compose up

# Stop all services
stop:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@echo "Make sure you have Go 1.19+, Docker, and Docker Compose installed"
	@echo "Run 'make build' to build all components"
	@echo "Run 'make run' to start all services"
