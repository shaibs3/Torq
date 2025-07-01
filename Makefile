# Variables
BINARY_NAME=torq
BINARY_UNIX=$(BINARY_NAME)_unix
DOCKER_IMAGE=torq
DOCKER_TAG=latest
PORT=8080

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_DIR=bin

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all build clean run test deps docker-build docker-run docker-stop help

## Default: run all
all: clean deps test build

## Build the application
build:
	@echo "Building..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR) -v ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@go clean -testcache

## Run the application
run:
	@echo "Running..."
	$(GOBUILD) -o $(BINARY_DIR) -v ./...
	./$(BINARY_DIR)/$(BINARY_NAME)

## Run the application with hot reload (requires air)
dev:
	@echo "Running in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## Test the application
test:
	@echo "Testing..."
	$(GOTEST) -v ./...

## Test with coverage
test-coverage:
	@echo "Testing with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Get dependencies
deps:
	@echo "Getting dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## Install dependencies
install-deps:
	@echo "No external dependencies to install"

## Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## Run linter
lint:
	@echo "Linting..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

## Build for Linux
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/$(BINARY_UNIX) -v ./...

## Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) --name $(BINARY_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)

## Docker stop
docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

## Docker clean
docker-clean:
	@echo "Cleaning Docker images..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

## Generate API documentation (if using swagger)
docs:
	@echo "Generating API documentation..."
	@if command -v swag > /dev/null; then \
		swag init; \
	else \
		echo "swag not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init; \
	fi

## Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Installing..."; \
		curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.19.0; \
		gosec ./...; \
	fi

## Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the application"
	@echo "  dev           - Run with hot reload (requires air)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Get dependencies"
	@echo "  install-deps  - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  build-linux   - Build for Linux"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-stop   - Stop Docker container"
	@echo "  docker-clean  - Clean Docker images"
	@echo "  docs          - Generate API documentation"
	@echo "  security      - Run security scan"
	@echo "  help          - Show this help message" 