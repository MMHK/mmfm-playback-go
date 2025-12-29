# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=mmfm-playback-go
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the project
build:
	$(GOBUILD) -mod=readonly -o bin/$(BINARY_NAME) -v ./cmd/mmfm-playback

# Run tests
test:
	$(GOTEST) -mod=readonly -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -mod=readonly -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build files
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -mod=readonly -a -installsuffix cgo -o bin/$(BINARY_UNIX) -v ./cmd/mmfm-playback

# Run the application
run: build
	./bin/$(BINARY_NAME)

# Install dependencies
deps:
	$(GOGET) -v ./...

.PHONY: build test test-coverage clean run build-linux deps