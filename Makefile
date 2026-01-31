# Makefile for kubegrid

.PHONY: build clean test run install lint fmt help

BINARY_NAME=kubegrid
BINARY_PATH=./$(BINARY_NAME)
MAIN_PATH=./cmd/kubegrid
INSTALL_PATH=$(HOME)/.local/bin

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Build with optimizations for production
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	@go build -ldflags="-s -w" -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Production build complete: $(BINARY_PATH)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_PATH)
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Install to user bin directory
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_PATH) $(INSTALL_PATH)/
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

# Uninstall from user bin directory
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled"

# Run linters
lint:
	@echo "Running linters..."
	@go vet ./...
	@gofmt -l .
	@echo "Lint complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Tidy complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

# Run in development mode with race detector
dev:
	@echo "Running in development mode..."
	@go run -race $(MAIN_PATH)

# Generate mocks for testing (requires mockgen)
mocks:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated"

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth
	@echo "Security check complete"

# Cross-compile for multiple platforms
build-all: clean
	@echo "Cross-compiling for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Cross-compilation complete"

# Display help
help:
	@echo "kubegrid - Makefile commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the application"
	@echo "  build-prod    Build with production optimizations"
	@echo "  clean         Remove build artifacts"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  run           Build and run the application"
	@echo "  install       Install to ~/.local/bin"
	@echo "  uninstall     Remove from ~/.local/bin"
	@echo "  lint          Run linters"
	@echo "  fmt           Format code"
	@echo "  tidy          Tidy dependencies"
	@echo "  deps          Download dependencies"
	@echo "  dev           Run with race detector"
	@echo "  mocks         Generate test mocks"
	@echo "  security      Check for vulnerabilities"
	@echo "  build-all     Cross-compile for all platforms"
	@echo "  help          Display this help message"
