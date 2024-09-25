# Define the Go source files and output binary names
RESOURCE_GO_FILE=cmd/resource/main.go
RESOURCE_BINARY=cmd/resource/resource
SERVER_GO_FILE=cmd/server/main.go
SERVER_BINARY=cmd/server/server

# Check if Go is installed
check-go:
	@command -v go &> /dev/null || { echo "Go could not be found. Please install Go and try again."; exit 1; }

# Refresh resources
refresh: check-go
	@echo "Refreshing resources..."
	go run $(RESOURCE_GO_FILE) --refresh || { echo "Failed to refresh resources."; exit 1; }

# Build resources
build-resource: check-go
	@echo "Building resources..."
	go build -o $(RESOURCE_BINARY) $(RESOURCE_GO_FILE) || { echo "Failed to build the resources."; exit 1; }

# Cleanup resources
cleanup: check-go
	@echo "Cleaning up resources..."
	go run $(RESOURCE_GO_FILE) --cleanup || { echo "Failed to clean up resources."; exit 1; }

# Build the server application
build-server: check-go
	@echo "Building the Go application..."
	go build -o $(SERVER_BINARY) $(SERVER_GO_FILE) || { echo "Failed to build the Go application."; exit 1; }

# Run the server application
run-server: build-server
	@echo "Running the Go application..."
	./$(SERVER_BINARY) || { echo "Failed to run the Go application."; exit 1; }
	@echo "Cleaning up the binary file..."
	rm -f $(SERVER_BINARY)

# Default target
.PHONY: all
all: refresh build-resource run-server
