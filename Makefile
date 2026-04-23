.PHONY: build run run-http test clean lint help

# Build the server binary
build:
	@echo "Building Allure MCP server..."
	go build -o bin/server ./cmd/server
	@echo "✓ Server built successfully at bin/server"

# Run server in stdio mode (default)
run: build
	@echo "Running server in stdio mode..."
	@echo "Make sure environment variables are set:"
	@echo "  ALLURE_BASE_URL, ALLURE_TOKEN"
	@./run-server.sh

# Run server in HTTP mode (for testing)
run-http: build
	@echo "Running server in HTTP mode on :3000..."
	@echo "Loading environment from .env..."
	@. ./.env && ./bin/server --http

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "✓ All tests passed"

# Run linting
lint:
	@echo "Running linters..."
	go vet ./...
	@echo "✓ No issues found"

# Build and run linting
check: lint test
	@echo "✓ All checks passed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "✓ Clean complete"

# Full rebuild
rebuild: clean build
	@echo "✓ Rebuild complete"

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✓ Formatting complete"

# Display configuration
info:
	@echo "Project: Allure MCP Server"
	@echo "Module: github.com/MimoJanra/TestOpsMCP"
	@echo "Go version required: 1.22+"
	@echo ""
	@echo "Commands:"
	@echo "  make build       - Build the server binary"
	@echo "  make run         - Run server in stdio mode (local development)"
	@echo "  make run-http    - Run server in HTTP mode (port 3000)"
	@echo "  make test        - Run unit tests"
	@echo "  make lint        - Run Go linter"
	@echo "  make check       - Run lint + tests"
	@echo "  make fmt         - Format Go code"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make rebuild     - Clean and rebuild"
	@echo "  make help        - Show this message"

# Default target
help: info
