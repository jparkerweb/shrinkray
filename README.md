# Shrinkray

**Less bytes, same vibes.**

A cross-platform CLI video compression tool powered by FFmpeg,  
featuring a wizard-style TUI built with Go and Charm's Bubble Tea.

<img src="docs/shrinkray.jpg" width="600">

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
| `discord-nitro` | Discord Nitro | Fits within Discord Nitro's 50MB limit at 1080p |
| `whatsapp` | WhatsApp | Fits within WhatsApp's 16MB limit, capped at 30fps |
| `twitter` | Twitter / X | Optimized for Twitter upload -- 1080p max, 60fps |
| `instagram` | Instagram | Optimized for Instagram -- 1080p max, 30fps |
| `tiktok` | TikTok | Optimized for TikTok -- 1080p max, 60fps, portrait-aware |
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
| `--inputs` | | Additional input video file paths (repeatable) |
| `--preset` | `-p` | Preset name (e.g., balanced, compact, tiny) |
| `--output` | `-o` | Output file path |
| `--no-tui` | | Disable interactive TUI, use headless mode |
| `--crf` | | CRF value override |
| `--codec` | | Video codec (h264, h265, av1, vp9) |
| `--resolution` | | Output resolution (e.g., 1920x1080) |
| `--fps` | | Maximum output framerate |
| `--speed-preset` | | Encoder speed preset (e.g., fast, medium, slow) |
| `--target-size` | | Target output size (e.g., 25mb, 1gb) -- forces two-pass |
| `--two-pass` | | Force two-pass encoding |
| `--hw-accel` | | Auto-detect and use hardware encoder |
| `--no-hw-accel` | | Force software encoding |
| `--audio-codec` | | Audio codec: aac, opus, copy, none |
| `--audio-bitrate` | | Audio bitrate: 64k, 96k, 128k, 192k, 256k |
| `--audio-channels` | | Audio channels: stereo, mono, source |
| `--suffix` | | Output filename suffix (default: _shrunk) |
| `--output-dir` | | Output directory (mirrors input structure) |
| `--overwrite` | | Overwrite existing output files |
| `--auto-rename` | | Auto-rename output if file exists |
| `--in-place` | | Replace source files after verification |
| `--jobs` | `-j` | Number of parallel encoding workers |
| `--recursive` | `-r` | Recurse into directories for video files |
| `--sort` | | Sort order: size-asc, size-desc, name, duration |
| `--skip-existing` | | Skip files whose output already exists |
| `--skip-optimal` | | Skip files already compressed with target codec |
| `--retry-failed` | | Retry failed jobs from persisted queue |
| `--max-retries` | | Maximum retry attempts per file (default: 2) |
| `--dry-run` | | Print FFmpeg command without executing |
| `--stdin` | | Read file paths from stdin (one per line) |
| `--extra-args` | | Extra FFmpeg arguments (repeatable) |
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
  skipExisting: false
  sort: size-asc

ui:
  theme: neon-dusk
  animations: true

ffmpeg:
  ffmpegPath: ""
  ffprobePath: ""
```

## Requirements

- **FFmpeg** and **FFprobe** must be installed and available on your PATH
  - macOS: `brew install ffmpeg`
  - Linux: `sudo apt install ffmpeg` (or your distro's package manager)
  - Windows: `scoop install ffmpeg` or `choco install ffmpeg`
- **Go 1.25+** (only required for building from source)

## Development

### Building from Source

```bash
git clone https://github.com/jparkerweb/shrinkray.git
cd shrinkray
make build
./shrinkray version
```

### Local Development (Windows)

Build and install to your local PATH for testing:

```powershell
.\scripts\install-local.ps1
```

This builds with version info from `git describe --tags`, installs to `%LOCALAPPDATA%\shrinkray\`, and adds it to your user PATH. The version string (e.g., `v0.2.0-3-gabcdef-dirty`) reflects the latest tag, commits since that tag, and whether there are uncommitted changes.

### Running Tests

```bash
make test          # Run tests
make lint          # Run linter
make ci            # Run lint + test + build (full CI pipeline locally)
```

### Releasing

This project uses [Semantic Versioning](https://semver.org/) with automated releases via GoReleaser.

**Workflow:**

1. Ensure all changes are committed and `CHANGELOG.md` `[Unreleased]` section is up to date
2. Rename `[Unreleased]` to `[X.Y.Z] - YYYY-MM-DD` in `CHANGELOG.md`
3. Add a fresh `[Unreleased]` section and update comparison links at the bottom
4. Commit: `git commit -m "Release vX.Y.Z"`
5. Tag: `git tag -a vX.Y.Z -m "vX.Y.Z"`
6. Push: `git push origin main --tags`

Pushing a `v*` tag triggers the GitHub Actions [release workflow](.github/workflows/release.yml), which runs tests and then GoReleaser builds binaries for Windows, macOS, and Linux (amd64 + arm64). Binaries appear on the [Releases page](https://github.com/jparkerweb/shrinkray/releases).

**Version bumping:**
- Breaking changes → major (X+1.0.0)
- New features → minor (X.Y+1.0)
- Bug fixes only → patch (X.Y.Z+1)

If using Claude Code, you can run `/release` to automate this workflow.

## License

MIT
