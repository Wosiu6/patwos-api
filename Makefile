.PHONY: test test-unit test-integration test-coverage lint security-scan clean build run docker-build docker-run help

help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-unit         - Run unit tests"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  lint              - Run linters"
	@echo "  security-scan     - Run security scanners"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container"
	@echo "  clean             - Clean build artifacts"

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v -race -short ./middleware/... ./service/... ./repository/... ./controllers/...

test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

lint:
	@echo "Running linters..."
	golangci-lint run --timeout=5m

security-scan:
	@echo "Running security scans..."
	gosec -no-fail -fmt=json -out=gosec-report.json ./...
	govulncheck ./...

build:
	@echo "Building application..."
	go build -v -o bin/patwos-api .

run:
	@echo "Running application..."
	go run main.go

docker-build:
	@echo "Building Docker image..."
	docker build -t patwos-api:latest .

docker-run:
	@echo "Running Docker container..."
	docker-compose up -d

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f gosec-report.json
	go clean -cache -testcache

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

format:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .
