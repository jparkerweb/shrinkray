# Technology Stack
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Language

**Go 1.23.0+** — chosen for cross-compilation, single static binary, excellent `os/exec` support, and the Charm TUI ecosystem.

## Core Dependencies

| Dependency | Import Path | Purpose |
|-----------|-------------|---------|
| Bubble Tea v2 | `charm.land/bubbletea/v2` | TUI framework (Model-View-Update) |
| Bubbles v2 | `charm.land/bubbles/v2` | Pre-built TUI components (progress, spinner, filepicker, list, table, viewport, textinput, help, key, timer) |
| Lip Gloss v2 | `charm.land/lipgloss/v2` | Terminal styling (colors, borders, layout, gradients) |
| Huh v2 | `charm.land/huh/v2` | Interactive forms (select, multi-select, input, confirm) |
| Fang | `github.com/charmbracelet/fang` | CLI skeleton (Cobra-based, styled help, completions) |
| Charm Log | `github.com/charmbracelet/log` | Structured logging (slog-compatible) |
| go-ffprobe v2 | `gopkg.in/vansante/go-ffprobe.v2` | FFprobe JSON parsing into Go structs |
| YAML v3 | `gopkg.in/yaml.v3` | Config file parsing |

## External Runtime Dependencies

| Dependency | Required | Detection |
|-----------|----------|-----------|
| FFmpeg | Yes | `exec.LookPath("ffmpeg")` |
| FFprobe | Yes | `exec.LookPath("ffprobe")` |

shrinkray does NOT bundle FFmpeg. If not found, it prints platform-specific install instructions.

## Build Tooling

| Tool | Purpose |
|------|---------|
| GoReleaser | Cross-compilation, packaging, GitHub Releases |
| Make | Dev shortcuts (`make build`, `make test`, `make lint`) |
| golangci-lint | Linting |
| VHS | Terminal recording for demo GIFs |
