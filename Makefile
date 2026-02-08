.PHONY: build test clean install run-daemon help

# Build variables
BINARY_NAME=cloudless
BUILD_DIR=.
INSTALL_PATH=/usr/local/bin

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build        - Build the cloudless binary"
	@echo "  test         - Run all tests"
	@echo "  integration  - Run integration tests"
	@echo "  clean        - Remove built binaries and clean state"
	@echo "  install      - Install cloudless to /usr/local/bin (requires sudo)"
	@echo "  run-daemon   - Build and run the daemon in foreground"
	@echo "  lint         - Run go vet and gofmt checks"
	@echo "  fmt          - Format all Go files"

## build: Build the cloudless binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/cloudless
	@echo "Build complete: $(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running unit tests..."
	@go test -v ./...
	@echo "Unit tests complete"

## integration: Run integration tests
integration: build
	@echo "Running integration tests..."
	@./test.sh
	@echo "Integration tests complete"

## clean: Remove built binaries and clean state
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf ~/.cloudless/
	@echo "Clean complete"

## install: Install cloudless binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "Installation complete"

## run-daemon: Build and run the daemon in foreground
run-daemon: build
	@echo "Starting daemon..."
	@./$(BINARY_NAME) daemon start

## lint: Run go vet and check formatting
lint:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Checking formatting..."
	@gofmt -l . | grep -v "^$$" && echo "Files need formatting (run 'make fmt')" || echo "All files properly formatted"

## fmt: Format all Go files
fmt:
	@echo "Formatting Go files..."
	@gofmt -w .
	@echo "Formatting complete"

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"
