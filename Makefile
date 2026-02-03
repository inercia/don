.PHONY: build clean test run lint lint-golangci format help

# Binary name
BINARY_NAME=don
# Build directory
BUILD_DIR=build

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)

# Default target
all: build

# Build the application
build:
	@echo ">>> Building $(BINARY_NAME)..."
	@mkdir -p $(GOBIN)
	@go build -o $(GOBIN)/$(BINARY_NAME) .
	@echo ">>> ... $(BINARY_NAME) built successfully"

# Clean build artifacts
clean:
	@echo ">>> Cleaning..."
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo ">>> Running tests..."
	@go test -v ./...
	@echo ">>> ... tests completed successfully"

# Run the exe command tests
test-e2e:
	@echo ">>> Running end-to-end tests..."
	@if [ ! -x "$(GOBIN)/$(BINARY_NAME)" ]; then \
		echo ">>> $(GOBIN)/$(BINARY_NAME) not found. Building..."; \
		$(MAKE) build; \
	fi
	@chmod +x tests/*.sh 2>/dev/null || true
	@tests/run_tests.sh
	@echo ">>> ... end-to-end tests completed"

# Run the application
run:
	@go run main.go

# Install the application
install:
	@echo ">>> Installing $(BINARY_NAME)..."
	@go install .
	@echo ">>> ... $(BINARY_NAME) installed successfully"

# Run linting (golangci-lint)
lint: lint-golangci

# Run golangci-lint (comprehensive linting)
lint-golangci:
	@echo ">>> Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi
	@echo ">>> ... golangci-lint completed successfully"

# Format code
format:
	@echo ">>> Formatting Go code..."
	@go fmt ./...
	@go mod tidy
	@echo ">>> ... code formatted successfully"

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  clean              - Remove build artifacts"
	@echo "  test               - Run tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  run                - Run the application"
	@echo "  install            - Install the application"
	@echo "  lint               - Run linting (alias for lint-golangci)"
	@echo "  lint-golangci      - Run golangci-lint (installs if not present)"
	@echo "  format             - Format Go code"
	@echo "  help               - Show this help"

