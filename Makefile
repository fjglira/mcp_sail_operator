# Makefile for MCP Sail Operator

.PHONY: build clean test run help

# Variables
BINARY_NAME=mcp-sail-operator
BUILD_DIR=./cmd/server
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build the MCP server binary"
	@echo "  clean      - Remove built artifacts"
	@echo "  test       - Run tests"
	@echo "  run        - Build and run the server"
	@echo "  run-config - Build and run with custom kubeconfig (set KUBECONFIG=path)"
	@echo "  tidy       - Clean up go.mod dependencies"
	@echo "  fmt        - Format Go code"
	@echo "  help       - Show this help message"

# Build the binary
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)

# Clean built artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Build and run
run: build
	@echo "Starting MCP Sail Operator server..."
	./$(BINARY_NAME)

# Build and run with custom kubeconfig
run-config: build
	@echo "Starting MCP Sail Operator server with custom kubeconfig..."
	./$(BINARY_NAME) --kubeconfig=$(KUBECONFIG)

# Tidy dependencies
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...