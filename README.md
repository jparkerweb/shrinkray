# Shrinkray

**Less bytes, same vibes.**

A cross-platform CLI video compression tool powered by FFmpeg,  
featuring a wizard-style TUI built with Go and Charm's Bubble Tea.

<img src=".docs/shrinkray.jpg" width="600">

## Features

- **Interactive TUI wizard** -- step-by-step guided compression with real-time preview
- **18 built-in presets** -- quality tiers, purpose-driven, and platform-specific (Discord, YouTube, TikTok, etc.)
- **Smart recommendations** -- analyzes your video and suggests the optimal preset
- **Hardware acceleration** -- auto-detects NVIDIA NVENC, AMD AMF, Apple VideoToolbox, Intel QSV
- **Batch processing** -- compress entire folders with parallel workers and progress tracking
- **Headless mode** -- scriptable `--no-tui` mode with stdin pipe support for CI/CD pipelines
- **Cross-platform** -- single static binary for Windows, macOS, and Linux

## Quick Install

### GitHub Releases

Download the latest binary for your platform from the [Releases page](https://github.com/jparkerweb/shrinkray/releases).

### Shell Script (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.sh | bash
```

### PowerShell (Windows)

```powershell
irm https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.ps1 | iex
```

### Go Install

```bash
go install github.com/jparkerweb/shrinkray/cmd/shrinkray@latest
```

## Quick Start

```bash
# Interactive TUI mode (default)
shrinkray

# Compress a single video with a preset
shrinkray -i video.mp4 -p balanced --no-tui

# Batch compress a folder
shrinkray -i ./videos/ -p compact -j 4 --recursive --no-tui
```

## Presets

### Quality Tier

| Key | Name | Description |
|-----|------|-------------|
| `lossless` | Lossless | Mathematically identical output -- no quality loss |
| `ultra` | Ultra Quality | Near-lossless H.265 -- ideal for archival or mastering |
| `high` | High Quality | Excellent quality with good compression |
| `balanced` | Balanced | Best trade-off between quality and file size (recommended) |
| `compact` | Compact | Smaller files capped at 720p -- good for sharing |
| `tiny` | Tiny | Maximum compression at 480p -- smallest files |

### Purpose-Driven

| Key | Name | Description |
|-----|------|-------------|
| `web` | Web Delivery | Optimized for web streaming and downloads |
| `email` | Email Friendly | Targets ~20MB output for email attachments |
| `archive` | Archive | Maximum quality retention with H.265 |
| `slideshow` | Slideshow / Screencast | Optimized for low-motion content |
| `4k-to-1080` | 4K to 1080p | Downscale 4K footage to 1080p |

### Platform-Specific

| Key | Name | Description |
|-----|------|-------------|
| `discord` | Discord | Fits within Discord's 10MB free-tier limit |
| `discord-nitro` | Discord Nitro | Fits within Discord Nitro's 50MB limit |
| `whatsapp` | WhatsApp | Fits within WhatsApp's 16MB limit |
| `twitter` | Twitter / X | Optimized for Twitter upload |
| `instagram` | Instagram | Optimized for Instagram |
| `tiktok` | TikTok | Optimized for TikTok |
| `youtube` | YouTube Upload | Upload-optimized for YouTube |

View preset details:

```bash
shrinkray presets              # List all presets
shrinkray presets show compact # Show detailed settings for a preset
```

## CLI Reference

```
shrinkray [flags]
shrinkray [command]
```

### Commands

| Command | Description |
|---------|-------------|
| `version` | Print version information |
| `presets` | List all available encoding presets |
| `presets show <key>` | Show detailed info about a preset |
| `probe` | Probe a video file and display metadata |
| `completion` | Generate shell completion scripts |
| `help` | Help about any command |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input video file or directory |
| `--preset` | `-p` | Preset name (e.g., balanced, compact, tiny) |
| `--output` | `-o` | Output file path |
| `--no-tui` | | Disable interactive TUI, use headless mode |
| `--crf` | | CRF value override |
| `--codec` | | Video codec (h264, h265, av1, vp9) |
| `--resolution` | | Output resolution (e.g., 1920x1080) |
| `--suffix` | | Output filename suffix (default: _shrunk) |
| `--jobs` | `-j` | Number of parallel encoding workers |
| `--recursive` | `-r` | Recurse into directories for video files |
| `--sort` | | Sort order: size-asc, size-desc, name, duration |
| `--skip-existing` | | Skip files whose output already exists |
| `--skip-optimal` | | Skip files already compressed with target codec |
| `--in-place` | | Replace source files after verification |
| `--output-dir` | | Output directory (mirrors input structure) |
| `--retry-failed` | | Retry failed jobs from persisted queue |
| `--max-retries` | | Maximum retry attempts per file (default: 2) |
| `--dry-run` | | Print FFmpeg command without executing |
| `--stdin` | | Read file paths from stdin (one per line) |
| `--strip-metadata` | | Remove all metadata from output |
| `--keep-metadata` | | Preserve source metadata (default: true) |
| `--metadata-title` | | Set output title metadata |
| `--open` | | Open output folder after completion |
| `--config` | | Path to config file |
| `--log-level` | | Log level (debug, info, warn, error) |
| `--no-color` | | Disable color output |

## Configuration

shrinkray looks for a config file at:

- **Linux/macOS:** `~/.config/shrinkray/config.yaml`
- **Windows:** `%APPDATA%\shrinkray\config.yaml`

Example config:

```yaml
defaults:
  preset: balanced
  codec: h265

output:
  suffix: _shrunk
  mode: suffix
  conflict: rename

batch:
  jobs: 2
  skip_existing: false
  sort: size-asc

ui:
  theme: neon-dusk
```

## Requirements

- **FFmpeg** and **FFprobe** must be installed and available on your PATH
  - macOS: `brew install ffmpeg`
  - Linux: `sudo apt install ffmpeg` (or your distro's package manager)
  - Windows: `scoop install ffmpeg` or `choco install ffmpeg`
- **Go 1.23+** (only required for building from source)

## Building from Source

```bash
git clone https://github.com/jparkerweb/shrinkray.git
cd shrinkray
make build
./shrinkray version
```

Build with version injection:

```bash
make build
# or manually:
go build -ldflags "-s -w -X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/shrinkray/
```

Run tests:

```bash
make test
make lint
```

## License

MIT
