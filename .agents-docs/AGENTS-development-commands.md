# Development Commands
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Build & Run

```bash
make build          # Build the shrinkray binary
make run            # Build and run
go run ./cmd/shrinkray/   # Run directly without building
```

## Testing

```bash
make test           # Run all tests
go test ./...       # Run all tests (direct)
go test ./internal/engine/...   # Run tests for a specific package
go test -run TestName ./internal/engine/   # Run a single test
```

## Linting

```bash
make lint           # Run golangci-lint
golangci-lint run   # Direct lint
```

## Release

```bash
make release-dry    # Dry-run GoReleaser (no publish)
```

## Module Management

```bash
go mod tidy         # Clean up go.mod/go.sum
go mod download     # Download dependencies
```

## Notes

- Go 1.25.0+ is required
- FFmpeg and FFprobe must be installed and on PATH for runtime functionality
- The Makefile is the canonical source for development commands — check it for any additions
