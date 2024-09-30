# Define the Go source files and output binary names
RESOURCE_GO_FILE=cmd/resource/main.go
RESOURCE_BINARY=cmd/resource/resource
SERVER_GO_FILE=cmd/server/main.go
SERVER_BINARY=cmd/server/server

# Check if Go is installed
check-go:
	@command -v go &> /dev/null || { echo "Go could not be found. Please install Go and try again."; exit 1; }

# Create tables in database
create-tables: check-go
	@echo "reate tables in database..."
	go build -o $(RESOURCE_BINARY) $(RESOURCE_GO_FILE) || { echo "Failed to build the resources."; exit 1; }

# Refresh data
refresh: check-go
	@echo "Refreshing data in database..."
	go run $(RESOURCE_GO_FILE) --refresh || { echo "Failed to refresh data in database."; exit 1; }

# Cleanup cache
cleanup: check-go
	@echo "Cleaning up cache..."
	go run $(RESOURCE_GO_FILE) --cleanup || { echo "Failed to clean up resources."; exit 1; }

# Build the server application
build: check-go
	@echo "Building the Go application..."
	go build -o $(SERVER_BINARY) $(SERVER_GO_FILE) || { echo "Failed to build the Go application."; exit 1; }

# Run the server application
run: build
	@echo "Running the Go application..."
	./$(SERVER_BINARY) || { echo "Failed to run the Go application."; exit 1; }
	@echo "Cleaning up the binary file..."
	rm -f $(SERVER_BINARY)


# Default target
.PHONY: all
all:  build run
