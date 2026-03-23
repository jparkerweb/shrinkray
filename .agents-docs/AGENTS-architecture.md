# Architecture
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Project Structure

```
shrinkray/
├── cmd/shrinkray/main.go          # Entry point — calls cli.Execute()
├── internal/
│   ├── cli/                       # CLI layer (Fang/Cobra commands)
│   │   ├── root.go                # Root command, global flags, FFmpeg detection
│   │   ├── run.go                 # Default command: TUI or headless
│   │   ├── presets.go             # `shrinkray presets` subcommand
│   │   ├── probe.go               # `shrinkray probe <file>` subcommand
│   │   ├── version.go             # `shrinkray version` subcommand
│   │   └── completion.go          # Shell completion generation (bash/zsh/fish/powershell)
│   ├── config/                    # Configuration management
│   │   ├── config.go              # Config struct, Load/Save, defaults
│   │   └── paths.go               # Cross-platform config directory resolution
│   ├── engine/                    # FFmpeg/FFprobe business logic (ZERO TUI deps)
│   │   ├── batch.go               # Batch event types (JobStarted, JobProgress, etc.)
│   │   ├── batch_persist.go       # Persist/load batch job queue to disk
│   │   ├── cancel.go              # Graceful FFmpeg termination (q → SIGINT → SIGKILL)
│   │   ├── encode.go              # FFmpeg process runner
│   │   ├── encode_args.go         # Preset → FFmpeg args translation
│   │   ├── errors.go              # FFmpeg error pattern matching + friendly messages
│   │   ├── estimate.go            # Output size estimation
│   │   ├── ffmpeg.go              # FFmpeg/FFprobe binary detection + version parsing
│   │   ├── hdr.go                 # HDR detection
│   │   ├── hwaccel.go             # HW encoder detection and selection
│   │   ├── hwaccel_probe.go       # Test-encode to verify HW encoder works
│   │   ├── job.go                 # Job struct and JobStatus enum for batch queue
│   │   ├── open.go                # Cross-platform folder opening
│   │   ├── open_unix.go           # Unix/Linux OpenFolder implementation
│   │   ├── open_windows.go        # Windows OpenFolder implementation
│   │   ├── output.go              # OutputMode, ConflictMode, OutputOptions
│   │   ├── platform.go            # OS detection helper for testing
│   │   ├── probe.go               # FFprobe wrapper
│   │   ├── probe_types.go         # Video/audio stream info structs
│   │   ├── progress.go            # FFmpeg progress output parser
│   │   ├── progress_types.go      # ProgressUpdate struct
│   │   └── target_size.go         # Two-pass bitrate calc for file size targeting
│   ├── presets/                   # Preset definitions and recommendation
│   │   ├── preset.go              # Preset struct definition
│   │   ├── quality.go             # 6 quality-tier presets
│   │   ├── purpose.go             # 5 purpose-driven presets
│   │   ├── platform.go            # 7 platform-specific presets
│   │   ├── registry.go            # Combined registry with lookup
│   │   ├── recommend.go           # Smart recommendation engine
│   │   └── custom_presets.go      # Custom preset CRUD (load/save from YAML)
│   ├── tui/                       # TUI presentation layer (Bubble Tea)
│   │   ├── app.go                 # Top-level model: screen routing, global keys
│   │   ├── keys.go                # Global key bindings
│   │   ├── screen.go              # Re-exports ScreenModel interface from style/
│   │   ├── styles.go              # Re-exports style functions from style/
│   │   ├── theme.go               # Re-exports Theme types from style/
│   │   ├── style/                 # Style subsystem
│   │   │   ├── screen.go          # ScreenModel interface definition
│   │   │   ├── styles.go          # Lip Gloss style definitions + color support
│   │   │   └── theme.go           # Theme struct and color definitions
│   │   ├── screens/               # One file per TUI screen
│   │   │   ├── splash.go          # Screen 1: ASCII logo
│   │   │   ├── filepicker.go      # Screen 2: File browser
│   │   │   ├── info.go            # Screen 3: Video metadata
│   │   │   ├── presets.go         # Screen 4: Preset selection
│   │   │   ├── advanced.go        # Screen 5: Options form
│   │   │   ├── preview.go         # Screen 6: Before/after confirm
│   │   │   ├── encoding.go        # Screen 7: Progress display
│   │   │   ├── complete.go        # Screen 8: Results
│   │   │   ├── batch_queue.go     # Screen 9a: Batch queue
│   │   │   ├── batch_progress.go  # Screen 9b: Batch progress
│   │   │   ├── batch_complete.go  # Screen 9c: Batch results
│   │   │   ├── help.go            # Help overlay data + key binding definitions
│   │   │   └── helpers.go         # Shared utility functions (formatBytes, etc.)
│   │   └── messages/messages.go   # Custom tea.Msg types
│   └── logging/logging.go         # slog setup
├── .goreleaser.yaml
├── go.mod / go.sum
├── Makefile
└── IDEA.md                        # Full product specification
```

## Key Architectural Principle: Separation of Concerns

**This is the most important rule.** The three core packages have strict dependency boundaries:

| Package | Depends On | Never Depends On |
|---------|-----------|-------------------|
| `engine/` | stdlib, go-ffprobe | tui/, presets/ (TUI libs) |
| `presets/` | stdlib | engine/, tui/ |
| `tui/` | engine/, presets/, Charm libs | — |
| `cli/` | engine/, presets/, tui/, config/ | — |

- **`engine/`** contains ALL FFmpeg/FFprobe logic. Zero dependency on Bubble Tea, Lip Gloss, or any TUI library. It can be tested independently and reused by future frontends.
- **`tui/`** is purely presentation. It calls `engine/` for video operations and displays results.
- **`presets/`** defines preset data structures and the recommendation algorithm. No deps on TUI or engine.

## Bubble Tea Model-View-Update Pattern

The app uses the Elm Architecture:

- **Model** — Go struct holding all app state
- **Update** — Receives events (keyboard, resize, custom msgs), returns updated model + commands
- **View** — Renders current state as a string for terminal display
- **Commands (`tea.Cmd`)** — Async I/O functions (probe file, run FFmpeg, detect HW) that return messages

## Screen Routing

The top-level `App` model has a `screen` enum determining which sub-model is active. Each screen has its own `Update`/`View` methods.

Screens communicate via custom `tea.Msg` types (defined in `messages/messages.go`), NOT by calling each other directly. The parent `App` model catches transition messages and switches screens.

## Dual Mode: TUI vs. Headless

- **TUI mode** (default when stdin is a terminal): Full interactive wizard
- **Headless mode** (`--no-tui` or non-TTY stdin): Simple progress line to stderr

Both modes use the same `engine/` package.

## Concurrency Model

- TUI event loop runs on the main goroutine
- FFmpeg processes spawn in separate goroutines via `tea.Cmd`
- Progress updates sent via `p.Send()` (injects `tea.Msg` from outside the event loop)
- Buffered channel (size 1) with non-blocking sends prevents progress parser stalls
- FFprobe and hardware detection also run as `tea.Cmd` goroutines
