.PHONY: build build-all install uninstall clean clean-all rebuild test fmt vet check help

# Binary name derived from current directory
BINARY_NAME=$(shell basename $$(pwd))

# Detect current platform
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
CURRENT_PLATFORM=$(GOOS)-$(GOARCH)

# Auto-detect project structure
HAS_SRC_DIR=$(shell [ -d src ] && echo "yes" || echo "no")
HAS_CMD_DIR=$(shell [ -d cmd ] && echo "yes" || echo "no")

# Set directories based on project structure
ifeq ($(HAS_SRC_DIR),yes)
	SRC_DIR=src
	CMD_PATH=$(SRC_DIR)
	BUILD_DIR=bin
	GO_MOD_PATH=$(SRC_DIR)/go.mod
	GO_SUM_PATH=$(SRC_DIR)/go.sum
else ifeq ($(HAS_CMD_DIR),yes)
	SRC_DIR=.
	CMD_PATH=./cmd/$(BINARY_NAME)
	BUILD_DIR=bin
	GO_MOD_PATH=go.mod
	GO_SUM_PATH=go.sum
else
	SRC_DIR=.
	CMD_PATH=.
	BUILD_DIR=bin
	GO_MOD_PATH=go.mod
	GO_SUM_PATH=go.sum
endif

# Platform-specific binary names
BINARY_LINUX=$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
BINARY_DARWIN_INTEL=$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
BINARY_DARWIN_ARM=$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
CURRENT_BINARY=$(BUILD_DIR)/$(BINARY_NAME)-$(CURRENT_PLATFORM)
LAUNCHER_SCRIPT=$(BUILD_DIR)/$(BINARY_NAME).sh

# Build for current platform only
build: $(CURRENT_BINARY)

# Build for all platforms and create launcher script
build-all: $(BINARY_LINUX) $(BINARY_DARWIN_INTEL) $(BINARY_DARWIN_ARM) $(LAUNCHER_SCRIPT)

rebuild: clean-all build

# Build targets for each platform
$(BINARY_LINUX): $(GO_SUM_PATH)
	@echo "Building $(BINARY_NAME) for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 go build -o ../$(BINARY_LINUX) .
else
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) $(CMD_PATH)
endif
	@echo "✓ Built: $(BINARY_LINUX)"

$(BINARY_DARWIN_INTEL): $(GO_SUM_PATH)
	@echo "Building $(BINARY_NAME) for macOS Intel (AMD64)..."
	@mkdir -p $(BUILD_DIR)
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=amd64 go build -o ../$(BINARY_DARWIN_INTEL) .
else
	@GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DARWIN_INTEL) $(CMD_PATH)
endif
	@echo "✓ Built: $(BINARY_DARWIN_INTEL)"

$(BINARY_DARWIN_ARM): $(GO_SUM_PATH)
	@echo "Building $(BINARY_NAME) for macOS Apple Silicon (ARM64)..."
	@mkdir -p $(BUILD_DIR)
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=arm64 go build -o ../$(BINARY_DARWIN_ARM) .
else
	@GOOS=darwin GOARCH=arm64 go build -o $(BINARY_DARWIN_ARM) $(CMD_PATH)
endif
	@echo "✓ Built: $(BINARY_DARWIN_ARM)"

# Create launcher script
$(LAUNCHER_SCRIPT): $(BINARY_LINUX) $(BINARY_DARWIN_INTEL) $(BINARY_DARWIN_ARM)
	@echo "Creating launcher script..."
	@mkdir -p $(BUILD_DIR)
	@echo '#!/bin/bash' > $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Auto-generated launcher script for $(BINARY_NAME)' >> $(LAUNCHER_SCRIPT)
	@echo '# Detects platform and executes the correct binary' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Get the directory where this script is located' >> $(LAUNCHER_SCRIPT)
	@echo 'SCRIPT_DIR="$$(cd "$$(dirname "$${BASH_SOURCE[0]}")" && pwd)"' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Detect OS' >> $(LAUNCHER_SCRIPT)
	@echo 'OS=$$(uname -s | tr "[:upper:]" "[:lower:]")' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Detect architecture' >> $(LAUNCHER_SCRIPT)
	@echo 'ARCH=$$(uname -m)' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Map architecture names to Go convention' >> $(LAUNCHER_SCRIPT)
	@echo 'case "$$ARCH" in' >> $(LAUNCHER_SCRIPT)
	@echo '    x86_64)' >> $(LAUNCHER_SCRIPT)
	@echo '        ARCH="amd64"' >> $(LAUNCHER_SCRIPT)
	@echo '        ;;' >> $(LAUNCHER_SCRIPT)
	@echo '    aarch64)' >> $(LAUNCHER_SCRIPT)
	@echo '        ARCH="arm64"' >> $(LAUNCHER_SCRIPT)
	@echo '        ;;' >> $(LAUNCHER_SCRIPT)
	@echo '    arm64)' >> $(LAUNCHER_SCRIPT)
	@echo '        ARCH="arm64"' >> $(LAUNCHER_SCRIPT)
	@echo '        ;;' >> $(LAUNCHER_SCRIPT)
	@echo '    *)' >> $(LAUNCHER_SCRIPT)
	@echo '        echo "Unsupported architecture: $$ARCH" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '        exit 1' >> $(LAUNCHER_SCRIPT)
	@echo '        ;;' >> $(LAUNCHER_SCRIPT)
	@echo 'esac' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Construct binary name' >> $(LAUNCHER_SCRIPT)
	@echo 'BINARY="$$SCRIPT_DIR/$(BINARY_NAME)-$$OS-$$ARCH"' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Check if binary exists' >> $(LAUNCHER_SCRIPT)
	@echo 'if [ ! -f "$$BINARY" ]; then' >> $(LAUNCHER_SCRIPT)
	@echo '    echo "Error: Binary not found for platform $$OS-$$ARCH" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '    echo "Expected: $$BINARY" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '    echo "" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '    echo "Available binaries:" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '    ls -1 "$$SCRIPT_DIR"/$(BINARY_NAME)-* 2>/dev/null | sed "s|^|  |" >&2' >> $(LAUNCHER_SCRIPT)
	@echo '    exit 1' >> $(LAUNCHER_SCRIPT)
	@echo 'fi' >> $(LAUNCHER_SCRIPT)
	@echo '' >> $(LAUNCHER_SCRIPT)
	@echo '# Execute the binary with all arguments' >> $(LAUNCHER_SCRIPT)
	@echo 'exec "$$BINARY" "$$@"' >> $(LAUNCHER_SCRIPT)
	@chmod +x $(LAUNCHER_SCRIPT)
	@echo "✓ Created launcher script: $(LAUNCHER_SCRIPT)"

# Generate go.sum
$(GO_SUM_PATH): $(GO_MOD_PATH)
	@echo "Downloading dependencies..."
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && go mod download
	@cd $(SRC_DIR) && go mod tidy
	@touch $(GO_SUM_PATH)
else
	@go mod download
	@go mod tidy
	@touch $(GO_SUM_PATH)
endif
	@echo "Dependencies downloaded"

# Generate go.mod (only if it doesn't exist)
$(GO_MOD_PATH):
	@echo "Initializing Go module..."
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && go mod init $(BINARY_NAME)
else
	@go mod init $(BINARY_NAME)
endif

# Install binary (installs the current platform binary)
install: build
	@if [ ! -f "$(CURRENT_BINARY)" ]; then \
		echo "Error: Binary for current platform ($(CURRENT_PLATFORM)) not found"; \
		echo "Run 'make build' or 'make build-all' first"; \
		exit 1; \
	fi
ifndef TARGET
	@echo "Installing $(BINARY_NAME) ($(CURRENT_PLATFORM)) to /usr/local/bin..."
	@sudo cp $(CURRENT_BINARY) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete!"
else
	@echo "Installing $(BINARY_NAME) ($(CURRENT_PLATFORM)) to $(TARGET)..."
	@cp $(CURRENT_BINARY) $(TARGET)/$(BINARY_NAME) 2>/dev/null || sudo cp $(CURRENT_BINARY) $(TARGET)/$(BINARY_NAME)
	@echo "Installation complete!"
endif

# Install launcher script (for multi-platform distribution)
install-launcher: build-all
ifndef TARGET
	@echo "Installing launcher script to /usr/local/bin/$(BINARY_NAME)..."
	@sudo cp $(LAUNCHER_SCRIPT) /usr/local/bin/$(BINARY_NAME)
	@echo "Installing platform binaries to /usr/local/lib/$(BINARY_NAME)/..."
	@sudo mkdir -p /usr/local/lib/$(BINARY_NAME)
	@sudo cp $(BINARY_LINUX) /usr/local/lib/$(BINARY_NAME)/
	@sudo cp $(BINARY_DARWIN_INTEL) /usr/local/lib/$(BINARY_NAME)/
	@sudo cp $(BINARY_DARWIN_ARM) /usr/local/lib/$(BINARY_NAME)/
	@echo "Installation complete!"
else
	@echo "Installing launcher script to $(TARGET)/$(BINARY_NAME)..."
	@cp $(LAUNCHER_SCRIPT) $(TARGET)/$(BINARY_NAME) 2>/dev/null || sudo cp $(LAUNCHER_SCRIPT) $(TARGET)/$(BINARY_NAME)
	@echo "Note: Platform binaries remain in $(BUILD_DIR)/"
	@echo "Installation complete!"
endif

# Uninstall binary
uninstall:
	@echo "Looking for $(BINARY_NAME) in system..."
	@BINARY_PATH=$$(which $(BINARY_NAME) 2>/dev/null); \
	if [ -z "$$BINARY_PATH" ]; then \
		echo "$(BINARY_NAME) not found in PATH"; \
		exit 0; \
	fi; \
	if [ -f "$$BINARY_PATH" ]; then \
		if [ "$$(basename $$(dirname $$BINARY_PATH))" = "bin" ]; then \
			echo "Found $(BINARY_NAME) at $$BINARY_PATH"; \
			echo "Removing..."; \
			sudo rm -f "$$BINARY_PATH"; \
			if [ -d "/usr/local/lib/$(BINARY_NAME)" ]; then \
				echo "Removing platform binaries..."; \
				sudo rm -rf "/usr/local/lib/$(BINARY_NAME)"; \
			fi; \
			echo "Uninstallation complete!"; \
		else \
			echo "$(BINARY_NAME) found at $$BINARY_PATH but not in a standard bin directory"; \
			echo "Please remove it manually if needed"; \
		fi; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Clean all (including go.mod and go.sum)
clean-all: clean
	@echo "Cleaning go.mod & go.sum..."
	@rm -f $(GO_MOD_PATH) $(GO_SUM_PATH)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && go test -v ./...
else
	@go test -v ./...
endif

# Format code
fmt:
	@echo "Formatting code..."
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && go fmt ./...
else
	@go fmt ./...
endif
	@echo "Format complete!"

# Run go vet
vet:
	@echo "Running go vet..."
ifeq ($(HAS_SRC_DIR),yes)
	@cd $(SRC_DIR) && go vet ./...
else
	@go vet ./...
endif
	@echo "Vet complete!"

# Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

# Show current platform info
info:
	@echo "Current platform: $(CURRENT_PLATFORM)"
	@echo "Binary name: $(BINARY_NAME)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Current binary: $(CURRENT_BINARY)"

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary for current platform ($(CURRENT_PLATFORM))"
	@echo "  build-all       - Build for all platforms and create launcher script"
	@echo "  rebuild         - Clean all and rebuild from scratch"
	@echo "  install         - Install current platform binary to /usr/local/bin (or TARGET)"
	@echo "  install-launcher - Install launcher script with all platform binaries"
	@echo "  uninstall       - Remove installed binary"
	@echo "  clean           - Remove build artifacts"
	@echo "  clean-all       - Remove build artifacts, go.mod, and go.sum"
	@echo "  test            - Run tests"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  check           - Run fmt, vet, and test"
	@echo "  info            - Show current platform information"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Platform-specific binaries are created in $(BUILD_DIR)/ with suffixes:"
	@echo "  -linux-amd64   - Linux (Intel/AMD 64-bit)"
	@echo "  -darwin-amd64  - macOS (Intel)"
	@echo "  -darwin-arm64  - macOS (Apple Silicon)"
	@echo ""
	@echo "The launcher script ($(BINARY_NAME).sh) automatically selects the right binary."
