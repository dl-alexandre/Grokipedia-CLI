.PHONY: build build-all build-linux test test-integration lint release clean

BINARY_NAME=grokipedia
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./main.go

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Linux builds
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./main.go

# macOS builds
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./main.go

# Windows builds
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe ./main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run integration tests (requires API access)
test-integration:
	go test -v -tags=integration ./...

# Run linter
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out

# Release build (optimized)
release: clean
	CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) ./main.go

# Development build with debug info
dev:
	go build -o $(BINARY_NAME) ./main.go

# Install locally
install: build
	go install ./...
