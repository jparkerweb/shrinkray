#!/usr/bin/env bash
# Build shrinkray for the current platform.
# Usage: ./scripts/build.sh [clean|run|ci]

set -euo pipefail
cd "$(dirname "$0")/.."

BINARY_NAME="shrinkray"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

case "${1:-build}" in
  build)
    echo "Building ${BINARY_NAME} ${VERSION}..."
    CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/shrinkray/
    echo "Built ./${BINARY_NAME}"
    ;;
  run)
    echo "Building and running ${BINARY_NAME}..."
    CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/shrinkray/
    ./"${BINARY_NAME}"
    ;;
  test)
    echo "Running tests..."
    go test ./...
    ;;
  lint)
    echo "Running linter..."
    golangci-lint run ./...
    ;;
  ci)
    echo "Running CI pipeline (lint → test → build)..."
    golangci-lint run ./...
    go test ./...
    CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/shrinkray/
    echo "CI passed."
    ;;
  clean)
    echo "Cleaning..."
    rm -f "${BINARY_NAME}" "${BINARY_NAME}.exe"
    rm -rf dist/
    echo "Clean."
    ;;
  *)
    echo "Usage: $0 [build|run|test|lint|ci|clean]"
    exit 1
    ;;
esac
