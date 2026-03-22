# shrinkray -- Complete Product Specification

> **Purpose of this document:** This is the exhaustive, self-contained specification for `shrinkray`, a glamorous cross-platform CLI video compression tool powered by FFmpeg and built with Go using Charm's TUI libraries. This document contains every detail needed for an AI agent or developer to build the entire application from scratch. Nothing is left to interpretation. Every feature, every screen, every flag, every preset, every edge case is documented below.

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Technology Stack](#2-technology-stack)
3. [Project Structure](#3-project-structure)
4. [Application Architecture](#4-application-architecture)
5. [CLI Interface](#5-cli-interface)
6. [Video Source Selection](#6-video-source-selection)
7. [Output Handling](#7-output-handling)
8. [The Preset System](#8-the-preset-system)
9. [FFmpeg Engine](#9-ffmpeg-engine)
10. [Hardware Acceleration](#10-hardware-acceleration)
11. [Progress Tracking](#11-progress-tracking)
12. [Batch Processing](#12-batch-processing)
13. [File Size Targeting](#13-file-size-targeting)
14. [Smart Recommendation Engine](#14-smart-recommendation-engine)
15. [TUI Screens](#15-tui-screens)
16. [Styling and Theming](#16-styling-and-theming)
17. [Configuration System](#17-configuration-system)
18. [Logging](#18-logging)
19. [Cross-Platform Considerations](#19-cross-platform-considerations)
20. [Distribution and Installation](#20-distribution-and-installation)
21. [Error Handling](#21-error-handling)
22. [Feature Roadmap](#22-feature-roadmap)

---

## 1. Product Overview

### 1.1 What is shrinkray?

`shrinkray` is an interactive terminal application for shrinking video files. It wraps FFmpeg with a beautiful, glamorous TUI (Terminal User Interface) that guides users through selecting videos, choosing compression settings, and monitoring encoding progress. It runs on Windows, macOS, and Linux.

### 1.2 The Problem

Video compression with FFmpeg requires memorizing arcane command-line flags (`-c:v libx265 -crf 28 -preset medium -c:a aac -b:a 128k -movflags +faststart`). Most users don't know what CRF is, what codec to pick, or how to target a specific file size. Online compression tools have file size limits, privacy concerns, and upload/download overhead. HandBrake's CLI has no interactive mode and a steep learning curve.

### 1.3 The Solution

shrinkray provides:
- A wizard-style TUI that walks users through video compression step by step
- 18 built-in presets with human-friendly names like "Discord", "Email", "Balanced"
- Automatic hardware acceleration detection (NVIDIA, Intel, AMD, Apple Silicon)
- Smart source analysis that recommends the best preset for each video
- Real-time progress tracking with ETA, speed, and compression stats
- Target file size mode ("make this fit in 10 MB for Discord")
- Batch processing with queue management, resume, and skip-if-compressed
- A non-interactive headless mode for scripting and CI/CD
- Cross-platform support (Windows, macOS, Linux) as a single binary

### 1.4 Name and Branding

- **Name:** `shrinkray`
- **Tagline:** "Less bytes, same vibes."
- **ASCII Logo:**
  ```
    ─═══════> ⚡  shrinkray
                  less bytes, same vibes.
  ```
- **Primary Color Palette ("Neon Dusk"):**
  - Primary: `#7B2FF7` (electric violet) -- highlights, active selections, progress bars
  - Secondary: `#FF2D95` (hot pink) -- accents, file sizes, important stats
  - Tertiary: `#00F0FF` (cyan) -- info text, codec details, metadata
  - Success: `#39FF14` (neon green) -- completion states, savings percentages
  - Warning: `#FFAB00` (amber) -- caution states, quality warnings
  - Muted: `#6C6C8A` (soft gray) -- borders, disabled items, secondary text
  - Text: `#E8E8F0` (near-white) -- primary readable text
- **Alternative Color Palette ("Electric Sunset"):**
  - Primary: `#FF6B6B` (coral red) -- headers, active elements
  - Secondary: `#FFD93D` (golden yellow) -- progress bars, highlights
  - Accent: `#FF8E53` (warm orange) -- gradient midpoint
  - Muted: `#6C6C6C` (gray) -- help text, borders

### 1.5 Competitive Positioning

| Tool | Ease of Use | Progress UI | Smart Defaults | Size Targeting | Presets | Batch | HW Accel |
|------|:-----------:|:-----------:|:--------------:|:--------------:|:-------:|:-----:|:--------:|
| HandBrake CLI | Low | Minimal | Via presets | Manual calc | Yes | Yes | Yes |
| Raw FFmpeg | Very Low | None | No | Manual 2-pass | No | Script | Yes |
| ffpb wrapper | Very Low | Bar only | No | No | No | No | Yes |
| GitHub scripts | Medium | None | Partial | Rare | Some | Some | Rare |
| Online tools | Very High | Yes | Yes | Yes | Yes | No | N/A |
| **shrinkray** | **High** | **Rich TUI** | **Yes** | **Yes** | **18+** | **Yes** | **Auto-detect** |

**Key differentiators:**
1. Zero-config smart mode that analyzes source video and recommends optimal settings
2. Target file size with platform presets (Discord 10MB, WhatsApp 16MB, Email 25MB)
3. Rich TUI with gradient progress bars, real-time encoding stats, animated ETA
4. Hardware acceleration auto-detection with graceful software fallback
5. Batch processing with persistent queue, resume, and skip-if-already-compressed

---

## 2. Technology Stack

### 2.1 Language: Go

Go is chosen for: cross-compilation to all target platforms from a single codebase, single static binary with zero runtime dependencies, excellent `os/exec` support for spawning FFmpeg processes, and the Charm ecosystem which is written entirely in Go.

**Minimum Go version:** 1.23.0+ (required by Huh v2)

### 2.2 Dependencies

| Dependency | Import Path | Version | Purpose |
|-----------|-------------|---------|---------|
| **Bubble Tea** | `charm.land/bubbletea/v2` | v2 | TUI framework (Model-View-Update architecture) |
| **Bubbles** | `charm.land/bubbles/v2` | v2 | Pre-built TUI components (progress, spinner, filepicker, list, table, viewport, textinput, help, key, timer, stopwatch) |
| **Lip Gloss** | `charm.land/lipgloss/v2` | v2 | Terminal styling (colors, borders, layout, tables, gradients) |
| **Huh** | `charm.land/huh/v2` | v2 | Interactive forms (select, multi-select, input, confirm, filepicker) |
| **Fang** | `github.com/charmbracelet/fang` | latest | CLI skeleton built on Cobra (styled help, completions, man pages) |
| **Charm Log** | `github.com/charmbracelet/log` | latest | Structured logging (slog-compatible, TUI-safe file output) |
| **go-ffprobe** | `gopkg.in/vansante/go-ffprobe.v2` | v2 | FFprobe JSON output parsing into Go structs |
| **YAML** | `gopkg.in/yaml.v3` | v3 | Configuration file parsing |

### 2.3 External Dependencies

| Dependency | Required | Detection | Install |
|-----------|----------|-----------|---------|
| **FFmpeg** | Yes | `exec.LookPath("ffmpeg")` | Per-platform instructions printed if missing |
| **FFprobe** | Yes | `exec.LookPath("ffprobe")` | Ships with FFmpeg |

shrinkray does NOT bundle FFmpeg. It detects FFmpeg at runtime and provides clear installation instructions if not found:
- Windows: `winget install ffmpeg` or `choco install ffmpeg` or `scoop install ffmpeg`
- macOS: `brew install ffmpeg`
- Linux: `apt install ffmpeg` / `dnf install ffmpeg` / `pacman -S ffmpeg`

### 2.4 Build Tooling

| Tool | Purpose |
|------|---------|
| **GoReleaser** | Cross-compilation, packaging, GitHub Releases, Homebrew/Scoop publishing |
| **Make** | Development shortcuts (`make build`, `make run`, `make test`, `make lint`) |
| **golangci-lint** | Linting |
| **VHS** | Terminal recording for README demo GIFs |

---

## 3. Project Structure

```
shrinkray/
├── cmd/
│   └── shrinkray/
│       └── main.go                     # Entry point. Tiny. Calls cli.Execute().
│
├── internal/
│   ├── cli/                            # CLI layer (Fang/Cobra commands)
│   │   ├── root.go                     # Root command: global flags, FFmpeg detection
│   │   ├── run.go                      # Default command: launches TUI or headless mode
│   │   ├── presets.go                  # `shrinkray presets` subcommand: list/show presets
│   │   ├── probe.go                    # `shrinkray probe <file>` subcommand: show video info
│   │   └── version.go                  # `shrinkray version` subcommand
│   │
│   ├── config/                         # Configuration management
│   │   ├── config.go                   # Config struct, Load(), Save(), defaults
│   │   ├── paths.go                    # Cross-platform config/cache/data directory resolution
│   │   └── custom_presets.go           # Custom preset CRUD operations
│   │
│   ├── engine/                         # FFmpeg/FFprobe business logic (NO TUI dependency)
│   │   ├── probe.go                    # FFprobe wrapper: video metadata extraction
│   │   ├── probe_types.go             # Go structs for video/audio stream info
│   │   ├── hdr.go                      # HDR detection (HDR10, HLG, Dolby Vision)
│   │   ├── encode.go                   # FFmpeg command builder + process runner
│   │   ├── encode_args.go             # Translate preset + options into FFmpeg argument arrays
│   │   ├── progress.go                # Parse FFmpeg `-progress pipe:1` output
│   │   ├── progress_types.go          # ProgressUpdate struct definition
│   │   ├── hwaccel.go                  # Hardware encoder detection and selection
│   │   ├── hwaccel_probe.go           # Test-encode to verify encoder actually works
│   │   ├── cancel.go                   # Graceful FFmpeg process termination (q → SIGINT → SIGKILL)
│   │   ├── estimate.go                # Output size estimation from source metadata + preset
│   │   └── target_size.go             # Two-pass bitrate calculation for file size targeting
│   │
│   ├── presets/                        # Preset definitions and recommendation
│   │   ├── preset.go                   # Preset struct definition
│   │   ├── quality.go                  # 6 quality-tier presets (lossless → potato)
│   │   ├── purpose.go                  # 5 purpose-driven presets (web, email, archive, etc.)
│   │   ├── platform.go                # 7 platform-specific presets (discord, whatsapp, etc.)
│   │   ├── registry.go                # Combined preset registry with lookup by key/tag/fuzzy
│   │   └── recommend.go               # Smart recommendation engine (analyze source → suggest)
│   │
│   ├── tui/                            # TUI presentation layer (Bubble Tea)
│   │   ├── app.go                      # Top-level model: screen routing, global keys, resize
│   │   ├── keys.go                     # Global key bindings (q, ?, ctrl+c, ctrl+t)
│   │   ├── styles.go                   # All Lip Gloss style definitions, theme struct
│   │   ├── theme.go                    # Theme switching logic (Neon Dusk, Electric Sunset)
│   │   ├── screens/
│   │   │   ├── splash.go              # Screen 1: ASCII logo, system capabilities
│   │   │   ├── filepicker.go          # Screen 2: File browser + path input
│   │   │   ├── info.go                # Screen 3: Source video metadata card
│   │   │   ├── presets.go             # Screen 4: Preset grid selection
│   │   │   ├── advanced.go            # Screen 5: Full options form (Huh)
│   │   │   ├── preview.go             # Screen 6: Before/after confirmation
│   │   │   ├── encoding.go            # Screen 7: Progress display
│   │   │   ├── complete.go            # Screen 8: Results + bar chart
│   │   │   ├── batch_queue.go         # Screen 9a: Batch queue overview
│   │   │   ├── batch_progress.go      # Screen 9b: Batch encoding progress
│   │   │   └── batch_complete.go      # Screen 9c: Batch completion summary
│   │   └── messages/
│   │       └── messages.go             # Custom tea.Msg types shared across screens
│   │
│   └── logging/
│       └── logging.go                  # slog setup: file output in TUI mode, stderr in headless
│
├── .goreleaser.yaml                    # GoReleaser config for cross-platform builds
├── go.mod
├── go.sum
├── Makefile                            # build, run, test, lint, release-dry targets
├── LICENSE                             # MIT
├── README.md
└── IDEA.md                             # This file
```

### 3.1 Key Architectural Principle: Separation of Concerns

The `engine/` package contains ALL FFmpeg/FFprobe business logic. It has ZERO dependency on Bubble Tea, Lip Gloss, or any TUI library. This means:
- The engine can be tested independently with unit tests
- The headless (non-interactive) mode uses the same engine as the TUI
- A future GUI or web frontend could reuse the engine package

The `tui/` package is purely a presentation layer. It calls into `engine/` for all video operations and displays the results.

The `presets/` package defines preset data structures and the recommendation algorithm. It has no dependency on the TUI or the engine — it just describes what settings to use.

---

## 4. Application Architecture

### 4.1 Bubble Tea Model-View-Update Pattern

shrinkray uses the Elm Architecture implemented by Bubble Tea:

- **Model:** A Go struct holding all application state
- **Update:** A function that receives events (keyboard, mouse, window resize, custom messages) and returns the updated model plus any commands to execute
- **View:** A function that renders the current model state as a string for terminal display
- **Commands (`tea.Cmd`):** Functions that perform I/O asynchronously (probe a file, run FFmpeg, detect hardware) and return a message when done

The top-level `App` model contains a `screen` enum that determines which sub-model is active. Each screen is its own sub-model with its own Update/View methods.

### 4.2 Screen Routing

```go
type screen int

const (
    screenSplash    screen = iota  // ASCII logo, system info
    screenFilePicker               // File browser or path input
    screenInfo                     // Source video metadata display
    screenPresets                  // Preset grid selection
    screenAdvanced                 // Full encoding options (Huh form)
    screenPreview                  // Before/after confirmation
    screenEncoding                 // Real-time progress
    screenComplete                 // Results summary
    screenBatchQueue               // Batch file list + settings
    screenBatchProgress            // Batch encoding progress
    screenBatchComplete            // Batch results summary
)
```

The App model's `Update` method:
1. Handles global keys (q to quit, ? for help, ctrl+c, ctrl+t for theme) regardless of screen
2. Handles `tea.WindowSizeMsg` and propagates to all sub-models
3. Delegates all other messages to the active screen's Update method
4. Watches for "transition" messages that signal moving to the next screen

The App model's `View` method:
1. Renders the persistent header bar (app name left, step indicator + help/quit right)
2. Calls the active screen's View method for the main content area
3. Renders the persistent footer bar (context-sensitive keyboard shortcuts)

### 4.3 Inter-Screen Communication via Messages

Screens do NOT call each other directly. Instead, they return custom `tea.Msg` types that the parent App model catches and uses to trigger transitions:

```go
// messages/messages.go

// FileSelectedMsg is sent when the user picks a file in the file picker
type FileSelectedMsg struct {
    Path string
}

// FilesSelectedMsg is sent when the user picks multiple files (batch mode)
type FilesSelectedMsg struct {
    Paths []string
}

// PresetSelectedMsg is sent when the user picks a preset
type PresetSelectedMsg struct {
    Preset presets.Preset
}

// EncodingStartMsg triggers the encoding screen
type EncodingStartMsg struct {
    Input   string
    Output  string
    Preset  presets.Preset
    Options engine.EncodeOptions
}

// ProgressMsg carries real-time encoding progress from the FFmpeg goroutine
type ProgressMsg struct {
    Percent  float64
    ETA      time.Duration
    Speed    float64
    FPS      float64
    Frame    int
    Size     int64
    Bitrate  string
    Done     bool
    Err      error
}

// EncodingCompleteMsg carries the final results
type EncodingCompleteMsg struct {
    InputPath   string
    OutputPath  string
    InputSize   int64
    OutputSize  int64
    Duration    time.Duration
    Err         error
}
```

### 4.4 Dual Mode: Interactive TUI vs. Headless CLI

shrinkray runs in two modes:

**Interactive TUI mode (default when stdin is a terminal):**
```bash
shrinkray                              # Opens TUI with file picker
shrinkray video.mp4                    # Opens TUI with file pre-loaded (skips to info screen)
shrinkray video1.mp4 video2.mp4        # Opens TUI in batch mode
```

**Headless mode (when `--no-tui` is set, or stdin is not a terminal):**
```bash
shrinkray --no-tui --input video.mp4 --preset balanced
shrinkray --no-tui --input video.mp4 --preset discord --output output.mp4
shrinkray --no-tui --input ./videos/ --preset balanced --recursive
echo video.mp4 | shrinkray --stdin --preset balanced
```

In headless mode, progress is printed to stderr as a simple updating line:
```
[====================          ] 63.2% | 1.4x | ETA 2:18 | 487 MB
```

Both modes use the same `engine/` package for all video operations. The only difference is the presentation layer.

### 4.5 Concurrency Model

- The TUI event loop runs on the main goroutine
- FFmpeg processes are spawned in separate goroutines via `tea.Cmd`
- Progress updates are sent from the FFmpeg goroutine to the TUI via `p.Send()` (injects a `tea.Msg` into the event loop from outside)
- A buffered channel (size 1) with non-blocking sends ensures the progress parser never stalls waiting for the TUI to consume
- FFprobe calls are also spawned as `tea.Cmd` goroutines and return metadata via messages
- Hardware detection is run once at startup as a `tea.Cmd` and the results are cached in the App model

---

## 5. CLI Interface

### 5.1 Commands

```
shrinkray                           # Default: launch TUI (or show help if no TTY)
shrinkray [files...]                # Launch TUI with file(s) pre-loaded
shrinkray run [files...]            # Explicit run command (same as default)
shrinkray presets                   # List all available presets
shrinkray presets show <name>       # Show details for a specific preset
shrinkray probe <file>              # Display video metadata (non-interactive)
shrinkray version                   # Show version, build info, FFmpeg version
shrinkray help                      # Show help
```

### 5.2 Global Flags

```
--no-tui                  Disable interactive TUI, use headless mode
--config <path>           Path to config file (default: ~/.config/shrinkray/config.yaml)
--ffmpeg-path <path>      Path to FFmpeg binary (overrides PATH lookup)
--verbose                 Enable verbose logging (debug level)
--quiet                   Suppress all non-error output (headless mode)
```

### 5.3 Encoding Flags (used in headless mode or to pre-configure TUI)

```
-i, --input <path>        Input file or directory
-o, --output <path>       Output file path (single file) or directory
--preset <name>           Preset to use (see `shrinkray presets`)
--codec <name>            Video codec: h264, h265, av1, vp9 (default: h265)
--crf <number>            CRF value (0-51 for h264/h265, 0-63 for av1/vp9)
--resolution <WxH>        Output resolution (e.g., 1920x1080, 1280x720)
--fps <number>            Output framerate (e.g., 30, 24)
--speed-preset <name>     Encoder speed: ultrafast..veryslow (default: medium)
--audio-codec <name>      Audio codec: aac, opus, copy, none (default: aac)
--audio-bitrate <rate>    Audio bitrate: 64k, 96k, 128k, 192k, 256k (default: 128k)
--target-size <size>      Target output file size: 10mb, 25mb, 50mb, 1gb
--two-pass                Enable two-pass encoding
--hw-accel                Force hardware acceleration (auto-detect by default)
--no-hw-accel             Disable hardware acceleration
--strip-metadata          Remove EXIF/GPS metadata from output
--suffix <string>         Output filename suffix (default: _shrunk)
--output-dir <path>       Write output files to this directory
--in-place                Replace source files (encode to .tmp, verify, replace)
--overwrite               Overwrite existing output files without asking
--skip-existing           Skip files that already have output
--auto-rename             Auto-rename on conflict: video_shrunk(1).mp4
--recursive               Recurse into subdirectories
--jobs <n>                Number of concurrent encoding jobs (default: 2)
--dry-run                 Show what would be done without encoding
--stdin                   Read input file paths from stdin (one per line)
```

### 5.4 Shell Completions

Fang (Cobra) provides automatic shell completion generation:
```bash
shrinkray completion bash > /etc/bash_completion.d/shrinkray
shrinkray completion zsh > ~/.zsh/completions/_shrinkray
shrinkray completion fish > ~/.config/fish/completions/shrinkray.fish
shrinkray completion powershell > shrinkray.ps1
```

Completions include:
- `--preset` completes to all preset names (lossless, high, balanced, low, tiny, potato, web, email, archive, slideshow, 4k-to-1080, discord, discord-nitro, whatsapp, twitter, instagram, tiktok, youtube)
- `--codec` completes to h264, h265, av1, vp9
- `--speed-preset` completes to ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
- `--audio-codec` completes to aac, opus, copy, none
- File arguments complete to video file extensions

---

## 6. Video Source Selection

### 6.1 Supported Input Formats

shrinkray accepts any file that FFprobe can read. The file browser in the TUI filters by extension for convenience, but the actual format support comes from whatever FFmpeg build is installed.

**File browser filter extensions:**
```
.mp4, .mkv, .mov, .avi, .webm, .flv, .wmv, .m4v, .ts, .mpg, .mpeg, .3gp, .ogv, .vob, .mts, .m2ts
```

Before encoding, shrinkray always validates the input with FFprobe. If FFprobe can parse it, shrinkray can shrink it — regardless of extension.

### 6.2 Input Modes

#### Mode A: CLI Arguments (power users, scripting)

```bash
# Single file
shrinkray video.mp4

# Multiple files (batch mode)
shrinkray video1.mp4 video2.mkv video3.mov

# Glob pattern (shrinkray expands, not the shell — use quotes)
shrinkray "~/Videos/**/*.mp4"

# Directory (all videos in top level)
shrinkray ~/Videos/

# Directory recursive
shrinkray ~/Videos/ --recursive

# Stdin pipe (one path per line)
find . -name "*.mp4" -size +100M | shrinkray --stdin --preset balanced --no-tui
```

When arguments are provided, shrinkray auto-detects whether each argument is a file or directory via `os.Stat()`. Directories are scanned for video files matching the supported extensions. Glob patterns are expanded by the application (not the shell) using Go's `filepath.Glob` or `doublestar` library for `**` support — this ensures consistent cross-platform behavior, especially on Windows where shells may not expand globs.

#### Mode B: Interactive TUI File Browser

When shrinkray is launched with no file arguments, the TUI opens a file browser (Screen 2). The file browser:

- Uses the Bubbles `filepicker` component as a foundation, with custom rendering
- Filters to video file extensions by default
- Shows inline metadata for each video file: file size, resolution, and duration
- Resolution and duration are obtained via a quick `ffprobe` call on each visible file (cached, not re-probed on every render)
- Directories are navigable with arrow keys or `h`/`l` (vim-style)
- `Space` toggles batch selection (multi-select) — selected files get a green checkmark
- `/` opens a fuzzy search filter (powered by the Bubbles list filter)
- `Tab` switches between the file browser and a direct path text input at the bottom of the screen
- `Enter` confirms the selection and moves to Screen 3 (info) or Screen 9a (batch queue)
- The starting directory is the current working directory, or the user's home directory if CWD has no video files

The path text input supports:
- Absolute paths: `/home/user/Videos/video.mp4`
- Relative paths: `./video.mp4`
- Home directory expansion: `~/Videos/video.mp4`
- Tab completion for filenames (using Bubbles textinput suggestions)

#### Mode C: Drag-and-Drop

Modern terminals (Windows Terminal, iTerm2, GNOME Terminal, Konsole) paste file paths when files are dragged onto the terminal window. This naturally works as CLI arguments — the user types `shrinkray ` then drags files, which pastes their paths.

### 6.3 Input Validation

After a file is selected (by any mode), shrinkray runs FFprobe to validate it:

1. Run `ffprobe -v quiet -print_format json -show_format -show_streams <file>`
2. Parse the JSON output into Go structs
3. Verify at least one video stream exists
4. Extract: duration, file size, video codec, video bitrate, resolution, framerate, pixel format, color space, audio codec, audio bitrate, audio channels, audio sample rate, HDR metadata
5. If the file cannot be probed (corrupt, not a video, unsupported format), show a clear error and return to the file picker

---

## 7. Output Handling

### 7.1 Output Modes

shrinkray supports three output modes, controlled by flags or config:

#### Mode 1: Suffix (default)

The output file is placed alongside the source file with a configurable suffix appended before the extension.

```
Input:  /home/user/Videos/vacation.mp4
Output: /home/user/Videos/vacation_shrunk.mp4

Input:  /home/user/Videos/interview.mkv
Output: /home/user/Videos/interview_shrunk.mkv
```

The suffix is configurable via `--suffix` flag or `output.suffix` config key. Default: `_shrunk`.

If the encoding changes the container format (e.g., VP9 requires `.webm`), the output extension changes accordingly:
```
Input:  vacation.mp4       (H.264 in MP4)
Output: vacation_shrunk.webm  (VP9 in WebM)
```

Extension mapping:
- H.264 → `.mp4`
- H.265 → `.mp4`
- AV1 → `.mp4` (or `.webm` if user chooses WebM container)
- VP9 → `.webm`

#### Mode 2: Output Directory

When `--output-dir <path>` is specified, all output files go into that directory. For recursive batch operations, the source directory structure is mirrored inside the output directory.

```bash
shrinkray --output-dir ./compressed/ ~/Videos/ --recursive
```
```
~/Videos/2024/trip.mp4     → ./compressed/2024/trip.mp4
~/Videos/2023/xmas.mp4     → ./compressed/2023/xmas.mp4
~/Videos/vacation.mp4      → ./compressed/vacation.mp4
```

The output directory is created if it does not exist. Intermediate directories in the mirrored structure are also created automatically.

When using output directory mode, the suffix is NOT applied by default (the output filename matches the source filename). However, if both `--output-dir` and `--suffix` are specified, the suffix IS applied.

#### Mode 3: Explicit Output Path

```bash
shrinkray video.mp4 -o compressed_video.mp4
```

This only works for single-file operations. If used with multiple files, shrinkray prints an error: "Cannot use -o with multiple input files. Use --output-dir instead."

#### Mode 4: In-Place

```bash
shrinkray video.mp4 --in-place
```

The in-place mode replaces the source file with the compressed version. Because this is destructive, it follows a safe multi-step process:

1. Encode to a temporary file: `vacation.mp4.shrinkray.tmp`
2. Run FFprobe on the temporary file and verify:
   - Duration matches the source within 2 seconds
   - At least one video stream is present
   - File size is greater than 0
3. If verification passes: atomically rename (move) the temp file to replace the source
4. If verification fails: delete the temp file, report the error, source is untouched
5. If `--backup` is also specified: rename the source to `vacation.mp4.bak` before replacing

### 7.2 Output Conflict Handling

When the output file already exists, behavior depends on the mode:

| Mode | Flag | Behavior |
|------|------|----------|
| **Ask** (TUI default) | -- | TUI prompts: "Output exists. Skip / Overwrite / Rename?" |
| **Skip** (headless default) | `--skip-existing` | Move to next file silently |
| **Overwrite** | `--overwrite` | Replace without asking |
| **Auto-rename** | `--auto-rename` | Append `(1)`, `(2)`, etc.: `video_shrunk(1).mp4` |

The `--skip-existing` flag is the default in headless mode because it is the safest for automation. In TUI mode, the default is to ask the user.

### 7.3 Temp File Strategy

During encoding, output is always written to a temp file first:

```
vacation.mp4                    ← source (NEVER modified during encoding)
vacation_shrunk.mp4.tmp         ← FFmpeg writes here
vacation_shrunk.mp4             ← renamed from .tmp ONLY on success
```

If encoding is canceled (user presses `c`, or Ctrl+C, or the process is killed):
- The `.tmp` file is deleted automatically (via defer/cleanup handler)
- The source file is completely untouched
- In batch mode, the queue state is saved so the batch can resume later

If encoding fails (FFmpeg exits with non-zero):
- The `.tmp` file is deleted automatically
- The error is logged and displayed
- In batch mode, the job is marked as `failed` and the next job starts

### 7.4 Output Metadata

By default, shrinkray copies metadata (title, artist, date, GPS, etc.) from the source to the output via FFmpeg's `-map_metadata 0` flag. If `--strip-metadata` is specified, metadata is stripped using `-map_metadata -1`.

For MP4 output files, shrinkray always adds `-movflags +faststart` to move the moov atom to the beginning of the file for progressive web playback.

For H.265 in MP4 containers, shrinkray always adds `-tag:v hvc1` for Apple/QuickTime compatibility.

---

## 8. The Preset System

### 8.1 Preset Data Structure

Every preset is defined as a Go struct:

```go
type Preset struct {
    // Identity
    Key         string   // unique identifier: "balanced", "discord", etc.
    Name        string   // display name: "Balanced", "Discord (Free)", etc.
    Description string   // one-line description: "Good quality, moderate file size"
    Category    string   // "quality", "purpose", "platform"
    Tags        []string // searchable tags: ["general", "default"]
    Icon        string   // emoji/unicode icon for TUI display: "💎", "📱", etc.

    // Video settings
    VideoCodec     string // "h264", "h265", "av1", "vp9" (empty = use default)
    CRF            int    // constant rate factor (0-63 depending on codec)
    SpeedPreset    string // "ultrafast".."veryslow" (empty = "medium")
    MaxResolution  string // "1920x1080", "1280x720", "" = keep original
    MaxFPS         int    // 0 = keep original, 30, 24, etc.
    TwoPass        bool   // use two-pass encoding
    Tune           string // x264/x265 tune: "film", "animation", "stillimage", etc.
    PixelFormat    string // "yuv420p" (default), "yuv420p10le" for 10-bit

    // Audio settings
    AudioCodec     string // "aac", "opus", "copy", "none"
    AudioBitrate   string // "128k", "96k", "64k", etc.
    AudioChannels  int    // 0 = keep original, 1 = mono, 2 = stereo
    AudioSampleRate int   // 0 = keep original, 44100, 48000, etc.

    // File size targeting (if set, overrides CRF with calculated bitrate + 2-pass)
    TargetSizeBytes int64 // 0 = not a target-size preset

    // Container
    Container      string // "mp4", "mkv", "webm" (empty = auto from codec)
    Faststart      bool   // add -movflags +faststart (default true for mp4)

    // Extra
    StripMetadata  bool   // remove EXIF/GPS data

    // Scale filter details
    ScaleFilter    string // custom scale filter (empty = auto from MaxResolution)
    ScaleAlgorithm string // "lanczos", "bicubic", "bilinear" (default: lanczos)
}
```

### 8.2 Category 1: Quality Tiers (6 presets)

These are the "how much do I care about quality?" presets. They default to H.265 for optimal compression but respect a global codec preference if the user has configured one.

#### `lossless` -- Zero Quality Loss
```yaml
key: lossless
name: Lossless
description: "Mathematically identical to source. Large files."
category: quality
icon: "🔒"
tags: [quality, lossless, master, editing]
video_codec: h264          # x264 lossless is faster and more compatible than x265 lossless
crf: 0
speed_preset: veryslow
max_resolution: ""         # keep original
max_fps: 0                 # keep original
audio_codec: copy          # never re-encode audio for lossless
strip_metadata: false
```
**When to use:** Creating editing masters, archiving irreplaceable footage, transcoding between containers without quality loss. Output will be VERY large — often larger than the source if the source used lossy compression.

#### `high` -- Near-Transparent Quality
```yaml
key: high
name: High Quality
description: "Visually indistinguishable from source. ~40-60% smaller."
category: quality
icon: "✨"
tags: [quality, high, archive, transparent]
video_codec: h265
crf: 20
speed_preset: slow
max_resolution: ""
max_fps: 0
audio_codec: aac
audio_bitrate: 192k
faststart: true
```
**When to use:** Long-term archival where you want real compression but zero visible quality loss. Good for home videos, important recordings.

#### `balanced` -- The Default
```yaml
key: balanced
name: Balanced
description: "Good quality, significant compression. The sweet spot."
category: quality
icon: "💎"
tags: [quality, balanced, default, general]
video_codec: h265
crf: 26
speed_preset: medium
max_resolution: ""
max_fps: 0
audio_codec: aac
audio_bitrate: 128k
faststart: true
```
**When to use:** General-purpose compression. This is the DEFAULT preset when none is specified. Good quality on any screen, meaningful file size reduction (typically 50-70% smaller than unoptimized source).

#### `low` -- Visible Quality Loss, Very Small Files
```yaml
key: low
name: Low Quality
description: "Noticeable quality loss but very small files."
category: quality
icon: "📉"
tags: [quality, low, draft, preview]
video_codec: h265
crf: 32
speed_preset: fast
max_resolution: 1280x720   # cap at 720p
max_fps: 30
audio_codec: aac
audio_bitrate: 96k
faststart: true
```
**When to use:** Drafts, previews, quick shares where file size matters more than quality. Artifacts visible on close inspection but watchable.

#### `tiny` -- Smallest Watchable File
```yaml
key: tiny
name: Tiny
description: "Aggressively compressed. Smallest watchable file."
category: quality
icon: "🐜"
tags: [quality, tiny, small, minimal]
video_codec: h265
crf: 36
speed_preset: fast
max_resolution: 854x480    # cap at 480p
max_fps: 24
audio_codec: aac
audio_bitrate: 64k
audio_channels: 1           # mono
faststart: true
strip_metadata: true
```
**When to use:** When you need the absolute smallest file that is still watchable. Email attachments of long videos, archiving content you rarely watch, freeing disk space aggressively.

#### `potato` -- "Prove This Video Exists"
```yaml
key: potato
name: Potato Quality
description: "Maximum compression. Quality is an afterthought."
category: quality
icon: "🥔"
tags: [quality, potato, minimum, extreme]
video_codec: h264           # h264 is faster at extreme compression, compatibility doesn't matter here
crf: 40
speed_preset: ultrafast
max_resolution: 640x360     # 360p
max_fps: 24
audio_codec: aac
audio_bitrate: 48k
audio_channels: 1
audio_sample_rate: 22050
faststart: true
strip_metadata: true
```
**When to use:** Thumbnails, placeholders, proving a video exists, sending a 2-hour lecture recording when only the audio matters. Don't use this for anything you actually want to watch.

### 8.3 Category 2: Purpose-Driven (5 presets)

These presets are optimized for specific use cases rather than quality levels.

#### `web` -- Web Streaming
```yaml
key: web
name: Web Streaming
description: "Optimized for browser playback with fast start."
category: purpose
icon: "🌐"
tags: [purpose, web, streaming, browser, html5]
video_codec: h264           # h264 for universal browser support
crf: 23
speed_preset: medium
max_resolution: 1920x1080   # cap at 1080p
max_fps: 0
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true             # critical for streaming
```
**When to use:** Embedding videos on websites, serving via HTML5 `<video>` tag, uploading to CMS platforms. H.264 is chosen over H.265 because browser support for H.264 is universal while H.265 requires OS-level decoders on some platforms.

**FFmpeg extras:** Adds `-maxrate 5M -bufsize 10M` to prevent bitrate spikes that could cause buffering.

#### `email` -- Email Attachment
```yaml
key: email
name: Email Friendly
description: "Under 25 MB for email attachments."
category: purpose
icon: "✉️"
tags: [purpose, email, attachment, small]
video_codec: h264
target_size_bytes: 26214400  # 25 MB (25 * 1024 * 1024)
max_resolution: 854x480
max_fps: 24
two_pass: true
audio_codec: aac
audio_bitrate: 96k
audio_channels: 2
container: mp4
faststart: true
strip_metadata: true
```
**When to use:** Sending video via Gmail, Outlook, or any email provider with a 25 MB attachment limit. Uses two-pass encoding to hit the target size accurately. Resolution is capped at 480p and framerate at 24fps to maximize duration per megabyte.

**Smart behavior:** If the source video duration makes 25 MB impossible at watchable quality (less than 100kbps video bitrate), shrinkray warns the user and suggests trimming the video or using a file-sharing service instead.

#### `archive` -- Long-Term Storage
```yaml
key: archive
name: Archive
description: "Minimal quality loss for long-term storage. H.265."
category: purpose
icon: "🗄️"
tags: [purpose, archive, storage, backup]
video_codec: h265
crf: 20
speed_preset: slow          # better compression efficiency
max_resolution: ""          # keep original
max_fps: 0
audio_codec: aac
audio_bitrate: 192k
container: mp4
faststart: true
```
**When to use:** Archiving a video library. Optimized for maximum compression with minimal quality loss. Uses H.265 with slow preset for best compression efficiency. Similar to `high` quality tier but explicitly positioned for the "I'm archiving my video library" use case.

#### `slideshow` -- Screen Recordings & Presentations
```yaml
key: slideshow
name: Slideshow / Screen Recording
description: "Optimized for mostly-static content with occasional motion."
category: purpose
icon: "🖥️"
tags: [purpose, slideshow, screen, recording, presentation, tutorial]
video_codec: h264
crf: 24
speed_preset: medium
tune: stillimage
max_resolution: ""
max_fps: 0
audio_codec: aac
audio_bitrate: 96k
container: mp4
faststart: true
```
**When to use:** Screen recordings, slideshow videos, tutorial recordings, presentation captures. The `tune stillimage` option tells x264 to optimize for content that is mostly static with occasional transitions, which dramatically improves compression for this type of content.

#### `4k-to-1080` -- 4K Downscale
```yaml
key: 4k-to-1080
name: 4K to 1080p
description: "Downscale 4K to 1080p with Lanczos. ~75% smaller."
category: purpose
icon: "📐"
tags: [purpose, 4k, downscale, resize, 1080p]
video_codec: h265
crf: 22
speed_preset: medium
max_resolution: 1920x1080
max_fps: 0
scale_algorithm: lanczos    # best quality downscale algorithm
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true
```
**When to use:** You shot in 4K but don't need 4K resolution for storage or sharing. Downscaling to 1080p reduces pixel count by 75%, and combined with H.265, typical total reduction is 70-85%. The Lanczos scaling algorithm preserves maximum sharpness during downscale.

**Smart behavior:** If the source is already 1080p or lower, shrinkray skips the resolution change and informs the user. It NEVER upscales.

### 8.4 Category 3: Platform-Specific (7 presets)

These presets target specific platform upload limits and encoding requirements. They all use H.264 because every social/messaging platform re-encodes uploads — uploading H.265 risks quality loss from double transcoding due to codec mismatch.

#### `discord` -- Discord Free (10 MB)
```yaml
key: discord
name: Discord (Free)
description: "Fit under Discord's 10 MB file size limit."
category: platform
icon: "🎮"
tags: [platform, discord, free, 10mb, gaming]
video_codec: h264
target_size_bytes: 10485760  # 10 MB
max_resolution: ""           # auto-scaled based on duration
max_fps: 30
two_pass: true
audio_codec: aac
audio_bitrate: 96k
container: mp4
faststart: true
strip_metadata: true
```
**Smart behavior:** Before encoding, calculate if the target is feasible:
- If source duration < 30s at 720p: proceed as-is
- If source duration 30s-2min: cap at 720p
- If source duration 2min-5min: cap at 480p
- If source duration > 5min: cap at 360p and warn about quality
- If source duration > 10min: warn that 10 MB is likely insufficient

#### `discord-nitro` -- Discord Nitro (50 MB)
```yaml
key: discord-nitro
name: Discord (Nitro)
description: "Fit under Discord Nitro's 50 MB file size limit."
category: platform
icon: "🎮"
tags: [platform, discord, nitro, 50mb]
video_codec: h264
target_size_bytes: 52428800  # 50 MB
max_resolution: 1920x1080
max_fps: 0
two_pass: true
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true
```

#### `whatsapp` -- WhatsApp (16 MB)
```yaml
key: whatsapp
name: WhatsApp
description: "Fit under WhatsApp's 16 MB file size limit."
category: platform
icon: "📱"
tags: [platform, whatsapp, mobile, 16mb]
video_codec: h264
target_size_bytes: 16777216  # 16 MB
max_resolution: 1280x720
max_fps: 30
two_pass: true
audio_codec: aac
audio_bitrate: 64k
audio_channels: 1            # mono saves precious bytes
container: mp4
faststart: true
strip_metadata: true
```

#### `twitter` -- Twitter/X
```yaml
key: twitter
name: Twitter / X
description: "Optimized for Twitter/X upload. H.264, 1080p max."
category: platform
icon: "🐦"
tags: [platform, twitter, x, social]
video_codec: h264
crf: 20
speed_preset: medium
max_resolution: 1920x1080
max_fps: 30
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true
```
**Note:** Twitter has a 512 MB limit but re-encodes everything. A CRF-based approach (not target-size) makes more sense here since the limit is generous. CRF 20 gives Twitter's encoder a high-quality source to work with.

**Smart behavior:** If source duration exceeds 140 seconds (Twitter's limit for most accounts), display a warning.

#### `instagram` -- Instagram Reels
```yaml
key: instagram
name: Instagram Reels
description: "Vertical 9:16, 1080x1920, optimized for Instagram."
category: platform
icon: "📸"
tags: [platform, instagram, reels, vertical, social]
video_codec: h264
crf: 18
speed_preset: medium
max_resolution: 1080x1920   # note: width x height (portrait)
max_fps: 30
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true
```
**Smart behavior for landscape source videos:**
- Default: pad with blurred background to fill 9:16 frame
- FFmpeg filter: `scale=1080:-2,pad=1080:1920:(ow-iw)/2:(oh-ih)/2:black` (or use `gblur` for blurred fill)
- Alternative (configurable): center-crop to 9:16

**Smart behavior for already-vertical source:**
- If source is already 9:16 or close, just scale to 1080x1920
- If source is a different portrait ratio (e.g., 3:4), pad to fill 9:16

#### `tiktok` -- TikTok
```yaml
key: tiktok
name: TikTok
description: "Vertical 9:16, 1080x1920, optimized for TikTok."
category: platform
icon: "🎵"
tags: [platform, tiktok, vertical, social, short]
video_codec: h264
crf: 18
speed_preset: medium
max_resolution: 1080x1920
max_fps: 30
audio_codec: aac
audio_bitrate: 128k
container: mp4
faststart: true
```
**Identical to Instagram preset.** They share the same technical requirements. Separated as distinct presets for discoverability — users searching for "tiktok" should find a preset immediately.

#### `youtube` -- YouTube Upload
```yaml
key: youtube
name: YouTube Upload
description: "High bitrate source for YouTube's re-encoder."
category: platform
icon: "▶️"
tags: [platform, youtube, upload, source]
video_codec: h264
crf: 16                      # intentionally high quality
speed_preset: slow
max_resolution: ""           # keep original (YouTube handles all resolutions)
max_fps: 0                   # keep original
audio_codec: aac
audio_bitrate: 256k
container: mp4
faststart: true
```
**Why CRF 16 (high quality)?** YouTube re-encodes everything. Giving YouTube a higher-quality source results in a better-looking final video on the platform. This preset is about optimizing the source for YouTube's encoder, not about shrinking the file. The file may be larger than the source if the source was heavily compressed.

**Note:** This preset is the only one that may produce a LARGER file than the source. shrinkray should note this to the user.

### 8.5 Preset Lookup and Resolution

Presets are looked up by key. The lookup algorithm:

1. **Exact match:** `"discord"` → discord preset
2. **Case-insensitive match:** `"Discord"` → discord preset
3. **Tag match:** if no key matches, search tags (e.g., `"gaming"` matches discord's tags)
4. **Fuzzy substring:** `"disc"` → discord preset (first match)

This allows flexible usage:
```bash
shrinkray video.mp4 --preset balanced
shrinkray video.mp4 --preset discord
shrinkray video.mp4 --preset POTATO
shrinkray video.mp4 --preset 4k          # fuzzy matches "4k-to-1080"
```

### 8.6 Custom Presets

Users can save custom presets to `~/.config/shrinkray/custom-presets.yaml`:

```yaml
presets:
  my-archive:
    name: "My Archive Preset"
    description: "Custom archive with AV1"
    video_codec: av1
    crf: 30
    speed_preset: "6"
    audio_codec: opus
    audio_bitrate: 96k

  family-videos:
    name: "Family Videos"
    description: "720p H.265 for family video library"
    video_codec: h265
    crf: 24
    max_resolution: 1280x720
    audio_codec: aac
    audio_bitrate: 128k
```

Custom presets override built-in presets if they share the same key. They appear in the TUI preset grid alongside built-in presets, in a separate "Custom" section.

Custom presets can be managed via CLI:
```bash
shrinkray presets                              # list all (built-in + custom)
shrinkray presets show my-archive              # show details
```

Or via the TUI: the advanced options screen (Screen 5) has a "Save as Preset" button that saves the current settings as a custom preset.

---

## 9. FFmpeg Engine

### 9.1 FFmpeg Command Building

The engine translates a `Preset` + `EncodeOptions` + source metadata into an FFmpeg argument array. The logic:

1. Start with `-y` (overwrite output, since we control output naming ourselves) and `-hide_banner`
2. Add `-i <input_path>`
3. Add `-progress pipe:1 -nostats` for machine-readable progress on stdout
4. Build the video codec arguments based on preset + detected hardware:
   - Codec selection: `h264` → `libx264` (or `h264_nvenc` etc. if HW accel)
   - CRF: `-crf <value>` (or `-cq <value>` for NVENC, `-global_quality <value>` for QSV, `-q:v <value>` for VideoToolbox)
   - Speed preset: `-preset <speed>` (software) or `-preset p<1-7>` (NVENC)
   - Tune: `-tune <tune>` if specified
   - Resolution: `-vf "scale='min(<width>,iw)':-2:flags=lanczos"` (never upscale, ensure even dimensions)
   - Framerate: `-r <fps>` or `-vf fps=<fps>` if specified
   - Pixel format: `-pix_fmt yuv420p` (default for compatibility)
5. Build the audio arguments:
   - `copy`: `-c:a copy`
   - `aac`: `-c:a aac -b:a <bitrate>` (+ `-ac <channels>` + `-ar <samplerate>` if specified)
   - `opus`: `-c:a libopus -b:a <bitrate>` (requires MKV or WebM container)
   - `none`: `-an`
6. Add container-specific flags:
   - MP4: `-movflags +faststart`
   - H.265 in MP4: `-tag:v hvc1`
7. Add metadata handling:
   - Default: `-map_metadata 0` (copy metadata)
   - Strip: `-map_metadata -1`
8. Add the output path (always the `.tmp` path during encoding)

For **two-pass encoding** (target-size presets or when `--two-pass` is set):
- Pass 1: all the above but replace CRF with `-b:v <calculated_bitrate>k`, add `-pass 1 -an -f null /dev/null` (or `NUL` on Windows in cmd, but `/dev/null` works in Git Bash on Windows)
- Pass 2: all the above with `-b:v <calculated_bitrate>k -pass 2`
- Clean up passlog files (`ffmpeg2pass-0.log`, `ffmpeg2pass-0.log.mbtree`) after encoding

### 9.2 Resolution Scaling Logic

The scale filter must handle several edge cases:

1. **Never upscale.** If the source is 720p and the preset says 1080p, keep 720p.
2. **Maintain aspect ratio.** Use `-2` for the auto-calculated dimension to ensure it's divisible by 2.
3. **Handle both landscape and portrait.** The `min()` expression should check both width and height.

The scale filter template:
```
scale='min(<target_w>,iw)':'min(<target_h>,ih)':force_original_aspect_ratio=decrease,pad=ceil(iw/2)*2:ceil(ih/2)*2:flags=lanczos
```

For portrait/vertical presets (Instagram, TikTok), the logic is different:
```
scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2:color=black
```

### 9.3 FFprobe Metadata Extraction

The `probe.go` module runs FFprobe and parses the output into typed Go structs:

```go
type VideoInfo struct {
    // File-level
    FilePath    string
    FileName    string
    FileSize    int64          // bytes
    Duration    time.Duration
    FormatName  string         // "mov,mp4,m4a,3gp,3g2,mj2"
    BitRate     int64          // overall bitrate in bps

    // Video stream (first video stream)
    VideoCodec      string     // "h264", "hevc", "av1", "vp9"
    VideoCodecLong  string     // "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10"
    VideoProfile    string     // "High", "Main", etc.
    Width           int
    Height          int
    PixelFormat     string     // "yuv420p", "yuv420p10le"
    ColorSpace      string     // "bt709", "bt2020nc"
    ColorTransfer   string     // "bt709", "smpte2084" (PQ = HDR10)
    ColorPrimaries  string     // "bt709", "bt2020"
    FrameRate       float64    // 23.976, 29.97, 30, 60
    FrameRateRaw    string     // "24000/1001"
    VideoBitRate    int64      // video stream bitrate in bps
    FrameCount      int        // nb_frames if available
    IsHDR           bool       // computed from color metadata
    HDRType         string     // "HDR10", "HLG", "Dolby Vision", ""

    // Audio stream (first audio stream)
    AudioCodec      string     // "aac", "opus", "mp3", "flac"
    AudioBitRate    int64      // audio stream bitrate in bps
    AudioChannels   int        // 1, 2, 6
    AudioChannelLayout string  // "mono", "stereo", "5.1"
    AudioSampleRate int        // 44100, 48000

    // Summary
    HasVideo    bool
    HasAudio    bool
    HasSubtitles bool
    StreamCount int
}
```

HDR detection logic:
- `color_transfer == "smpte2084"` → HDR10 (PQ curve)
- `color_transfer == "arib-std-b67"` → HLG
- `color_primaries == "bt2020"` → wide color gamut (strong HDR indicator)
- `side_data_list` contains "Mastering display metadata" → HDR10 static metadata
- `side_data_list` contains "DOVI configuration record" → Dolby Vision

Framerate parsing: FFprobe returns fractions like `"24000/1001"`. Parse by splitting on `/`, converting both parts to float64, and dividing. Handle the edge case where it's already a plain number.

---

## 10. Hardware Acceleration

### 10.1 Detection Strategy

At startup (or when first needed), shrinkray probes for available hardware encoders. The detection is NOT just checking `ffmpeg -encoders` output — that only shows what FFmpeg was compiled with, not what hardware is actually present. Instead, shrinkray performs a real test encode for each candidate:

```bash
ffmpeg -f lavfi -i "nullsrc=s=256x256:d=1:r=1" -frames:v 1 -c:v <encoder_name> -f null -
```

If this command exits with code 0, the encoder is actually functional. If it fails, the hardware is not present or drivers are missing.

### 10.2 Encoder Candidates by Platform

**Windows:**
1. `h264_nvenc`, `hevc_nvenc`, `av1_nvenc` (NVIDIA GPU)
2. `h264_amf`, `hevc_amf`, `av1_amf` (AMD GPU)
3. `h264_qsv`, `hevc_qsv`, `av1_qsv` (Intel iGPU)

**macOS:**
1. `h264_videotoolbox`, `hevc_videotoolbox` (Apple Silicon / Intel Mac)

**Linux:**
1. `h264_nvenc`, `hevc_nvenc`, `av1_nvenc` (NVIDIA GPU)
2. `h264_vaapi`, `hevc_vaapi`, `av1_vaapi` (Intel/AMD via VAAPI)
3. `h264_qsv`, `hevc_qsv`, `av1_qsv` (Intel via QSV)

### 10.3 Selection Priority

When hardware acceleration is enabled (default), shrinkray tries encoders in this priority order for each codec family:

**H.264:** NVENC → VideoToolbox → QSV → AMF → VAAPI → libx264
**H.265:** NVENC → VideoToolbox → QSV → AMF → VAAPI → libx265
**AV1:** NVENC → QSV → VAAPI → libsvtav1
**VP9:** VAAPI → QSV → libvpx-vp9

### 10.4 Hardware vs. Software Quality Note

Hardware encoders are 5-20x faster but produce approximately 15-25% larger files at equivalent visual quality. shrinkray shows the user which encoder is being used in the preview screen and encoding progress screen, along with a note:

```
Encoder: NVIDIA NVENC HEVC (GPU)
Note: Hardware encoding is faster but slightly less efficient than software.
```

The user can force software encoding with `--no-hw-accel` if they prefer smaller files over encoding speed.

### 10.5 CRF/Quality Mapping Across Encoders

Different encoders use different quality parameters. shrinkray maps the preset's CRF value to the appropriate parameter for each encoder:

| Encoder | Parameter | CRF 20 equivalent | CRF 26 equivalent | CRF 32 equivalent |
|---------|-----------|-------------------|-------------------|-------------------|
| libx264 | `-crf 20` | 20 | 26 | 32 |
| libx265 | `-crf 20` | 20 | 26 | 32 |
| h264_nvenc | `-cq 20` | 20 | 24 | 28 |
| hevc_nvenc | `-cq 20` | 20 | 24 | 28 |
| h264_videotoolbox | `-q:v 70` | 70 | 55 | 40 |
| hevc_videotoolbox | `-q:v 70` | 70 | 55 | 40 |
| h264_qsv | `-global_quality 20` | 20 | 24 | 28 |
| libsvtav1 | `-crf 25` | 25 | 32 | 38 |

These mappings are approximate and aim for visually equivalent quality across encoders.

### 10.6 Caching

Hardware detection results are cached in memory for the duration of the process. They are NOT cached to disk because hardware availability can change between sessions (e.g., external GPU connected/disconnected, driver updates).

---

## 11. Progress Tracking

### 11.1 FFmpeg Progress Output

shrinkray uses FFmpeg's `-progress pipe:1` flag to get structured progress data on stdout. The output is key=value pairs in blocks:

```
frame=120
fps=30.00
stream_0_0_q=28.0
bitrate=1500.2kbits/s
total_size=1234567
out_time_us=4000000
out_time_ms=4000000
out_time=00:00:04.000000
dup_frames=0
drop_frames=0
speed=2.50x
progress=continue
```

Each block ends with `progress=continue` (encoding in progress) or `progress=end` (encoding finished).

### 11.2 Progress Parsing

A dedicated goroutine reads FFmpeg's stdout line by line, accumulates key-value pairs into a map, and emits a `ProgressUpdate` struct when a complete block is received:

```go
type ProgressUpdate struct {
    Percent     float64       // 0.0 to 100.0
    Frame       int           // current frame number
    FPS         float64       // encoding frames per second
    Speed       float64       // e.g., 2.5 = 2.5x realtime
    Bitrate     string        // e.g., "1500.2kbits/s"
    CurrentSize int64         // bytes written so far
    ETA         time.Duration // estimated time remaining
    Elapsed     time.Duration // time since encoding started
    Done        bool          // true when progress=end
    Err         error         // non-nil if FFmpeg exited with error
}
```

**Percentage calculation:** `percent = (out_time_us / 1000000) / total_duration_seconds * 100`
The total duration is obtained from FFprobe before encoding starts.

**ETA calculation using exponential moving average:**
```go
rawETA = (totalDuration - currentTime) / speed
smoothedETA = alpha * rawETA + (1-alpha) * previousSmoothedETA  // alpha = 0.3
```
The smoothing prevents jittery ETA display.

### 11.3 TUI Integration

The progress goroutine sends updates to the TUI via a buffered channel (size 1). The TUI consumes updates via a `tea.Cmd` that reads from the channel:

```go
func waitForProgress(ch <-chan ProgressUpdate) tea.Cmd {
    return func() tea.Msg {
        p, ok := <-ch
        if !ok {
            return progressMsg{Done: true}
        }
        return progressMsg(p)
    }
}
```

When the TUI's Update method receives a `progressMsg`, it updates the model and subscribes for the next update by returning `waitForProgress(ch)` as a command.

Non-blocking channel sends ensure the progress parser never blocks if the TUI is slow to consume:
```go
select {
case ch <- update:
default:
    // Drop old value, send new
    select { case <-ch: default: }
    ch <- update
}
```

### 11.4 Headless Mode Progress

In headless mode, progress is printed to stderr as a continuously-overwritten line:
```
[============================            ] 63.2% | 1.4x | ETA 2:18 | 487 MB
```

For piped/non-TTY output, print one line per update (not overwritten):
```
progress: 10.0% speed=1.4x eta=5:30
progress: 20.0% speed=1.5x eta=4:50
...
```

---

## 12. Batch Processing

### 12.1 Entering Batch Mode

Batch mode is activated when:
- Multiple files are passed as CLI arguments
- A directory is passed as a CLI argument (with or without `--recursive`)
- Multiple files are selected in the TUI file picker via `Space` multi-select
- A glob pattern matches multiple files

### 12.2 Job Queue

Each file becomes a `Job` in an in-memory queue:

```go
type JobStatus string

const (
    JobPending  JobStatus = "pending"
    JobRunning  JobStatus = "running"
    JobDone     JobStatus = "done"
    JobFailed   JobStatus = "failed"
    JobSkipped  JobStatus = "skipped"
)

type Job struct {
    ID          string        // deterministic hash of absolute input path
    InputPath   string
    OutputPath  string
    Status      JobStatus
    Progress    float64       // 0-100 for running jobs
    InputSize   int64
    OutputSize  int64         // populated on completion
    Error       string        // populated on failure
    Attempts    int           // number of encode attempts
    StartedAt   time.Time
    CompletedAt time.Time
}
```

### 12.3 Queue Ordering

Default: smallest files first (`--sort size-asc`). This gives fast early feedback — small files complete quickly, building user confidence. Configurable via `--sort`:
- `size-asc` (default): smallest first
- `size-desc`: largest first (free disk space quickly)
- `name`: alphabetical
- `none`: as discovered/provided

### 12.4 Parallel Encoding

Default concurrency: 2 jobs (`--jobs 2`). FFmpeg already uses multiple CPU cores internally, so running more than 2-3 concurrent encodes usually causes thrashing rather than speedup.

For hardware-accelerated encoding, higher concurrency may be beneficial since the GPU handles encoding with minimal CPU usage. Consumer NVIDIA GPUs have a session limit (historically 5, relaxed in newer drivers).

### 12.5 Queue Persistence and Resume

Before encoding starts, the queue is serialized to a JSON file: `.shrinkray-queue.json` in the output directory (or temp directory if using suffix mode).

```json
{
    "version": 1,
    "preset": "balanced",
    "created_at": "2026-03-21T10:30:00Z",
    "jobs": [
        {
            "id": "abc123",
            "input_path": "/home/user/Videos/video1.mp4",
            "output_path": "/home/user/Videos/video1_shrunk.mp4",
            "status": "done",
            "input_size": 1073741824,
            "output_size": 536870912,
            "attempts": 1
        },
        {
            "id": "def456",
            "input_path": "/home/user/Videos/video2.mp4",
            "output_path": "/home/user/Videos/video2_shrunk.mp4",
            "status": "pending",
            "input_size": 2147483648,
            "attempts": 0
        }
    ]
}
```

On startup, if a queue file exists, shrinkray offers to resume:
- TUI: "Found 38 pending jobs from a previous run. Resume? [Y/n]"
- Headless: resumes automatically (use `--no-resume` to start fresh)

Jobs marked `running` from a crashed session are reset to `pending`. Jobs marked `done` are skipped. The queue file is deleted after all jobs complete.

### 12.6 Skip-If-Already-Compressed Logic

Before encoding each file, shrinkray checks whether it should be skipped:

1. **Output exists and is valid:** If the output file exists, run FFprobe and check if its duration matches the source within 2 seconds. If so, skip.
2. **Source already optimal:** If the source already uses the target codec and its bitrate is at or below what the preset would produce (within 10% tolerance), skip. Display: "Already using H.265 at 3.2 Mbps (preset targets ~4 Mbps). Skipping."
3. **Size threshold:** If the source is below a reasonable size-per-duration threshold for its resolution (e.g., < 2 MB/min for 720p), consider it already compressed and skip.

Skipped files get `JobSkipped` status with a reason string.

### 12.7 Failed Job Retry

- Each job gets `max_attempts = 2` (one retry)
- On first failure: retry once automatically with the same settings
- If it fails again: mark as `JobFailed`, store error, move to next job
- After batch completion: display failed jobs with their error messages
- User can re-run with `--retry-failed` to attempt only the failed jobs

Before retrying, delete any partial `.tmp` output from the failed attempt.

### 12.8 Batch Completion Summary

After all jobs complete, display (in both TUI and headless modes):
```
Batch Complete
──────────────
Files processed:    53
Successful:         51
Skipped:             1  (already compressed)
Failed:              1  (see log)

Total input size:    142.3 GB
Total output size:    38.7 GB
Space saved:         103.6 GB (72.8%)

Wall time:           1h 23m
Avg compression:     72.8%
Avg speed:           2.1x realtime
```

---

## 13. File Size Targeting

### 13.1 The Math

When a preset specifies `target_size_bytes` (or the user uses `--target-size`), shrinkray calculates the required video bitrate:

```
effective_target = target_size_bytes * safety_factor  // 0.97-0.99 depending on platform
total_bitrate_bps = (effective_target * 8) / duration_seconds
video_bitrate_bps = total_bitrate_bps - audio_bitrate_bps
video_bitrate_kbps = video_bitrate_bps / 1000
```

Safety factors by platform:
| Platform | Safety Factor | Rationale |
|----------|--------------|-----------|
| Discord | 0.97 | Tight limit, rejection on exceed |
| WhatsApp | 0.97 | Tight limit |
| Email | 0.98 | Some email clients are strict |
| General | 0.98 | Safe default |
| Twitter/Telegram | 0.99 | Generous limits |

### 13.2 Two-Pass Encoding

Target-size presets always use two-pass encoding:

**Pass 1 (analysis):**
```bash
ffmpeg -y -i input.mp4 -c:v libx264 -b:v <calculated>k -pass 1 -an -f null /dev/null
```

**Pass 2 (encode):**
```bash
ffmpeg -y -i input.mp4 -c:v libx264 -b:v <calculated>k -pass 2 -c:a aac -b:a 96k -movflags +faststart output.mp4
```

In the TUI, the progress bar shows "Pass 1/2" and "Pass 2/2" with combined percentage (Pass 1 = 0-50%, Pass 2 = 50-100%).

### 13.3 Feasibility Detection

Before encoding, shrinkray checks if the target is realistic:

```
bits_per_pixel = video_bitrate / (width * height * fps)

if bits_per_pixel < 0.005:
    "IMPOSSIBLE: Target too small for this video duration and resolution."
    "Suggestions: trim the video, reduce resolution, or increase target size."

if bits_per_pixel < 0.02:
    "WARNING: Quality will be poor at this bitrate."
    "Consider reducing resolution from 1080p to 720p."

if bits_per_pixel < 0.05:
    "Quality will be acceptable with some visible compression artifacts."

if bits_per_pixel >= 0.05:
    "Target is reasonable. Good quality expected."
```

If the calculated video bitrate is negative (audio alone exceeds the target), display:
"IMPOSSIBLE: Even at 0 video bitrate, the audio track exceeds the target size. Consider reducing audio bitrate or removing audio."

### 13.4 Adaptive Resolution for Size Targets

For tight size targets, shrinkray auto-scales resolution to maintain watchable quality. The algorithm:

1. Calculate video bitrate at source resolution
2. Compute bits_per_pixel
3. If bpp < 0.02, calculate scale factor: `sqrt(0.03 / current_bpp)`
4. Apply scale factor to resolution (round to nearest even number)
5. Recalculate bitrate at new resolution
6. Repeat until bpp >= 0.02 or resolution is at 360p minimum

This auto-scaling is shown to the user in the preview screen: "Resolution will be reduced to 854x480 to maintain watchable quality at 10 MB target."

---

## 14. Smart Recommendation Engine

### 14.1 How It Works

After probing the source video, shrinkray analyzes the metadata and generates ranked preset recommendations. Each recommendation includes a human-readable reason.

### 14.2 Recommendation Rules

The engine evaluates these rules in order. Multiple rules can fire, producing a ranked list:

| Condition | Recommended Preset | Reason |
|-----------|-------------------|--------|
| Resolution >= 3840x2160 | `4k-to-1080` | "Source is 4K. Downscaling to 1080p saves ~75% with no visible loss on most screens." |
| Orientation is portrait AND duration < 90s | `tiktok`, `instagram` | "Short vertical video — perfect for social media." |
| File size > 500 MB AND duration < 10 min | `balanced` | "Large file for its duration. Balanced compression should save 50-70%." |
| File size > 2 GB | `tiny` | "Very large file. Consider aggressive compression." |
| Duration > 10 min AND file size > 100 MB | `web` | "Long video — web-optimized preset with streaming support." |
| FPS <= 15 | `slideshow` | "Low framerate suggests slideshow/presentation content." |
| Video codec is already H.265 AND bitrate is reasonable | (skip suggestion) | "Already compressed with H.265. Further shrinking may degrade quality." |
| Audio bitrate > 256kbps | (any preset) + note | "Audio bitrate is high (320kbps). Compression will reduce it to 128kbps." |
| HDR detected | note | "HDR content detected. Output will be SDR unless using a codec that preserves HDR." |

### 14.3 Presentation in TUI

On Screen 3 (video info), recommendations appear as a highlighted tip line:
```
💡 Recommended: "Balanced" — Good quality, ~60% smaller. Or "4K to 1080p" for maximum savings.
```

On Screen 4 (preset grid), recommended presets are marked with a star icon and sorted to the top.

---

## 15. TUI Screens

### 15.1 Global Layout

Every screen shares a consistent layout:

```
┌─ Header ─────────────────────────────────────────────── Help ─ Quit ─┐
│                                                                       │
│   Screen Title                                                        │
│   ────────────                                                        │
│                                                                       │
│   [ Main content area - varies per screen ]                          │
│                                                                       │
│                                                                       │
│                                                                       │
│                                                                       │
│   ── keybinding1 ── keybinding2 ── keybinding3 ── keybinding4 ──    │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

- **Header bar:** App name "SHRINKRAY" on the left (in primary color), step indicator "step 2/4" (in muted), `?` help and `q` quit on the right
- **Footer bar:** Context-sensitive keyboard shortcuts for the current screen
- **Main content area:** 100% of remaining space, controlled by the active screen

All terminal dimensions are obtained from `tea.WindowSizeMsg` and propagated to all sub-models on resize.

### 15.2 Screen 1: Splash

Displayed for 1.5 seconds or until any keypress. Shows:
- ASCII art logo in primary color (the ray gun + "shrinkray" text)
- Tagline in muted italic
- Version number and FFmpeg version
- Detected system capabilities: available encoders, GPU name (if applicable), CPU thread count
- "Press any key or drop a file..." pulsing between dim and bright on a 1-second `tea.Tick` interval

If shrinkray was launched with file arguments, this screen is skipped entirely.

### 15.3 Screen 2: File Picker

See section 6.2 for full details. The screen contains:
- A file browser (upper portion) showing the current directory, subdirectories, and video files with metadata
- A divider line "── or ──"
- A text input for direct path entry (lower portion)
- Tab to switch focus between browser and text input
- Status line at bottom: "5 videos | 10.4 GB total" (for the current directory)

Batch selection:
- `Space` toggles selection on the highlighted file (checkmark appears)
- Selected file count shown in header: "3 files selected"
- `Enter` with multiple selections → goes to Screen 9a (batch queue)
- `Enter` with single selection → goes to Screen 3 (video info)

### 15.4 Screen 3: Source Video Info

A styled card displaying all source video metadata:

```
┌─ interview_raw.mp4 ─────────────────────────────────────────────┐
│                                                                   │
│   VIDEO                            AUDIO                         │
│   ─────                            ─────                         │
│   Codec     H.264 (High)          Codec     AAC-LC              │
│   Res       1920 x 1080           Channels  Stereo              │
│   FPS       29.97                  Bitrate   320 kbps            │
│   Bitrate   18.2 Mbps             Sample    48000 Hz            │
│   Duration  12m 34s                                              │
│                                                                   │
│   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐                  │
│   │  SIZE  │ │QUALITY │ │ COLOR  │ │STREAMS │                  │
│   │ 2.4 GB │ │  High  │ │SDR 420 │ │  V+A   │                  │
│   └────────┘ └────────┘ └────────┘ └────────┘                  │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘

💡 High bitrate source — significant shrink possible. Try "Balanced" preset.

    enter continue    b back    a advanced options    p presets
```

Quality heuristic based on bits-per-pixel:
- `bpp > 0.2` → "Excessive" (amber warning)
- `bpp > 0.1` → "High"
- `bpp > 0.05` → "Medium"
- `bpp < 0.05` → "Low"

The four mini stat boxes use small Lip Gloss rounded-border boxes in a horizontal row.

### 15.5 Screen 4: Preset Selection

A 2-column grid of preset cards. Each card is a Lip Gloss box showing:
- Icon + preset name (bold)
- Estimated output size (computed from source duration + preset bitrate)
- Key settings: codec, CRF, resolution
- A mini compression bar (block characters showing kept vs. removed portions)
- Quality label ("Excellent", "Good", "Acceptable")
- Estimated savings percentage in bold

Navigation: arrow keys move between cards. Active card has a double-line border in primary color with brighter text. Inactive cards have single-line muted borders.

The grid layout:
- Row 1: Balanced (default, pre-selected), High
- Row 2: Low, Tiny
- Row 3: Web, Email
- Row 4: [Smart Recommendations from section 14]
- Row 5: Custom (opens advanced options)
- Additional rows for platform presets if relevant (shown if smart engine detects they'd be useful)

`Enter` selects the highlighted preset and goes to Screen 6 (preview).

The full preset list is scrollable if it doesn't fit the terminal height.

### 15.6 Screen 5: Advanced Options

A full-screen Huh form with grouped fields:

**Group 1: Video**
- Video Codec: Select[string] with options: H.264, H.265, AV1, VP9
- Encoder: Select[string] with options dynamically filtered by detected hardware (e.g., "Software (x265)", "NVIDIA NVENC", "Intel QSV"). Unavailable encoders shown as disabled.
- CRF / Quality: A custom slider component (0-51 range for H.264/H.265, 0-63 for AV1/VP9). As the slider moves, a label updates: "Visually Lossless" (18), "Excellent" (20), "Very Good" (23), "Good" (26), "Acceptable" (28), "Fair" (32), "Poor" (36), "Very Poor" (40+). The slider uses a color gradient: green at low CRF, amber in mid-range, red at high CRF.
- Resolution: Select with options: "Original (1920x1080)", "1080p", "720p", "480p", "360p", "Custom". Custom opens width/height text inputs with aspect ratio lock.
- Framerate: Select with options: "Original (29.97)", "60", "30", "24"

**Group 2: Audio**
- Audio Codec: Select: AAC, Opus, Copy (keep original), None (strip audio)
- Audio Bitrate: Select: 64k, 96k, 128k, 192k, 256k (hidden if codec is "copy" or "none")
- Audio Channels: Select: Original, Stereo, Mono (hidden if codec is "copy" or "none")

**Group 3: Options**
- Speed Preset: Select: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
- Two-Pass: Confirm checkbox
- Strip Metadata: Confirm checkbox (default: unchecked)
- Hardware Acceleration: Confirm checkbox (default: checked if available)

**Group 4: Confirmation**
- Note showing a summary of all selected options
- Confirm: "Apply Settings" / "Cancel"

All groups use the Huh `WithHideFunc` for conditional visibility (e.g., audio options hidden when audio codec is "none").

The Huh form's `OptionsFunc` dynamically updates quality descriptions when the codec changes.

### 15.7 Screen 6: Preview / Confirmation

Side-by-side BEFORE and AFTER cards with a central savings box:

```
┌─ BEFORE ────────────────┐    ┌─ AFTER (estimated) ──────────┐
│                          │    │                               │
│  interview_raw.mp4       │    │  interview_raw_shrunk.mp4    │
│                          │    │                               │
│  Size       2.4 GB       │ →  │  Size       ~800 MB          │
│  Codec      H.264        │    │  Codec      H.265            │
│  Res        1920x1080    │    │  Res        1280x720         │
│  Bitrate    18.2 Mbps    │    │  Bitrate    ~6.1 Mbps        │
│  Audio      AAC 320k     │    │  Audio      AAC 128k         │
│  Duration   12m 34s      │    │  Duration   12m 34s          │
│                          │    │                               │
└──────────────────────────┘    └───────────────────────────────┘

                  ┌──────────────────────┐
                  │   ESTIMATED SAVINGS  │
                  │                      │
                  │     ▼ 67%            │
                  │    1.6 GB freed      │
                  │    ETA: ~4m 12s      │
                  │                      │
                  │   Encoder: x265 (SW) │
                  └──────────────────────┘

Output: /home/user/Videos/interview_raw_shrunk.mp4  [e to edit]

       enter START ENCODING     e edit settings     b back
```

- BEFORE card: muted border
- AFTER card: primary color border
- Changed values in AFTER card: highlighted in tertiary (cyan) color
- Arrow between cards: secondary color (pink), bold
- Savings percentage: success color (green), bold, large
- ETA computed from frame count / estimated encoding speed
- Output path shown at bottom with `e` to edit inline
- The encoder being used is shown (software or which hardware encoder)

Size estimation formula:
```
estimated_size = (target_video_bitrate + audio_bitrate) * duration_seconds / 8
// Add 3% for container overhead
estimated_size *= 1.03
```

For CRF-based presets (not target-size), estimation uses a lookup table of typical bitrate-per-pixel values for each CRF level and codec.

### 15.8 Screen 7: Encoding Progress

The most visually dynamic screen:

```
Encoding                         interview_raw.mp4

███████████████████████████████░░░░░░░░░░░░░░░░░░░  63.2%

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   PROGRESS   │  │     ETA      │  │    SPEED     │
│  14,198 /    │  │   2m 18s     │  │   1.4x       │
│  22,471      │  │   remaining  │  │   42.3 fps   │
│  frames      │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  CURRENT sz  │  │  BITRATE     │  │  ELAPSED     │
│   487 MB     │  │   5.8 Mbps   │  │   3m 56s     │
│   of ~800 MB │  │   avg        │  │              │
└──────────────┘  └──────────────┘  └──────────────┘

┌─ Log ────────────────────────────────────────────────────────┐
│  [3:56] frame=14284 fps=42.1 q=28.0 size=490MB bitrate=5.8M │
└──────────────────────────────────────────────────────────────┘

           p pause    c cancel    l toggle log
```

- Progress bar uses gradient fill (primary → secondary color) via block characters
- Speed stat is color-coded: green ≥1x, amber 0.5-1x, red <0.5x
- The 6 stat boxes use rounded borders in primary color
- Log panel is a Bubbles viewport showing the last 2 lines of FFmpeg output (expandable with `l`)
- `p` pauses (sends SIGSTOP on Unix, not available on Windows — show "not available" on Windows)
- `c` prompts confirmation: "Cancel encoding? The partial file will be deleted. [y/N]"
- For two-pass encoding: progress bar label shows "Pass 1/2" (0-50%) and "Pass 2/2" (50-100%)

### 15.9 Screen 8: Completion

```
                      ✓  COMPLETE

┌─ BEFORE ────────────────┐    ┌─ AFTER ──────────────────────┐
│  interview_raw.mp4       │    │  interview_raw_shrunk.mp4    │
│  2.4 GB                  │ →  │  762 MB                      │
│  H.264  1920x1080        │    │  H.265  1280x720             │
│  18.2 Mbps               │    │  5.4 Mbps                    │
└──────────────────────────┘    └───────────────────────────────┘

         ┌──────────────────────────────────────────┐
         │                                          │
         │      ██████████                          │
         │      ██████████  ███                     │
         │      ██████████  ███                     │
         │      ██████████  ███                     │
         │        BEFORE    AFTER                   │
         │        2.4 GB    762 MB                  │
         │                                          │
         │    Saved 1.66 GB  ·  68.2% smaller       │
         │    Encoding time: 4m 02s  ·  1.4x speed │
         └──────────────────────────────────────────┘

Output: /home/user/Videos/interview_raw_shrunk.mp4

   enter new file    o open folder    r re-encode    q quit
```

- "COMPLETE" in bold success color (green) with checkmark
- Bar chart: BEFORE bar in secondary color (pink), AFTER bar in success color (green), proportional to file sizes
- Hero stat "68.2% smaller" in bold green
- `o` opens the containing folder in the system file manager (`xdg-open` on Linux, `open` on macOS, `explorer` on Windows)
- `enter` returns to Screen 2 for another file
- `r` goes back to Screen 4/5 to try different settings on the same source file

If the output is larger than the source (possible with `youtube` preset or already-efficient sources), show a warning in amber:
"⚠ Output (850 MB) is larger than source (800 MB). The source was already well-compressed."

### 15.10 Screen 9a: Batch Queue

```
Batch Queue                          Preset: Balanced (H.265)

#  File                       Size     Est. Output   Savings
── ────────────────────────── ──────── ─────────── ──────────
1  interview_raw.mp4          2.4 GB    ~800 MB      -67%
2  vacation_2024.mov          890 MB    ~290 MB      -67%
3  screen_capture.mkv         1.1 GB    ~370 MB      -66%
4  drone_footage.mp4          5.7 GB    ~1.9 GB      -67%
── ────────────────────────── ──────── ─────────── ──────────
   Total                      10.1 GB   ~3.4 GB      ~6.7 GB

┌────────────────────────────────────────────────────────────┐
│  Estimated total time: ~18 minutes                         │
│  Output: same directory with "_shrunk" suffix              │
│  On conflict: Skip existing                                │
└────────────────────────────────────────────────────────────┘

 enter START ALL   s per-file settings   x remove   p preset   b back
```

### 15.11 Screen 9b: Batch Progress

```
Batch Progress                                    2/4 done

Overall  ██████████████████████████░░░░░░░░░░░░░░░░░  52%

✓ interview_raw.mp4       2.4 GB → 762 MB     -68%    4m 02s
✓ vacation_2024.mov       890 MB → 284 MB     -68%    1m 47s
▸ screen_capture.mkv      1.1 GB → encoding   37%     ~1m 20s
  drone_footage.mp4       5.7 GB → queued             ~8m est

┌─ Current: screen_capture.mkv ────────────────────────────────┐
│  ██████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  37.1%     │
│  Frames  8,312 / 22,471   Speed  1.2x   ETA  1m 20s        │
│  Size    142 MB / ~370 MB                                    │
└──────────────────────────────────────────────────────────────┘

Elapsed: 6m 12s    Remaining: ~9m 30s    Saved so far: 2.24 GB

      p pause all    c cancel all    s skip current    l log
```

### 15.12 Screen 9c: Batch Completion

```
                   ✓  BATCH COMPLETE

File                        Before    After     Saved    Time
─────────────────────────── ──────── ──────── ──────── ────────
interview_raw.mp4           2.4 GB    762 MB   -68.2%   4m 02s
vacation_2024.mov           890 MB    284 MB   -68.1%   1m 47s
screen_capture.mkv          1.1 GB    351 MB   -68.0%   3m 31s
drone_footage.mp4           5.7 GB    1.82 GB  -68.1%   8m 14s
─────────────────────────── ──────── ──────── ──────── ────────
TOTALS                      10.1 GB   3.22 GB           17m 34s

               ┌────────────────────────────┐
               │   Total saved: 6.88 GB     │
               │   Average: 68.1% smaller   │
               └────────────────────────────┘

Output directory: /home/user/Videos/

       enter new batch    o open folder    q quit
```

### 15.13 Help Overlay

Pressing `?` on any screen displays a semi-transparent modal overlay with context-sensitive key bindings. The overlay is centered, has a muted background, and lists all keys for the current screen plus global keys. Press `?` again or `Esc` to dismiss.

---

## 16. Styling and Theming

### 16.1 Style Definitions

All Lip Gloss styles are defined in a single `styles.go` file as package-level variables, referencing a `Theme` struct:

```go
type Theme struct {
    Primary    lipgloss.Color
    Secondary  lipgloss.Color
    Tertiary   lipgloss.Color
    Success    lipgloss.Color
    Warning    lipgloss.Color
    Error      lipgloss.Color
    Muted      lipgloss.Color
    Text       lipgloss.Color
    Background lipgloss.Color
}
```

Two built-in themes are available:
- **Neon Dusk** (default): violet/pink/cyan
- **Electric Sunset**: coral/gold/orange

Theme is switchable at runtime via `Ctrl+T`. The current theme preference is saved to config.

### 16.2 Consistent Visual Elements

- All cards use `lipgloss.RoundedBorder()`
- Active/selected elements use primary color border with bold text
- Inactive elements use muted border
- Progress bars use gradient fill via Lip Gloss color interpolation
- File sizes are always rendered in secondary (pink) color — the "hero metric"
- Savings percentages are always in success (green) color
- Warnings in amber, errors in red
- Keyboard shortcuts in the footer use `key` in primary color + `description` in muted color
- All text uses the `Text` color from the theme (near-white for dark backgrounds)

### 16.3 Adaptive Color Support

Lip Gloss v2 auto-detects terminal color capabilities (TrueColor, ANSI256, ANSI16, no color) and downsamples accordingly. shrinkray uses `lipgloss.Println()` instead of `fmt.Println()` for all styled output. For light terminal backgrounds, `lipgloss.HasDarkBackground()` is checked and colors are adjusted via `LightDark()`.

### 16.4 Responsive Layout

Three layout breakpoints:
- **Wide (80+ columns):** Side-by-side cards, 2-column preset grid, full 6-box stat grid
- **Medium (60-79 columns):** Stacked cards, single-column presets, compact 3-box stats
- **Narrow (<60 columns):** Minimal single-column, abbreviated text, no borders on cards

---

## 17. Configuration System

### 17.1 Config File Location

- **Linux:** `~/.config/shrinkray/config.yaml` (XDG)
- **macOS:** `~/Library/Application Support/shrinkray/config.yaml`
- **Windows:** `%APPDATA%\shrinkray\config.yaml`

Resolved via `os.UserConfigDir()`.

### 17.2 Config File Structure

```yaml
# shrinkray configuration

# Default encoding settings
defaults:
  codec: h265              # h264, h265, av1, vp9
  preset: balanced         # default preset when none specified
  speed: medium            # encoder speed preset
  hw_accel: true           # auto-detect and use hardware acceleration

# Output settings
output:
  mode: suffix             # suffix, directory, in-place
  suffix: "_shrunk"        # suffix to append to filename
  directory: ""            # output directory (empty = same as source)
  on_conflict: ask         # ask, skip, overwrite, auto-rename
  preserve_metadata: true
  faststart: true          # add -movflags +faststart for MP4

# UI settings
ui:
  theme: neon-dusk         # neon-dusk, electric-sunset
  show_splash: true        # show splash screen on launch
  show_log: false          # show FFmpeg log by default during encoding

# Batch settings
batch:
  jobs: 2                  # concurrent encoding jobs
  sort: size-asc           # size-asc, size-desc, name, none
  auto_resume: true        # automatically resume interrupted batches
  skip_compressed: true    # skip files that appear already compressed

# FFmpeg settings
ffmpeg:
  path: ""                 # custom FFmpeg path (empty = auto-detect from PATH)
```

### 17.3 Config Loading Priority

1. Built-in defaults (hardcoded in `config/defaults.go`)
2. Config file values (override defaults)
3. CLI flags (override config file)

Missing config file is NOT an error — shrinkray uses built-in defaults.

### 17.4 Custom Presets File

Separate from the main config, stored at:
- **Linux:** `~/.config/shrinkray/custom-presets.yaml`
- **macOS:** `~/Library/Application Support/shrinkray/custom-presets.yaml`
- **Windows:** `%APPDATA%\shrinkray\custom-presets.yaml`

See section 8.6 for format.

---

## 18. Logging

### 18.1 Log Destination

In TUI mode, logs go to a file (never stdout/stderr, which are owned by Bubble Tea):
- **Linux:** `~/.cache/shrinkray/shrinkray.log`
- **macOS:** `~/Library/Caches/shrinkray/shrinkray.log`
- **Windows:** `%LOCALAPPDATA%\shrinkray\shrinkray.log`

In headless mode, logs go to stderr (standard for CLI tools).

### 18.2 Log Levels

- **Debug:** FFmpeg command lines, probe output, hardware detection details (enabled with `--verbose`)
- **Info:** Encoding start/complete, file operations, preset selection
- **Warn:** Quality warnings, skip reasons, fallback to software encoding
- **Error:** FFmpeg failures, file not found, permission errors
- **Fatal:** FFmpeg not installed, unrecoverable errors

### 18.3 Structured Logging

All log entries use structured key-value pairs via Charm Log:
```
INFO encoding started input=video.mp4 preset=balanced codec=h265 crf=26
INFO encoding complete input=video.mp4 output=video_shrunk.mp4 input_size=2.4GB output_size=762MB savings=68.2% duration=4m02s
WARN skipping file input=already_small.mp4 reason="already compressed (H.265 at 2.1 Mbps)"
ERROR encoding failed input=corrupt.mp4 error="ffmpeg exit code 1: Invalid data found when processing input"
```

---

## 19. Cross-Platform Considerations

### 19.1 Path Handling

- Always use `filepath.Join()` for constructing paths (OS-appropriate separator)
- Always use `filepath.Abs()` to resolve relative paths before passing to FFmpeg
- Use `os/exec.Command()` with separate arguments (NOT shell string interpolation) — this handles spaces in paths automatically
- Use `os.UserConfigDir()`, `os.UserCacheDir()` for platform-appropriate directories
- Handle Windows drive letters (C:) via `filepath.VolumeName()`
- The `/dev/null` output for FFmpeg pass 1 works in Git Bash on Windows; for native cmd.exe, use `NUL`. Detect the shell environment or use a temp file approach.

### 19.2 FFmpeg Process Management

- On Unix: `SIGSTOP`/`SIGCONT` for pause/resume
- On Windows: pause is not supported via signals. The TUI shows "Pause not available on Windows" when `p` is pressed. Alternative: could use Windows Job Objects, but this is a stretch goal.
- Graceful stop: send `q` to FFmpeg's stdin. This works cross-platform and produces a valid output file (FFmpeg finalizes the container).
- The 3-tier cancellation (q → SIGINT → SIGKILL) uses `os.Interrupt` for SIGINT. On Windows, `os.Interrupt` sends `CTRL_BREAK_EVENT` which may not work for non-console processes. The stdin `q` approach is the reliable cross-platform method.

### 19.3 Terminal Compatibility

- Lip Gloss handles color profile detection and downsampling automatically
- Unicode characters (block elements for progress bars, box-drawing for borders) work on all modern terminals (Windows Terminal, iTerm2, GNOME Terminal, Konsole). Legacy Windows console (`cmd.exe` without Windows Terminal) may have issues — consider ASCII fallback characters for box drawing.
- `lipgloss.EnableLegacyWindowsANSI()` for older Windows terminals

### 19.4 "Open Folder" Command

On the completion screen, `o` opens the containing folder:
- **macOS:** `open <directory>`
- **Linux:** `xdg-open <directory>`
- **Windows:** `explorer <directory>`

---

## 20. Distribution and Installation

### 20.1 Build Configuration

GoReleaser produces:
- `shrinkray_linux_amd64.tar.gz`
- `shrinkray_linux_arm64.tar.gz`
- `shrinkray_darwin_amd64.tar.gz`
- `shrinkray_darwin_arm64.tar.gz`
- `shrinkray_windows_amd64.zip`
- `shrinkray_windows_arm64.zip`
- `checksums.txt` (SHA256)

All builds use `CGO_ENABLED=0` for static binaries. Version info is injected via `-ldflags -X main.version=...`.

### 20.2 Installation Methods

**Go install:**
```bash
go install github.com/OWNER/shrinkray@latest
```

**Homebrew (macOS/Linux):**
```bash
brew install OWNER/tap/shrinkray
```
The Homebrew formula declares `depends_on "ffmpeg"` so FFmpeg is installed automatically.

**Scoop (Windows):**
```bash
scoop bucket add OWNER https://github.com/OWNER/scoop-bucket
scoop install shrinkray
```
The Scoop manifest declares `"depends": "ffmpeg"`.

**Install script (Unix):**
```bash
curl -sSfL https://raw.githubusercontent.com/OWNER/shrinkray/main/install.sh | sh
```
The script detects OS/arch, downloads the correct binary, verifies checksum, installs to `~/.local/bin` or `/usr/local/bin`, and checks for FFmpeg (printing install instructions if not found).

**Direct download:**
From GitHub Releases: `https://github.com/OWNER/shrinkray/releases/latest`

### 20.3 FFmpeg Dependency

shrinkray requires FFmpeg and FFprobe to be installed and available in PATH (or specified via `--ffmpeg-path` or config). On first run, if FFmpeg is not found, shrinkray displays:

```
Error: FFmpeg is required but not found.

Install FFmpeg for your platform:
  macOS:     brew install ffmpeg
  Windows:   winget install ffmpeg
  Ubuntu:    sudo apt install ffmpeg
  Fedora:    sudo dnf install ffmpeg
  Arch:      sudo pacman -S ffmpeg

Or specify the path: shrinkray --ffmpeg-path /path/to/ffmpeg

More info: https://ffmpeg.org/download.html
```

---

## 21. Error Handling

### 21.1 Error Categories

| Category | Example | TUI Behavior | Headless Behavior |
|----------|---------|-------------|-------------------|
| **Fatal** | FFmpeg not installed | Full-screen error, exit | Print error, exit code 1 |
| **Encoding failure** | FFmpeg exits non-zero | Show error on progress screen, offer retry | Print error, continue to next file in batch |
| **Validation** | Invalid CRF value, file not found | Inline error message below the field | Print error, exit code 1 |
| **Warning** | Output larger than input, quality will be poor | Amber warning text, proceed | Print warning to stderr, proceed |
| **Transient** | Disk full mid-encode | Error message, clean up temp file, offer retry | Print error, exit code 1 |

### 21.2 Error Display in TUI

- Inline validation errors: red text below the relevant form field
- Encoding errors: the progress screen switches to an error state with the FFmpeg error message in a red-bordered box, with options to retry, go back, or quit
- Fatal errors: a full-screen centered error message in a red-bordered box
- Warnings: a single amber line above the footer, auto-dismissing after 5 seconds

### 21.3 Temp File Cleanup

On ANY error or interruption during encoding:
1. Check if a `.tmp` output file exists
2. Delete it
3. This is implemented via `defer` in the encode function AND a signal handler for SIGINT/SIGTERM

### 21.4 FFmpeg Error Parsing

Common FFmpeg errors and user-friendly messages:
- "No such file or directory" → "Input file not found: <path>"
- "Invalid data found when processing input" → "The input file appears to be corrupt or not a valid video."
- "Could not write header" → "Cannot write output file. Check disk space and permissions."
- "Encoder not found" → "The selected encoder (<name>) is not available in your FFmpeg build."
- Exit code 255 → "FFmpeg was interrupted."

---

## 22. Feature Roadmap

All features described in this document are planned. Here is the suggested implementation order:

### Phase 1: Foundation
- [ ] Project scaffolding (Go module, directory structure, Makefile)
- [ ] FFmpeg/FFprobe detection and validation
- [ ] FFprobe metadata extraction (probe.go, probe_types.go)
- [ ] Basic FFmpeg command building (encode.go, encode_args.go)
- [ ] Progress parsing from `-progress pipe:1`
- [ ] Graceful cancellation (stdin `q` → SIGINT → SIGKILL)
- [ ] 6 quality-tier presets (lossless, high, balanced, low, tiny, potato)
- [ ] Headless mode: `shrinkray --no-tui -i video.mp4 --preset balanced`
- [ ] Config file loading (config.go, paths.go)
- [ ] Logging setup (file in TUI mode, stderr in headless)

### Phase 2: TUI Core
- [ ] Bubble Tea app scaffolding (app.go, screen routing)
- [ ] Lip Gloss theme and style definitions
- [ ] Screen 1: Splash
- [ ] Screen 2: File picker (single file selection)
- [ ] Screen 3: Source video info card
- [ ] Screen 4: Preset selection grid (6 quality presets)
- [ ] Screen 6: Preview / confirmation
- [ ] Screen 7: Encoding progress with real-time stats
- [ ] Screen 8: Completion with before/after comparison
- [ ] Global keys: q, ?, ctrl+c
- [ ] Responsive layout (80/60 col breakpoints)

### Phase 3: Full Preset Catalog
- [ ] 5 purpose-driven presets (web, email, archive, slideshow, 4k-to-1080)
- [ ] 7 platform-specific presets (discord, discord-nitro, whatsapp, twitter, instagram, tiktok, youtube)
- [ ] Two-pass encoding for target-size presets
- [ ] File size targeting math + feasibility detection
- [ ] Adaptive resolution for tight size targets
- [ ] Smart recommendation engine
- [ ] Custom preset save/load

### Phase 4: Advanced Features
- [ ] Screen 5: Advanced options (Huh form)
- [ ] Hardware acceleration detection and selection
- [ ] HW encoder quality mapping
- [ ] AV1 (SVT-AV1) and VP9 codec support
- [ ] HDR detection and handling
- [ ] Portrait/landscape handling for social media presets

### Phase 5: Batch Processing
- [ ] Multi-file selection in TUI (Space to toggle)
- [ ] Screen 9a: Batch queue overview
- [ ] Screen 9b: Batch progress display
- [ ] Screen 9c: Batch completion summary
- [ ] Parallel encoding (--jobs N)
- [ ] Queue persistence and resume
- [ ] Skip-if-already-compressed logic
- [ ] Failed job retry
- [ ] Output directory mode with structure mirroring

### Phase 6: Polish
- [ ] Theme switching (Ctrl+T)
- [ ] Help overlay (? key)
- [ ] Shell completions (bash, zsh, fish, powershell)
- [ ] `shrinkray presets` and `shrinkray probe` subcommands
- [ ] Dry-run mode (--dry-run)
- [ ] Open folder on completion (o key)
- [ ] Man page generation via Fang
- [ ] Version check / update notification

### Phase 7: Distribution
- [ ] GoReleaser configuration
- [ ] GitHub Actions CI/CD workflow
- [ ] Homebrew tap
- [ ] Scoop bucket
- [ ] Install script (bash + powershell)
- [ ] README with VHS demo GIF
- [ ] AUR package
- [ ] Docker image

---

*This specification was compiled from the research of 20 parallel AI agents analyzing the Charm ecosystem, FFmpeg capabilities, competitive landscape, UX patterns, distribution strategies, and more. It is designed to be a complete, self-contained reference for building shrinkray from scratch.*
