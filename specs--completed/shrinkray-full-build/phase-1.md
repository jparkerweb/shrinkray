# Phase 1: Foundation

- **Status:** In Progress
- **Estimated Tasks:** 18
- **Goal:** Working headless encoder — `shrinkray --no-tui -i video.mp4 --preset balanced`

---

## Overview

Build the core engine, configuration system, preset foundation, and headless CLI mode. After this phase, shrinkray can compress a video file from the command line without any TUI, producing a valid output file with real-time progress on stderr. This establishes the foundational packages that all subsequent phases build upon.

---

## Prerequisites

- Go 1.23+ installed
- FFmpeg and FFprobe installed and on PATH
- IDEA.md and AGENTS.md read for full context on data structures, API signatures, and encoding parameters

---

## Tasks

### Project Setup

- [x] **Task 1.1:** Create project scaffolding — `go.mod` (module `github.com/jparkerweb/shrinkray`), directory structure (`cmd/shrinkray/`, `internal/engine/`, `internal/presets/`, `internal/config/`, `internal/logging/`, `internal/cli/`, `internal/tui/`), and `Makefile` with targets: `build`, `run`, `test`, `lint`, `clean`. Entry point: `cmd/shrinkray/main.go` calling `cli.Execute()`. Run `go mod tidy` to initialize dependencies.

### Configuration

- [x] **Task 1.2:** Create `internal/config/paths.go` — export functions `ConfigDir() (string, error)`, `CacheDir() (string, error)`, `LogDir() (string, error)` using `os.UserConfigDir()` and `os.UserCacheDir()` with `"shrinkray"` subdirectory. Handle `SHRINKRAY_CONFIG` env var override for config path.

- [x] **Task 1.3:** Create `internal/config/config.go` — define `Config` struct with sections: `Defaults` (preset, codec, crf), `Output` (mode, suffix, conflict, directory), `UI` (theme, animations), `Batch` (jobs, sort, skipExisting), `FFmpeg` (ffmpegPath, ffprobePath). Implement `DefaultConfig() *Config`, `Load(path string) (*Config, error)` using `yaml.v3`, and `Save() error`. Implement 3-tier merge: defaults → config file → CLI flag overrides.

### Logging

- [x] **Task 1.4:** Create `internal/logging/logging.go` — export `Setup(mode string, level string) error`. In TUI mode, log to file at `LogDir()/shrinkray.log` (rotating/truncate on startup). In headless mode, log to stderr. Use `charmbracelet/log` with slog compatibility. Support levels: debug, info, warn, error. Read `SHRINKRAY_LOG_LEVEL` env var.

### FFmpeg/FFprobe Detection

- [x] **Task 1.5:** Create `internal/engine/ffmpeg.go` — export `DetectFFmpeg() (*FFmpegInfo, error)` and `DetectFFprobe() (*FFprobeInfo, error)`. Check `FFMPEG_PATH`/`FFPROBE_PATH` env vars first, then `exec.LookPath`. Run `ffmpeg -version` and `ffprobe -version` to extract version strings. Return struct with `Path string`, `Version string`. If not found, return error with platform-specific install guidance (brew on macOS, winget/choco on Windows, apt on Linux).

### Video Probing

- [x] **Task 1.6:** Create `internal/engine/probe_types.go` — define `VideoInfo` struct with fields: `Path`, `Format`, `Duration`, `Size`, `Codec`, `CodecLong`, `Width`, `Height`, `Framerate` (float64), `Bitrate` (int64), `PixelFormat`, `ColorSpace`, `ColorTransfer`, `ColorPrimaries`, `IsHDR` (bool), `HDRFormat` (string), `AudioCodec`, `AudioBitrate`, `AudioChannels`, `AudioSampleRate`, `StreamCount`, `SubtitleCount`. Include computed fields: `Resolution() string` (e.g., "1920x1080"), `AspectRatio() string`, `IsPortrait() bool`.

- [x] **Task 1.7:** Create `internal/engine/probe.go` — export `Probe(ctx context.Context, path string) (*VideoInfo, error)`. Use `gopkg.in/vansante/go-ffprobe.v2` to extract metadata. Map ffprobe JSON fields to `VideoInfo` struct. Handle missing fields gracefully (default to zero values). Validate file exists before probing.

- [x] **Task 1.8:** Create `internal/engine/hdr.go` — export `DetectHDR(info *VideoInfo) (bool, string)`. Check `ColorTransfer` for "smpte2084" (HDR10/HDR10+), "arib-std-b67" (HLG). Check side data for Dolby Vision RPU. Return bool + format string ("HDR10", "HLG", "Dolby Vision", "SDR"). Integrate into `Probe()` to populate `IsHDR` and `HDRFormat`.

### Presets

- [x] **Task 1.9:** Create `internal/presets/preset.go` — define `Preset` struct with fields: `Key` (string, unique identifier), `Name`, `Description`, `Category` (enum: quality/purpose/platform), `Codec` (h264/h265/av1/vp9), `Container` (mp4/mkv/webm), `CRF` (int), `MaxBitrate` (string, e.g., "8M"), `AudioCodec` (aac/opus/copy), `AudioBitrate` (string), `AudioChannels` (int), `Resolution` (string, empty = keep source), `MaxFPS` (int, 0 = keep), `TargetSizeMB` (float64, 0 = no target), `TwoPass` (bool), `ExtraArgs` ([]string), `Tags` ([]string), `Icon` (string, emoji). Define `Category` type with constants.

- [x] **Task 1.10:** Create `internal/presets/quality.go` — define 6 quality-tier presets and register them: `lossless` (CRF 0, h264, copy audio), `ultra` (CRF 16, h265), `high` (CRF 20, h265), `balanced` (CRF 23, h265), `compact` (CRF 28, h265, 720p max), `tiny` (CRF 32, h264, 480p max, 128k audio). Each preset must have all `Preset` fields populated including tags, icon, and description.

- [x] **Task 1.11:** Create `internal/presets/registry.go` — package-level `var registry []Preset` populated by `init()`. Export: `All() []Preset`, `Lookup(query string) (Preset, bool)` with match priority: exact key match → case-insensitive key → tag match → fuzzy substring in name/description. Export `Register(p Preset)` for custom presets. Export `ByCategory(cat Category) []Preset`.

### Encoding

- [x] **Task 1.12:** Create `internal/engine/progress_types.go` — define `ProgressUpdate` struct: `Percent` (float64, 0-100), `Frame` (int64), `FPS` (float64), `Speed` (float64, e.g., 1.5x), `Bitrate` (string, e.g., "2500kbits/s"), `Size` (int64, bytes written), `TimeElapsed` (time.Duration), `ETA` (time.Duration), `Pass` (int, 0 for single-pass, 1 or 2 for two-pass), `Done` (bool), `Error` (error).

- [x] **Task 1.13:** Create `internal/engine/encode_args.go` — export `BuildArgs(opts EncodeOptions) []string`. Define `EncodeOptions` struct: `Input` (string), `Output` (string), `Preset` (presets.Preset), `HWEncoder` (string, empty = software), `CRFOverride` (*int), `ResolutionOverride` (string), `Pass` (int, 0/1/2), `PassLogFile` (string), `ExtraArgs` ([]string). Build FFmpeg arg array: `-i input`, codec flags, CRF, resolution scaling (`-vf scale=W:H:force_original_aspect_ratio=decrease`), audio settings, `-progress pipe:1`, `-y`, output path. Handle container-specific settings (mp4: `-movflags +faststart`, webm: specific flags).

- [x] **Task 1.14:** Create `internal/engine/progress.go` — export `ParseProgress(reader io.Reader, duration time.Duration) <-chan ProgressUpdate`. Parse FFmpeg's `-progress pipe:1` key=value output (keys: `frame`, `fps`, `bitrate`, `total_size`, `out_time_us`, `progress`). Calculate percent from `out_time_us / duration`. Calculate ETA from elapsed time and percent. Send on buffered channel (size 1, non-blocking send to prevent TUI stalls). Send final update with `Done: true` when `progress=end`.

- [x] **Task 1.15:** Create `internal/engine/encode.go` — export `Encode(ctx context.Context, opts EncodeOptions) (<-chan ProgressUpdate, error)`. Build args via `BuildArgs()`, start `exec.CommandContext` with `cmd.Stdin` pipe (for `q` cancel), stdout piped to `ParseProgress()`, stderr captured for error reporting. Return progress channel. On context cancellation, write `q` to stdin, wait 3s, then kill process. Define `EncodeError` type wrapping FFmpeg stderr output.

- [x] **Task 1.16:** Create `internal/engine/cancel.go` — export `Cancel(proc *os.Process) error`. Cross-platform graceful shutdown: write `q` to FFmpeg stdin pipe → wait 3 seconds → `proc.Signal(os.Interrupt)` on Unix / `proc.Kill()` on Windows → wait 2 seconds → `proc.Kill()`. Return error if process didn't exit cleanly.

### Output & Temp Files

- [x] **Task 1.17:** Create `internal/engine/output.go` — export `ResolveOutput(input string, opts OutputOptions) (string, error)`. Define `OutputOptions`: `Mode` (suffix/directory/explicit/inplace), `Suffix` (default "_shrunk"), `Directory` (string), `ExplicitPath` (string), `ConflictMode` (skip/overwrite/autorename). For suffix mode: insert suffix before extension. For directory mode: mirror relative path structure. Handle conflict detection (file exists check). Export `TempPath(output string) string` returning `output + ".shrinkray.tmp"`. Temp file cleanup: caller uses `defer os.Remove(tempPath)` pattern; on success, `os.Rename(temp, output)`.

### CLI

- [x] **Task 1.18:** Create `internal/cli/root.go` — define root Fang command with global flags: `--config` (string), `--log-level` (string), `--no-color` (bool). Create `internal/cli/run.go` — define `run` as default command with flags: `--input`/`-i` (string, required in headless), `--preset`/`-p` (string), `--output`/`-o` (string), `--no-tui` (bool), `--crf` (int), `--codec` (string), `--resolution` (string), `--suffix` (string). Create `internal/cli/version.go` — `version` subcommand printing version/commit/date (injected via ldflags, hardcoded defaults for dev). In headless mode (`--no-tui`): validate input exists, probe video, lookup preset, build encode options, run encode, print progress to stderr as updating line (`\r` overwrite with percent/speed/ETA), print final summary to stdout. Create `cmd/shrinkray/main.go` calling `cli.Execute()`.

### Phase Testing

- [x] **Task 1.19:** Write unit tests for Phase 1 packages. **engine/**: test `BuildArgs()` produces correct FFmpeg arg arrays for various preset configurations (h264, h265, with/without resolution scaling, with/without audio copy); test `ParseProgress()` with sample FFmpeg progress output; test `ResolveOutput()` for suffix/directory/conflict modes; test `DetectHDR()` with known color transfer values. **presets/**: test `All()` returns 6 presets; test `Lookup()` exact, case-insensitive, tag, and fuzzy matches; test each quality preset has valid required fields. **config/**: test `DefaultConfig()` returns non-nil with expected defaults; test `Load()` with valid/invalid YAML. Target: `go test ./internal/engine/... ./internal/presets/... ./internal/config/...` all pass.

---

## Acceptance Criteria

- `go build ./cmd/shrinkray/` compiles without errors
- `shrinkray version` prints version info
- `shrinkray --no-tui -i test.mp4 --preset balanced` compresses a video file, shows progress on stderr, produces valid output
- `shrinkray --no-tui -i test.mp4 --preset lossless` produces a lossless copy
- Output file uses `_shrunk` suffix by default (e.g., `test_shrunk.mp4`)
- Temp file is cleaned up on success and on cancel (Ctrl+C)
- FFmpeg/FFprobe not found produces helpful install guidance message
- `go test ./internal/engine/... ./internal/presets/... ./internal/config/...` passes
- All 6 quality-tier presets are registered and lookupable by key

---

## Notes

- The `tui/` directory is created but empty — populated in Phase 2
- Fang may need to be swapped for Cobra directly if subcommand features are insufficient (low risk per PLAN-DRAFT)
- `EncodeOptions` struct will be extended in later phases (HW accel in Phase 4, batch fields in Phase 5)
- Two-pass encoding is deferred to Phase 3 (target-size presets)

---

## Phase Completion Summary

| Field | Value |
|-------|-------|
| Date completed | 2026-03-21 |
| Implementer | Claude Opus 4.6 |
| What was done | Implemented all 18 Phase 1 tasks: project scaffolding, config system (paths, config with YAML, 3-tier merge), logging (charmbracelet/log with slog, TUI/headless modes), FFmpeg/FFprobe detection with platform install guidance, video probing via go-ffprobe.v2, HDR detection (HDR10/HLG/Dolby Vision), 6 quality-tier presets with registry and multi-strategy lookup, encoding pipeline (BuildArgs, ParseProgress, Encode with graceful cancellation), output path resolution with conflict handling and temp files, and CLI via Fang/Cobra with headless mode and version subcommand. All 37 unit tests pass. |
| Files changed | go.mod, go.sum, Makefile, cmd/shrinkray/main.go, internal/cli/{root,run,version}.go, internal/config/{paths,config}.go, internal/config/config_test.go, internal/logging/logging.go, internal/engine/{ffmpeg,probe_types,probe,hdr,progress_types,encode_args,progress,encode,cancel,output,platform}.go, internal/engine/{encode_args,progress,output,hdr}_test.go, internal/presets/{preset,quality,registry}.go, internal/presets/presets_test.go, internal/tui/.gitkeep |
| Issues encountered | (1) Fang v1.0.0 has transitive dependency conflicts between x/cellbuf and x/ansi requiring Go 1.25+ and `go get -u ./...` to resolve; (2) go-ffprobe.v2 Format.DurationSeconds is the field (not Duration which is a method); (3) Streams are returned as `*Stream` pointers via helper methods like FirstVideoStream(); (4) Progress parser needed blocking send for final Done update to prevent it being dropped |
