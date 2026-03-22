# Phase 3: Full Preset Catalog

- **Status:** Complete
- **Estimated Tasks:** 14
- **Goal:** All 18 presets, size targeting, and smart recommendations

---

## Overview

Complete the preset catalog with all 18 built-in presets (adding 5 purpose-driven and 7 platform-specific), implement the size estimation engine, two-pass encoding for file-size targeting, and the smart recommendation engine that analyzes source video metadata to suggest optimal presets. Integrate recommendations and real size estimates into the TUI screens.

---

## Prerequisites

- Phase 2 complete (TUI core with all 7 screens functional)
- Understanding of FFmpeg two-pass encoding workflow (`-pass 1 -f null /dev/null` then `-pass 2`)
- IDEA.md section 6 (Preset System) for exact preset parameters and section 8 for recommendation rules

---

## Tasks

### Presets

- [x] **Task 3.1:** Create `internal/presets/purpose.go` — define and register 5 purpose-driven presets: `web` (h264, CRF 23, 1080p max, AAC 128k — optimized for web delivery), `email` (h264, CRF 28, 720p max, AAC 96k, target ~20MB), `archive` (h265, CRF 18, keep resolution, copy audio — maximum quality retention), `slideshow` (h264, CRF 32, 720p max, AAC 64k — photo slideshows/screencasts with low motion), `4k-to-1080` (h265, CRF 20, force 1080p, AAC 192k — downscale 4K to 1080p). Each preset fully populated with tags, icon, description.

- [x] **Task 3.2:** Create `internal/presets/platform.go` — define and register 7 platform-specific presets: `discord` (h264, target 10MB, 720p max, AAC 128k, two-pass), `discord-nitro` (h264, target 50MB, 1080p max, AAC 192k, two-pass), `whatsapp` (h264, target 16MB, 720p max, 30fps max, AAC 128k, two-pass), `twitter` (h264, CRF 23, 1080p max, 60fps max, AAC 192k, 140s max warn), `instagram` (h264, CRF 23, 1080p max, 30fps max, AAC 128k, 60s max warn), `tiktok` (h264, CRF 23, 1080p max, 60fps max, AAC 128k, portrait-aware), `youtube` (h265, CRF 18, keep resolution, AAC 192k — upload optimized). Each includes platform-specific constraints in description and tags.

### Size Estimation

- [x] **Task 3.3:** Create `internal/engine/estimate.go` — export `EstimateSize(info *engine.VideoInfo, preset presets.Preset) int64`. For CRF-based presets: use bitrate-per-pixel lookup table keyed by codec and CRF value (e.g., h265 CRF 23 ≈ 0.08 bits/pixel/frame). Calculate: `estimatedBitrate = bpp * width * height * fps`, then `estimatedBytes = (estimatedBitrate * duration) / 8 + audioBytes`. For target-size presets: return `TargetSizeMB * 1024 * 1024`. Account for resolution scaling (use target resolution in calculation if preset specifies one). Return value in bytes.

- [x] **Task 3.4:** Create `internal/engine/target_size.go` — export `CalculateBitrate(targetBytes int64, duration time.Duration, audioBitrate int64) int64`. Calculate: `videoBitrate = ((targetBytes * 8) / durationSeconds) - audioBitrate`. Apply safety factor of 0.97 to account for container overhead. Export `FeasibilityCheck(targetBytes int64, duration time.Duration, width int, height int) FeasibilityResult`. Define `FeasibilityResult` with `Status` (impossible/warning/acceptable), `BitsPerPixel` (float64), `Message` (string). Thresholds: below 0.01 bpp = impossible, 0.01-0.04 = warning (quality loss likely), above 0.04 = acceptable.

- [x] **Task 3.5:** Implement adaptive resolution scaling in `internal/engine/target_size.go` — export `AdaptiveResolution(targetBytes int64, duration time.Duration, width int, height int, fps float64) (int, int)`. When feasibility check returns impossible or warning at source resolution, iteratively try lower resolutions (source → 1080p → 720p → 480p → 360p) until bpp reaches acceptable range. Return the recommended resolution. Preserve aspect ratio. Integrate into `BuildArgs()` when preset has `TargetSizeMB > 0` and source resolution would result in poor quality.

### Two-Pass Encoding

- [x] **Task 3.6:** Add two-pass encoding support to `internal/engine/encode_args.go` — when `opts.Pass == 1`: add `-pass 1 -passlogfile <logfile> -f null` and output to `/dev/null` (or `NUL` on Windows). When `opts.Pass == 2`: add `-pass 2 -passlogfile <logfile>` with actual output path. The passlogfile path should use `TempPath()` with `.passlog` extension in the same directory as output.

- [x] **Task 3.7:** Add two-pass orchestration to `internal/engine/encode.go` — export `EncodeTwoPass(ctx context.Context, opts EncodeOptions) (<-chan ProgressUpdate, error)`. Run pass 1 (progress 0-50%), then pass 2 (progress 50-100%). Combine progress updates so the consumer sees a single 0-100% progression. Clean up passlog file after pass 2 completes. If pass 1 fails, don't start pass 2. Use `opts.TwoPass` or `opts.Preset.TargetSizeMB > 0` to trigger two-pass mode. Verify output size after completion — if exceeds target by >3%, log warning.

### Recommendations

- [x] **Task 3.8:** Create `internal/presets/recommend.go` — define `Recommendation` struct: `Preset` (Preset), `Score` (int, 0-100), `Reason` (string), `Tags` ([]string, e.g., "best-quality", "smallest", "fastest"). Export `Recommend(info *engine.VideoInfo) []Recommendation`. Rule-based analysis: if source bitrate > 10Mbps → suggest `balanced` and `compact` (high compression potential); if 4K resolution → suggest `4k-to-1080`; if source is already h265 with low bitrate → warn "already well-compressed"; if short duration (<120s) → suggest platform presets (discord, whatsapp); if large file (>500MB) → suggest `compact` or `email`. Sort by score descending. Return top 5 recommendations.

- [x] **Task 3.9:** Create `internal/presets/custom_presets.go` — export `LoadCustomPresets(configDir string) ([]Preset, error)` and `SaveCustomPreset(configDir string, preset Preset) error`. Custom presets stored in `<configDir>/custom_presets.yaml` as YAML array. Each custom preset has `custom: true` field. Loaded custom presets are registered in the global registry via `Register()`. Validate custom preset has required fields (key, name, codec, crf or targetSizeMB). Key must not conflict with built-in preset keys.

### TUI Integration

- [x] **Task 3.10:** Update `internal/tui/screens/info.go` — after video probe completes, run `Recommend()` and display recommendation badges on the info card. Show top 3 recommendations with star icon, preset name, and reason. Add "Recommended for you" section below the video metadata. Each recommendation is clickable (Enter on highlighted recommendation navigates to preset grid with that preset pre-selected).

- [x] **Task 3.11:** Update `internal/tui/screens/presets.go` — show all 18 presets across 3 category tabs (Quality, Purpose, Platform). Add estimated output size to each preset card (using `EstimateSize()`). Mark recommended presets with a star badge and sort them first within each category. Show feasibility warning badge on target-size presets where `FeasibilityCheck()` returns warning/impossible. Update preset detail panel to show full preset info including estimated compression ratio.

- [x] **Task 3.12:** Update `internal/tui/screens/encoding.go` — handle two-pass encoding display. Show "Pass 1/2" or "Pass 2/2" label above the progress bar. Progress bar shows combined 0-100% across both passes. Add pass indicator to stat boxes. Update `internal/cli/presets.go` — implement `shrinkray presets` subcommand listing all 18 presets in a formatted table (key, name, category, codec, CRF/target) and `shrinkray presets show <key>` showing full preset details.

### Phase Testing

- [x] **Task 3.13:** Write unit tests for Phase 3 packages. **presets/**: test `All()` returns 18 presets; test all purpose and platform presets have valid fields; test `ByCategory()` returns correct counts (6 quality, 5 purpose, 7 platform). **engine/**: test `EstimateSize()` returns reasonable estimates for known video dimensions (e.g., 1080p 60s video with balanced preset should estimate between 5-50MB); test `CalculateBitrate()` math with known values; test `FeasibilityCheck()` returns correct status at boundary thresholds; test `AdaptiveResolution()` scales down for tight targets. **presets/**: test `Recommend()` returns recommendations for various video profiles (4K source, short clip, already-compressed). **presets/**: test `LoadCustomPresets()` and `SaveCustomPreset()` round-trip. Target: all new tests pass.

---

## Acceptance Criteria

- `shrinkray presets` lists all 18 presets in a formatted table
- `shrinkray presets show discord` shows full discord preset details including 10MB target
- `shrinkray --no-tui -i video.mp4 --preset discord` performs two-pass encoding targeting 10MB
- Two-pass encoding output is within 3% of target file size
- TUI preset grid shows all 18 presets across 3 category tabs
- TUI preset cards show estimated output sizes
- TUI info screen shows smart recommendations based on source video
- Feasibility warning appears for impossible/marginal target-size presets
- Custom presets can be saved and loaded from YAML config
- `go test ./internal/presets/... ./internal/engine/...` passes with new tests

---

## Notes

- Platform presets with duration limits (twitter 140s, instagram 60s) show warnings but don't enforce — user decides
- Custom preset YAML format should be documented in a comment at top of the file when first created
- The recommendation engine uses simple rule-based heuristics, not ML — sufficient for v1
- Two-pass encoding is approximately 2x slower than single-pass — warn user in TUI preview screen

---

## Phase Completion Summary

| Field | Value |
|-------|-------|
| Date completed | 2026-03-21 |
| Implementer | Claude Code |
| What was done | All 13 tasks complete: 5 purpose + 7 platform presets, size estimation engine with BPP lookup, target-size bitrate calculation with feasibility checks and adaptive resolution, two-pass encoding orchestration, rule-based recommendation engine, custom preset YAML persistence, TUI info screen with recommendations, TUI preset screen with 3 category tabs and estimated sizes, TUI encoding screen with two-pass display, CLI presets subcommand, comprehensive unit tests |
| Files changed | internal/presets/purpose.go (new), internal/presets/platform.go (new), internal/presets/registry.go (modified), internal/presets/recommend.go (new), internal/presets/custom_presets.go (new), internal/engine/estimate.go (new), internal/engine/target_size.go (new), internal/engine/encode_args.go (modified - added VideoBitrate field and bitrate-mode encoding), internal/engine/encode.go (modified - added EncodeTwoPass, ShouldUseTwoPass, cleanup helpers), internal/tui/screens/info.go (modified - recommendations section), internal/tui/screens/presets.go (modified - 3 category tabs, estimated sizes, feasibility warnings), internal/tui/screens/encoding.go (modified - two-pass display), internal/tui/screens/preview.go (modified - use engine.EstimateSize), internal/cli/presets.go (new - presets subcommand), internal/cli/root.go (modified - register presets subcommand), internal/cli/run.go (modified - two-pass headless support), internal/presets/presets_test.go (modified), internal/presets/phase3_test.go (new), internal/engine/phase3_test.go (new) |
| Issues encountered | Circular dependency between engine and presets packages resolved by introducing presets.VideoMetadata struct instead of importing engine.VideoInfo in the recommendation engine |
