.PHONY: build install uninstall clean clean-all rebuild test fmt vet

# Binary name
BINARY_NAME=$(shell basename $$(pwd))
SRC_DIR=src
BUILD_DIR=.

# Source files
SOURCES=$(shell find $(SRC_DIR) -name '*.go')

# Build target
build: $(BUILD_DIR)/$(BINARY_NAME)

rebuild: clean-all build

$(BUILD_DIR)/$(BINARY_NAME): $(SOURCES) $(SRC_DIR)/go.sum
	@echo "Building $(BINARY_NAME)..."
	cd $(SRC_DIR) && go build -o ../$(BINARY_NAME) .
	@echo "Build complete! Binary: $(BINARY_NAME)"

# Generate go.sum
$(SRC_DIR)/go.sum: $(SRC_DIR)/go.mod
	@echo "Downloading dependencies..."
	@cd $(SRC_DIR) && go mod download
	@cd $(SRC_DIR) && go mod tidy
	@touch $(SRC_DIR)/go.sum
	@echo "Dependencies downloaded"

# Generate go.mod (only if it doesn't exist)
$(SRC_DIR)/go.mod:
	@echo "Initializing Go module..."
	@cd $(SRC_DIR) && go mod init $(BINARY_NAME)


# Install binary
install: build
ifndef TARGET
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete!"
else
	@echo "Installing $(BINARY_NAME) to $(TARGET)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(TARGET)/ 2>/dev/null || sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(TARGET)/
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
			echo "Uninstallation complete!"; \
		else \
			echo "$(BINARY_NAME) found at $$BINARY_PATH but not in a standard bin directory"; \
			echo "Please remove it manually if needed"; \
		fi; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete!"

clean-all: clean
	@echo "Cleaning go.mod & go.sum"
	@rm -f $(SRC_DIR)/go.mod $(SRC_DIR)/go.sum
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	cd $(SRC_DIR) && go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	cd $(SRC_DIR) && go fmt ./...
	@echo "Format complete!"

# Run go vet
vet:
	@echo "Running go vet..."
	cd $(SRC_DIR) && go vet ./...
	@echo "Vet complete!"

# Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  rebuild    - Clean all and rebuild from scratch"
	@echo "  install    - Build and install to /usr/local/bin (or TARGET env variable)"
	@echo "  uninstall  - Remove installed binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  clean-all  - Remove build artifacts, go.mod, and go.sum"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run fmt, vet, and test"
	@echo "  help       - Show this help message"
