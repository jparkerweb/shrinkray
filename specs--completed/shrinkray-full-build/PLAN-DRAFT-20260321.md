# PLAN-DRAFT: shrinkray Full Build

- **Created:** 2026-03-21
- **Status:** Complete
- **Confidence:** 92% (Requirements 24/25, Feasibility 22/25, Integration 23/25, Risk 23/25)
- **Conversation Log:** [PLAN-CONVERSATION-20260321.md](./PLAN-CONVERSATION-20260321.md)

---

## 1. Executive Summary

Build shrinkray, a cross-platform CLI video compression tool powered by FFmpeg with a wizard-style TUI built on Go's Charm ecosystem (Bubble Tea v2, Lip Gloss v2, Huh v2). The project spans 7 implementation phases: foundation engine with headless mode, TUI core with 8 screens, full 18-preset catalog with smart recommendations, advanced features (hardware acceleration, advanced options form), batch processing with queue management, UX polish, and distribution packaging. Total scope: 80 tasks across 7 packages.

---

## 2. Requirements

### 2.1 Functional Requirements

- [ ] FR-1: FFmpeg/FFprobe runtime detection and validation with platform-specific install guidance
- [ ] FR-2: Video metadata extraction via FFprobe (codec, resolution, framerate, bitrate, HDR, audio streams)
- [ ] FR-3: FFmpeg command building from preset + options + source metadata
- [ ] FR-4: Real-time progress parsing from FFmpeg `-progress pipe:1` output
- [ ] FR-5: Graceful FFmpeg process cancellation (stdin `q` → SIGINT → SIGKILL)
- [ ] FR-6: 18 built-in presets across 3 categories (quality, purpose, platform)
- [ ] FR-7: Preset lookup by key, case-insensitive, tag match, and fuzzy substring
- [ ] FR-8: Custom preset save/load from YAML config
- [ ] FR-9: Smart recommendation engine analyzing source metadata to suggest presets
- [ ] FR-10: Hardware acceleration auto-detection via real test-encode probing
- [ ] FR-11: HW encoder quality mapping (CRF → NVENC/QSV/VideoToolbox equivalents)
- [ ] FR-12: Two-pass encoding for file-size targeting with feasibility detection
- [ ] FR-13: Adaptive resolution scaling for tight size targets
- [ ] FR-14: HDR detection (HDR10, HLG, Dolby Vision)
- [ ] FR-15: Interactive TUI with 11 screens (splash → filepicker → info → presets → advanced → preview → encoding → complete + 3 batch screens)
- [ ] FR-16: Bubble Tea MVU architecture with screen routing and inter-screen messaging
- [ ] FR-17: Two color themes (Neon Dusk, Electric Sunset) switchable at runtime via Ctrl+T
- [ ] FR-18: Responsive layout at 3 breakpoints (80+, 60-79, <60 columns)
- [ ] FR-19: Headless CLI mode with `--no-tui` and all encoding flags
- [ ] FR-20: Batch processing with job queue, parallel encoding, persistence, resume, skip-if-compressed
- [ ] FR-21: Output modes: suffix (default), output directory with structure mirroring, explicit path, in-place with verification
- [ ] FR-22: Output conflict handling: ask (TUI), skip, overwrite, auto-rename
- [ ] FR-23: Temp file strategy with cleanup on cancel/failure
- [ ] FR-24: Configuration system (YAML) with 3-tier priority (defaults → config file → CLI flags)
- [ ] FR-25: Structured logging (file in TUI mode, stderr in headless)
- [ ] FR-26: CLI subcommands: `presets`, `probe`, `version`, `help` via Fang/Cobra
- [ ] FR-27: Shell completions (bash, zsh, fish, powershell)
- [ ] FR-28: Help overlay (?) with context-sensitive key bindings
- [ ] FR-29: Open folder on completion (platform-specific file manager)
- [ ] FR-30: Dry-run mode
- [ ] FR-31: GoReleaser cross-platform builds, Homebrew tap, Scoop bucket, install script

### 2.2 Non-Functional Requirements

- [ ] NFR-1: Cross-platform support (Windows, macOS, Linux) as single static binary (CGO_ENABLED=0)
- [ ] NFR-2: Zero runtime dependencies beyond FFmpeg/FFprobe
- [ ] NFR-3: Strict separation of concerns: engine/ (no TUI deps), tui/ (presentation only), presets/ (data only)
- [ ] NFR-4: Non-blocking progress channel (buffered size 1) to prevent TUI stalls
- [ ] NFR-5: Safe temp file handling — source file NEVER modified during encoding
- [ ] NFR-6: Adaptive terminal color support (TrueColor → ANSI256 → ANSI16 → no color)
- [ ] NFR-7: In-place mode uses atomic rename with FFprobe verification before replacing source
- [ ] NFR-8: Portrait/landscape handling for social media presets (padding, aspect ratio)

### 2.3 Out of Scope

- Video trimming / cutting
- Audio-only extraction
- Subtitle extraction or burn-in
- Network streaming (RTMP, HLS output)
- GUI or web frontend
- Bundling FFmpeg with the binary
- `**` glob expansion (use `--recursive` + directory args instead)

### 2.4 Testing Strategy

| Aspect | Choice | Rationale |
|--------|--------|-----------|
| Types | Unit + Integration | Unit for engine/presets logic; integration for FFmpeg arg building and progress parsing |
| Phase Testing | Yes, after each phase | Catch regressions as layers build |
| Coverage | Moderate (~60-80%) | Focus on engine/, presets/, config/ — testable pure logic |
| TUI Testing | Not directly tested | Bubble Tea models are hard to unit test with good ROI |

---

## 3. Tech Stack

| Category | Technology | Version | Justification |
|----------|-----------|---------|---------------|
| Language | Go | 1.23.0+ | Cross-compilation, static binaries, Charm ecosystem |
| TUI Framework | Bubble Tea | v2 | Model-View-Update architecture for terminal apps |
| TUI Components | Bubbles | v2 | Pre-built components (filepicker, progress, list, table, viewport, etc.) |
| TUI Styling | Lip Gloss | v2 | Terminal styling, borders, layout, color profiles |
| Forms | Huh | v2 | Interactive forms for advanced options screen |
| CLI Framework | Fang (Cobra) | latest | Subcommands, flags, shell completions, man pages |
| Logging | Charm Log | latest | slog-compatible, TUI-safe file output |
| FFprobe Parser | go-ffprobe | v2 | FFprobe JSON output → Go structs |
| Config Parser | YAML v3 | v3 | Configuration file parsing |
| Build | GoReleaser | latest | Cross-platform builds, GitHub Releases, Homebrew/Scoop |
| Lint | golangci-lint | latest | Standard Go linting |
| Dev Automation | Make | -- | Build/run/test/lint shortcuts |

**Removed from spec:** `doublestar` library — unnecessary. `filepath.WalkDir()` + extension filtering + `--recursive` flag covers all practical recursive file discovery use cases.

---

## 4. Architecture

### 4.1 Pattern

**Layered Architecture with Message-Passing UI (Elm Architecture)**

Three clean layers with downward-only dependencies:
- **Presentation:** `cli/` (Fang commands), `tui/` (Bubble Tea screens) — user interaction only
- **Business Logic:** `engine/` — all FFmpeg/FFprobe operations, zero TUI dependency
- **Data:** `presets/`, `config/`, `logging/` — pure data definitions and configuration

The TUI uses Bubble Tea's Model-View-Update pattern internally. Screens communicate via typed `tea.Msg` values through the parent App model, never by direct function calls.

**Rationale:** The spec requires headless mode reusing the same engine as TUI, presets independent of both layers, and potential future frontends. This separation is essential, not premature.

### 4.2 System Context Diagram

```
                    ┌─────────────┐
                    │   cmd/      │  Entry point
                    │  main.go    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │    cli/     │  Flag parsing, mode selection
                    │  (Fang)    │
                    └──┬─────┬───┘
                       │     │
            ┌──────────▼┐  ┌▼───────────┐
            │   tui/    │  │  Headless   │  Presentation
            │(BubbleTea)│  │  (cli/)     │
            └─────┬─────┘  └──────┬─────┘
                  │               │
                  └───────┬───────┘
                          │
                   ┌──────▼──────┐
                   │   engine/   │  FFmpeg/FFprobe operations
                   └──────┬──────┘        │
                          │          ┌────▼─────────┐
              ┌───────────┼──────┐   │ FFmpeg proc  │
              │           │      │   │ FFprobe proc │
        ┌─────▼────┐ ┌───▼────┐ │   └──────────────┘
        │ presets/  │ │config/ │ │
        └──────────┘ └────────┘ │
                          ┌─────▼────┐
                          │ logging/ │
                          └──────────┘
```

### 4.3 Components

| Component | Responsibility | Dependencies |
|-----------|---------------|-------------|
| `cmd/shrinkray/main.go` | Entry point, calls cli.Execute() | cli |
| `internal/cli/root.go` | Root command, global flags, FFmpeg detection | config, engine |
| `internal/cli/run.go` | TUI vs headless mode selection | tui, engine |
| `internal/cli/presets.go` | `presets` subcommand | presets |
| `internal/cli/probe.go` | `probe` subcommand | engine |
| `internal/cli/version.go` | `version` subcommand | -- |
| `internal/engine/probe.go` | FFprobe wrapper → VideoInfo | go-ffprobe |
| `internal/engine/probe_types.go` | VideoInfo struct | -- |
| `internal/engine/hdr.go` | HDR detection from color metadata | -- |
| `internal/engine/encode.go` | FFmpeg process runner | engine/* |
| `internal/engine/encode_args.go` | Preset + options → FFmpeg arg array | presets |
| `internal/engine/progress.go` | Parse `-progress pipe:1` output | -- |
| `internal/engine/progress_types.go` | ProgressUpdate struct | -- |
| `internal/engine/hwaccel.go` | HW encoder detection + selection | -- |
| `internal/engine/hwaccel_probe.go` | Test-encode verification | -- |
| `internal/engine/cancel.go` | Graceful FFmpeg termination | -- |
| `internal/engine/estimate.go` | Output size estimation | presets |
| `internal/engine/target_size.go` | Two-pass bitrate calculation | -- |
| `internal/presets/preset.go` | Preset struct definition | -- |
| `internal/presets/quality.go` | 6 quality-tier presets | -- |
| `internal/presets/purpose.go` | 5 purpose-driven presets | -- |
| `internal/presets/platform.go` | 7 platform-specific presets | -- |
| `internal/presets/registry.go` | Preset lookup (exact, fuzzy, tag) | -- |
| `internal/presets/recommend.go` | Smart recommendation engine | engine (VideoInfo type) |
| `internal/config/config.go` | Config load/save/merge/defaults | yaml.v3 |
| `internal/config/paths.go` | Platform config/cache/log dirs | -- |
| `internal/config/custom_presets.go` | Custom preset CRUD | yaml.v3 |
| `internal/tui/app.go` | Top-level model, screen routing, header/footer | tui/screens, tui/messages |
| `internal/tui/keys.go` | Global key bindings | -- |
| `internal/tui/styles.go` | Lip Gloss style definitions | -- |
| `internal/tui/theme.go` | Theme struct + switching logic | -- |
| `internal/tui/screens/*.go` | 11 screen models (splash through batch_complete) | engine, presets |
| `internal/tui/messages/messages.go` | Inter-screen tea.Msg types | -- |
| `internal/logging/logging.go` | slog setup (file vs stderr) | charm log |

### 4.4 Data Model

No database. Key data structures (all defined in IDEA.md):

- **`presets.Preset`** — ~20 fields defining encoding settings (codec, CRF, resolution, audio, target size, container, etc.)
- **`engine.VideoInfo`** — ~25 fields of probed source metadata (codec, resolution, framerate, bitrate, HDR, audio)
- **`engine.EncodeOptions`** — merged preset + user overrides + HW accel selection + input/output paths
- **`engine.ProgressUpdate`** — real-time stats (percent, frame, FPS, speed, bitrate, size, ETA, elapsed)
- **`engine.Job`** — batch queue item (ID, paths, status, progress, sizes, error, attempts, timestamps)
- **`config.Config`** — YAML-backed preferences (defaults, output, UI, batch, FFmpeg sections)

### 4.5 API Design

```go
// engine — public API
func Probe(ctx context.Context, path string) (*VideoInfo, error)
func Encode(ctx context.Context, opts EncodeOptions) (<-chan ProgressUpdate, error)
func Cancel(proc *os.Process) error
func DetectHardware(ctx context.Context) ([]HWEncoder, error)
func EstimateSize(info *VideoInfo, preset presets.Preset) int64
func CalculateBitrate(targetBytes int64, duration time.Duration, audioBitrate int64) int64

// presets — public API
func All() []Preset
func Lookup(query string) (Preset, bool)
func Recommend(info *engine.VideoInfo) []Recommendation

// config — public API
func Load(path string) (*Config, error)
func (c *Config) Save() error
func DefaultConfig() *Config
func ConfigDir() (string, error)
func CacheDir() (string, error)
```

---

## 5. Implementation Phases

### Phase 1: Foundation
**Goal:** Working headless encoder — `shrinkray --no-tui -i video.mp4 --preset balanced`
**Dependencies:** None

- [ ] Task 1.1: Project scaffolding (go.mod, directory structure, Makefile)
- [ ] Task 1.2: Config system (config.go, paths.go, defaults)
- [ ] Task 1.3: Logging setup (file vs stderr based on mode)
- [ ] Task 1.4: FFmpeg/FFprobe detection and validation
- [ ] Task 1.5: FFprobe metadata extraction (probe.go, probe_types.go)
- [ ] Task 1.6: HDR detection (hdr.go)
- [ ] Task 1.7: Preset struct definition + 6 quality-tier presets
- [ ] Task 1.8: Preset registry with lookup (exact, case-insensitive, tag, fuzzy)
- [ ] Task 1.9: FFmpeg command builder (encode_args.go)
- [ ] Task 1.10: FFmpeg process runner + progress parsing (encode.go, progress.go, progress_types.go)
- [ ] Task 1.11: Graceful cancellation (cancel.go)
- [ ] Task 1.12: Output path resolution (suffix mode, conflict handling)
- [ ] Task 1.13: Temp file strategy with defer cleanup
- [ ] Task 1.14: CLI root command + run command + global flags (Fang)
- [ ] Task 1.15: Headless mode with stderr progress bar
- [ ] Task 1.16: `shrinkray version` subcommand
- [ ] Task 1.17: Unit tests for engine/ and presets/ packages

### Phase 2: TUI Core
**Goal:** Full single-file TUI workflow (splash → pick → info → presets → preview → encode → complete)
**Dependencies:** Phase 1

- [ ] Task 2.1: Theme struct + Neon Dusk color definitions (theme.go, styles.go)
- [ ] Task 2.2: Global key bindings (keys.go)
- [ ] Task 2.3: Inter-screen message types (messages/messages.go)
- [ ] Task 2.4: App model with screen routing + header/footer (app.go)
- [ ] Task 2.5: Screen 1 — Splash (ASCII logo, system capabilities, auto-advance)
- [ ] Task 2.6: Screen 2 — File picker (single-file selection, path input, Tab toggle)
- [ ] Task 2.7: Screen 3 — Source video info card (metadata display, quality heuristic, mini stat boxes)
- [ ] Task 2.8: Screen 4 — Preset selection grid (2-column cards, arrow nav, estimated sizes)
- [ ] Task 2.9: Screen 6 — Preview/confirmation (before/after cards, savings estimate, output path)
- [ ] Task 2.10: Screen 7 — Encoding progress (gradient bar, 6 stat boxes, log panel, cancel)
- [ ] Task 2.11: Screen 8 — Completion (bar chart, savings, open folder, re-encode)
- [ ] Task 2.12: Responsive layout (80/60 col breakpoints)
- [ ] Task 2.13: TUI launch integration in cli/run.go

### Phase 3: Full Preset Catalog
**Goal:** All 18 presets, size targeting, and smart recommendations
**Dependencies:** Phase 2

- [ ] Task 3.1: 5 purpose-driven presets (web, email, archive, slideshow, 4k-to-1080)
- [ ] Task 3.2: 7 platform-specific presets (discord, discord-nitro, whatsapp, twitter, instagram, tiktok, youtube)
- [ ] Task 3.3: Size estimation engine (estimate.go — CRF-based bitrate-per-pixel lookup)
- [ ] Task 3.4: Two-pass encoding support in command builder + process runner
- [ ] Task 3.5: Target-size bitrate calculation with safety factors (target_size.go)
- [ ] Task 3.6: Feasibility detection (bits-per-pixel thresholds, impossible/warning/acceptable)
- [ ] Task 3.7: Adaptive resolution scaling for tight size targets
- [ ] Task 3.8: Smart recommendation engine (recommend.go — rule-based analysis)
- [ ] Task 3.9: Recommendation display in info screen + preset grid (star badges, sorting)
- [ ] Task 3.10: Custom preset save/load from YAML (custom_presets.go)
- [ ] Task 3.11: `shrinkray presets` and `shrinkray presets show` subcommands
- [ ] Task 3.12: Two-pass progress display in TUI (Pass 1/2 label, combined percentage)
- [ ] Task 3.13: Unit tests for target_size, estimation, recommendation, and feasibility

### Phase 4: Advanced Features
**Goal:** Hardware acceleration, advanced options form, full codec support
**Dependencies:** Phase 3

- [ ] Task 4.1: Hardware encoder detection via test-encode probing (hwaccel.go, hwaccel_probe.go)
- [ ] Task 4.2: Encoder selection priority per platform (NVENC → VTB → QSV → AMF → VAAPI → software)
- [ ] Task 4.3: CRF/quality mapping across encoders (hwaccel quality table)
- [ ] Task 4.4: HW encoder integration in command builder
- [ ] Task 4.5: AV1 (SVT-AV1) codec support in command builder
- [ ] Task 4.6: VP9 codec support in command builder + WebM container handling
- [ ] Task 4.7: Portrait/landscape handling for social media presets (padding, aspect ratio filters)
- [ ] Task 4.8: Screen 5 — Advanced options Huh form (4 groups, dynamic visibility, codec-dependent ranges)
- [ ] Task 4.9: Splash screen update — display detected HW encoders and GPU name
- [ ] Task 4.10: Preview screen — show encoder name (SW vs HW) and efficiency note
- [ ] Task 4.11: `shrinkray probe` subcommand

### Phase 5: Batch Processing
**Goal:** Multi-file workflow with queue management, parallelism, and resume
**Dependencies:** Phase 4

- [ ] Task 5.1: Job struct and in-memory queue management
- [ ] Task 5.2: Multi-file selection in TUI file picker (Space toggle, selected count)
- [ ] Task 5.3: Queue ordering (size-asc default, configurable sort)
- [ ] Task 5.4: Screen 9a — Batch queue overview (file table, estimated totals)
- [ ] Task 5.5: Screen 9b — Batch progress display (overall + per-file progress, current file detail)
- [ ] Task 5.6: Screen 9c — Batch completion summary (table, total savings)
- [ ] Task 5.7: Parallel encoding (goroutine pool, --jobs N)
- [ ] Task 5.8: Queue persistence to JSON + resume on startup
- [ ] Task 5.9: Skip-if-already-compressed logic (output exists, source optimal, size threshold)
- [ ] Task 5.10: Failed job retry (max 2 attempts, --retry-failed flag)
- [ ] Task 5.11: Output directory mode with structure mirroring
- [ ] Task 5.12: In-place mode with verification (encode → probe → atomic rename)
- [ ] Task 5.13: Headless batch mode with progress lines
- [ ] Task 5.14: Batch-related CLI flags (--recursive, --jobs, --sort, --skip-existing, etc.)

### Phase 6: Polish
**Goal:** Complete UX, all remaining features
**Dependencies:** Phase 5

- [ ] Task 6.1: Electric Sunset theme + Ctrl+T runtime switching
- [ ] Task 6.2: Help overlay (? key, context-sensitive, semi-transparent modal)
- [ ] Task 6.3: Shell completions (bash, zsh, fish, powershell via Fang)
- [ ] Task 6.4: Dry-run mode (--dry-run)
- [ ] Task 6.5: Open folder on completion (platform-specific: explorer/open/xdg-open)
- [ ] Task 6.6: Stdin pipe input mode (--stdin, one path per line)
- [ ] Task 6.7: Metadata handling flags (--strip-metadata, -map_metadata)
- [ ] Task 6.8: Auto-rename conflict resolution (video_shrunk(1).mp4)
- [ ] Task 6.9: Adaptive terminal color support (LightDark detection, legacy Windows ANSI)
- [ ] Task 6.10: FFmpeg error parsing into user-friendly messages

### Phase 7: Distribution
**Goal:** Release-ready packaging and installation
**Dependencies:** Phase 6

- [ ] Task 7.1: GoReleaser configuration (.goreleaser.yaml)
- [ ] Task 7.2: GitHub Actions CI/CD (build, test, lint, release)
- [ ] Task 7.3: Homebrew tap formula (with ffmpeg dependency)
- [ ] Task 7.4: Scoop bucket manifest (with ffmpeg dependency)
- [ ] Task 7.5: Install script (bash for Unix, powershell for Windows)
- [ ] Task 7.6: Version injection via ldflags
- [ ] Task 7.7: README with usage examples

---

## 6. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Charm v2 API breaking changes or incomplete docs | Medium | Medium | Pin exact versions in go.mod; check pkg.go.dev before each phase; fall back to v1 patterns if needed |
| Fang insufficient for subcommand needs | Low | Low | Drop-in replace with Cobra directly — Fang wraps Cobra |
| FFmpeg progress parsing edge cases (no duration, variable framerate, concatenated streams) | Medium | Low | Fallback to frame-count-based percentage; handle missing fields gracefully |
| Hardware encoder test-encode fails silently on some platforms | Medium | Medium | Timeout test-encodes at 5s; log failures at debug level; always have software fallback |
| Two-pass encoding overshoots target size | Medium | Medium | Safety factor (0.97-0.99); verify output size post-encode; warn user if exceeded |
| Windows signal handling differs from Unix (no SIGSTOP, CTRL_BREAK issues) | High | Low | Use stdin `q` as primary cross-platform cancel; disable pause on Windows; document limitation |
| Terminal rendering inconsistencies (narrow terminals, no Unicode, light backgrounds) | Medium | Low | 3 layout breakpoints; LightDark detection; test on Windows Terminal + cmd.exe |

---

## 7. Success Criteria

- [ ] `shrinkray --no-tui -i video.mp4 --preset balanced` produces a valid compressed output
- [ ] All 18 built-in presets produce correct FFmpeg arguments
- [ ] TUI wizard flow completes end-to-end (splash → file pick → encode → completion)
- [ ] Hardware acceleration auto-detects and falls back to software gracefully
- [ ] Two-pass encoding hits target file size within 3% tolerance
- [ ] Batch mode processes multiple files with parallel encoding and resume
- [ ] Cross-platform: builds and runs on Windows, macOS, and Linux
- [ ] `go test ./...` passes with >=60% coverage on engine/ and presets/
- [ ] GoReleaser produces binaries for all 6 platform/arch combinations
- [ ] Single static binary with zero Go runtime dependencies

---

## 8. Assumptions

1. Charm v2 packages (Bubble Tea, Bubbles, Lip Gloss, Huh) have stable APIs matching IDEA.md's usage patterns — if wrong, API adjustments needed but not architectural changes
2. Fang provides Cobra-compatible subcommand/flag features — if wrong, use Cobra directly (trivial swap)
3. FFmpeg's `-progress pipe:1` output format is consistent across FFmpeg versions 4.x and 5.x — if wrong, add version-specific parsing
4. `go-ffprobe` v2 supports all metadata fields needed (HDR side data, channel layout) — if wrong, fall back to raw JSON parsing via encoding/json
5. GoReleaser supports current Go version and all target platforms — if wrong, use manual build scripts
6. User has FFmpeg installed or can install it — shrinkray does not bundle FFmpeg

---

## Planning Metrics
<!-- METRICS_JSON {"confidence": 92, "clarification_rounds": 1, "functional_requirements_count": 31, "non_functional_requirements_count": 8, "risk_count": 7, "phase_count": 7, "verification_gaps_found": 0, "confidence_breakdown": {"requirements": 24, "feasibility": 22, "integration": 23, "risk": 23}} -->
confidence: 92
clarification_rounds: 1
functional_requirements_count: 31
non_functional_requirements_count: 8
risk_count: 7
phase_count: 7
verification_gaps_found: 0
