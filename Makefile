BINARY_NAME := shrinkray
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build run test lint ci clean snapshot

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/shrinkray/

run: build
	./$(BINARY_NAME)

test:
	go test ./...

lint:
	golangci-lint run ./...

ci: lint test build

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf dist/

snapshot:
	goreleaser build --snapshot --clean
