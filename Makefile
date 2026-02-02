.PHONY: lint lint-fix fmt test build clean install-tools

# Run all linters
lint:
	golangci-lint run ./...

# Run linters and auto-fix issues (like ruff --fix)
lint-fix:
	golangci-lint run --fix ./...

# Format code (like ruff format)
fmt:
	gofmt -s -w .
	goimports -w .

# Run tests
test:
	go test -v -race ./...

# Build
build:
	go build -v ./...

# Clean
clean:
	go clean
	rm -f coverage.out

# Install required tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run all checks (lint + test)
check: lint test
