.PHONY: build test clean run docker-build docker-up docker-down help

# Build the application
build:
	go build -o products ./cmd/api

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f products
	rm -f coverage.out

# Run the application locally
run: build
	./products

# Build Docker image
docker-build:
	docker build -t products-api .

# Start services with Docker Compose
docker-up:
	docker-compose up --build

# Start services in background
docker-up-d:
	docker-compose up --build -d

# Stop services
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy
	go mod download

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Build and run the application"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start services with Docker Compose"
	@echo "  docker-up-d    - Start services in background"
	@echo "  docker-down    - Stop services"
	@echo "  docker-logs    - View service logs"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  deps           - Install dependencies"
	@echo "  help           - Show this help message"
