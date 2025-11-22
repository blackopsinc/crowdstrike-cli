# Makefile for cross-compiling crowdstrike-cli

BINARY_NAME=crowdstrike-cli
SOURCE_FILE=crowdstrike-cli.go

# Build for all platforms
.PHONY: all
all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-386

# Build for Linux AMD64
.PHONY: linux-amd64
linux-amd64:
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(SOURCE_FILE)
	@chmod +x bin/$(BINARY_NAME)-linux-amd64

# Build for Linux ARM64
.PHONY: linux-arm64
linux-arm64:
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 $(SOURCE_FILE)
	@chmod +x bin/$(BINARY_NAME)-linux-arm64

# Build for macOS AMD64 (Intel)
.PHONY: darwin-amd64
darwin-amd64:
	@echo "Building for macOS AMD64 (Intel)..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(SOURCE_FILE)
	@chmod +x bin/$(BINARY_NAME)-darwin-amd64

# Build for macOS ARM64 (Apple Silicon)
.PHONY: darwin-arm64
darwin-arm64:
	@echo "Building for macOS ARM64 (Apple Silicon)..."
	@GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 $(SOURCE_FILE)
	@chmod +x bin/$(BINARY_NAME)-darwin-arm64

# Build for Windows AMD64 (64-bit)
.PHONY: windows-amd64
windows-amd64:
	@echo "Building for Windows AMD64 (64-bit)..."
	@GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_FILE)

# Build for Windows 386 (32-bit)
.PHONY: windows-386
windows-386:
	@echo "Building for Windows 386 (32-bit)..."
	@GOOS=windows GOARCH=386 go build -o bin/$(BINARY_NAME)-windows-386.exe $(SOURCE_FILE)

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	@go build -o bin/$(BINARY_NAME) $(SOURCE_FILE)
	@chmod +x bin/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# Create bin directory if it doesn't exist
bin:
	@mkdir -p bin

# Ensure bin directory exists before building
$(BINARY_NAME)-linux-amd64 $(BINARY_NAME)-linux-arm64 $(BINARY_NAME)-darwin-amd64 $(BINARY_NAME)-darwin-arm64 $(BINARY_NAME)-windows-amd64.exe $(BINARY_NAME)-windows-386.exe: bin

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Build for all platforms (Linux, macOS, Windows)"
	@echo "  linux-amd64      - Build for Linux AMD64"
	@echo "  linux-arm64      - Build for Linux ARM64"
	@echo "  darwin-amd64     - Build for macOS AMD64 (Intel)"
	@echo "  darwin-arm64     - Build for macOS ARM64 (Apple Silicon)"
	@echo "  windows-amd64    - Build for Windows AMD64 (64-bit)"
	@echo "  windows-386      - Build for Windows 386 (32-bit)"
	@echo "  build            - Build for current platform"
	@echo "  clean            - Remove build artifacts"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make all              # Build for all platforms"
	@echo "  make windows-amd64    # Build for Windows 64-bit"
	@echo "  make linux-amd64      # Build for Linux 64-bit"
	@echo "  make darwin-arm64     # Build for macOS Apple Silicon"

