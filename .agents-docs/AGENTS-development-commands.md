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

## CI (Local Verification)

```bash
make ci             # Run lint → test → build (mirrors GitHub Actions pipeline)
```

Always run `make ci` before pushing to catch lint errors, test failures, and build issues locally.

## Release Snapshot

```bash
make snapshot       # Dry-run GoReleaser build (no publish, local testing)
```

## Cleanup

```bash
make clean          # Remove built binaries and dist/ directory
```

## Cross-Platform Build Scripts

For environments where `make` is unavailable, equivalent scripts are provided:

```bash
# Unix/macOS
./scripts/build.sh [build|run|test|lint|ci|clean]

# Windows PowerShell
.\scripts\build.ps1 -Command [build|run|test|lint|ci|clean]
```

## Local Development Install (Windows)

```powershell
.\scripts\install-local.ps1
```

Builds with version info from `git describe --tags`, installs to `%LOCALAPPDATA%\shrinkray\`, and adds it to your user PATH.

## Package Manager Update Scripts

```bash
./scripts/update-homebrew.sh    # Generate Homebrew formula with version/SHA256
./scripts/update-scoop.sh       # Generate Scoop manifest with version/SHA256
```

## Remote Install Scripts

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.ps1 | iex
```

## Module Management

```bash
go mod tidy         # Clean up go.mod/go.sum
go mod download     # Download dependencies
```

## Windows Dev Setup

Prerequisites: Go, Make, golangci-lint. Install via PowerShell:

```powershell
winget install GoLang.Go
winget install GnuWin32.Make
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

Restart the terminal after installing so PATH updates take effect. Note: golangci-lint is not available in winget — use `go install` instead.

## Notes

- Go 1.25.0+ is required
- FFmpeg and FFprobe must be installed and on PATH for runtime functionality
- The Makefile is the canonical source for development commands — check it for any additions
