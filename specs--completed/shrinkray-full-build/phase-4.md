# Phase 4: Advanced Features

- **Status:** Complete
- **Estimated Tasks:** 12
- **Goal:** Hardware acceleration, advanced options form, full codec support

---

## Overview

Add hardware-accelerated encoding (NVENC, VideoToolbox, QSV, AMF, VAAPI) with auto-detection via test-encode probing, extend codec support to AV1 and VP9, handle portrait/landscape video for social media presets, and build the advanced options form screen using Huh v2. This phase transforms shrinkray from a preset-only tool into one that supports fine-grained encoding control.

---

## Prerequisites

- Phase 3 complete (all 18 presets, two-pass encoding, size estimation)
- Understanding of FFmpeg hardware encoder flags per platform (NVENC: `-c:v h264_nvenc`, VideoToolbox: `-c:v h264_videotoolbox`, QSV: `-c:v h264_qsv`, etc.)
- IDEA.md section 10 (Hardware Acceleration) for encoder priority tables and quality mapping

---

## Tasks

### Hardware Detection

- [x] **Task 4.1:** Create `internal/engine/hwaccel.go` — define `HWEncoder` struct: `Name` (string, e.g., "nvenc"), `DisplayName` (string, e.g., "NVIDIA NVENC"), `GPU` (string, GPU name from test output), `Codecs` ([]string, supported codecs: h264/h265/av1), `Available` (bool). Define encoder priority per platform: Windows/Linux: NVENC → QSV → AMF → VAAPI → software; macOS: VideoToolbox → software. Export `DetectHardware(ctx context.Context) ([]HWEncoder, error)`. For each candidate encoder, attempt a test-encode (Task 4.2). Return sorted list with available encoders first.

- [x] **Task 4.2:** Create `internal/engine/hwaccel_probe.go` — export `ProbeEncoder(ctx context.Context, encoder string, codec string) (bool, error)`. Run a minimal test encode: generate 1-second color bars with `ffmpeg -f lavfi -i color=c=black:s=64x64:d=0.1 -c:v <encoder> -f null -` with 5-second timeout. If exit code 0, encoder is available. Catch common failure modes: "No NVENC capable devices", "Unknown encoder", GPU driver errors. Log results at debug level. Cache results in memory for the session (don't re-probe).

- [x] **Task 4.3:** Create quality mapping table in `internal/engine/hwaccel.go` — export `MapQuality(softwareCRF int, encoder string) map[string]string`. Map CRF values to hardware encoder equivalents: NVENC uses `-cq` (CRF 18→CQ 20, CRF 23→CQ 25, CRF 28→CQ 30); VideoToolbox uses `-q:v` (CRF 18→45, CRF 23→55, CRF 28→65); QSV uses `-global_quality` (CRF 18→22, CRF 23→27, CRF 28→32); AMF uses `-rc cqp -qp_i/-qp_p` (CRF 18→20, CRF 23→25, CRF 28→30). Return map of FFmpeg flag key-value pairs. Include codec-specific variations (h264 vs h265 vs av1 have different quality scales).

- [x] **Task 4.4:** Update `internal/engine/encode_args.go` `BuildArgs()` — when `opts.HWEncoder` is set (non-empty string): replace software codec with HW encoder name (e.g., `libx265` → `hevc_nvenc`), replace CRF flags with mapped quality flags from `MapQuality()`, add platform-specific device flags (`-hwaccel cuda` for NVENC, `-hwaccel videotoolbox` for VTB, `-hwaccel qsv` for QSV). Ensure `-movflags +faststart` is still applied for mp4. If HW encoder doesn't support two-pass, fall back to single-pass CQ mode and log warning.

### Additional Codecs

- [x] **Task 4.5:** Add AV1 support in `internal/engine/encode_args.go` — when `opts.Preset.Codec == "av1"`: use `-c:v libsvtav1` for software encoding, `-svtav1-params tune=0` for quality optimization. CRF range for SVT-AV1 is 0-63 (vs 0-51 for x264/x265) — adjust preset CRF values accordingly. Map presets: CRF 18→24, CRF 23→32, CRF 28→38. Container must be mp4 or webm (not mkv for broad compatibility). Add HW AV1 encoder support: `av1_nvenc`, `av1_qsv`, `av1_amf` when available.

- [x] **Task 4.6:** Add VP9 support in `internal/engine/encode_args.go` — when `opts.Preset.Codec == "vp9"`: use `-c:v libvpx-vp9`, CRF mode with `-b:v 0` (required for CRF-only mode in VP9). Container must be webm. Audio codec must be opus (not AAC). Adjust audio args: `-c:a libopus -b:a <bitrate>`. VP9 two-pass uses different passlog format — handle in two-pass orchestration. Add `-row-mt 1` for multi-threaded encoding.

### Social Media Handling

- [x] **Task 4.7:** Implement portrait/landscape handling in `internal/engine/encode_args.go` — add logic for presets with `Tags` containing "social" or platform category. When source video `IsPortrait()` and preset targets a specific resolution (e.g., 1080p), swap width/height in scale filter (e.g., `-vf scale=1080:1920` instead of `scale=1920:1080`). For Instagram/TikTok presets, add padding filter for non-9:16 portrait videos: `-vf "scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2:black"`. Detect portrait via `VideoInfo.IsPortrait()` (height > width).

### Advanced Options Screen

- [x] **Task 4.8:** Create `internal/tui/screens/advanced.go` — implement Screen 5 using Huh v2 form library. Four form groups: **Video** (codec select: h264/h265/av1/vp9, CRF slider with codec-dependent range, resolution dropdown: source/2160p/1440p/1080p/720p/480p/custom, max FPS input), **Audio** (codec select: aac/opus/copy, bitrate dropdown: 64k/96k/128k/192k/256k/320k, channels: source/stereo/mono), **Encoder** (hardware encoder dropdown populated from `DetectHardware()` results — only show available encoders, with "Software" default), **Output** (suffix text input, conflict mode select: ask/skip/overwrite/autorename). Form fields dynamically show/hide: VP9 forces opus audio, AV1 hides incompatible options, HW encoder selection updates quality slider range. Pre-populate form with selected preset values. On submit, create modified `EncodeOptions` and navigate to preview screen.

### TUI Updates

- [x] **Task 4.9:** Update `internal/tui/screens/splash.go` — after `DetectHardware()` completes (run as async `tea.Cmd` on init), display detected HW encoders with GPU name. Show green checkmark for available encoders, dim text for unavailable. Example: "✓ NVENC (NVIDIA RTX 3080) — h264, h265, av1" / "✗ VideoToolbox — not available". Keep auto-advance timer but reset if HW detection takes longer than 2s.

- [x] **Task 4.10:** Update `internal/tui/screens/preview.go` — show encoder name in the "After" card (e.g., "Encoder: NVENC (h265_nvenc)" or "Encoder: Software (libx265)"). Add efficiency note: if HW encoding, show "~2-3x faster, slightly larger file" hint. If two-pass, show "Two-pass encoding — slower but precise file size" hint. Show "Advanced options" link — pressing `a` navigates to advanced screen (back returns to preview with updated options).

- [x] **Task 4.11:** Create `internal/cli/probe.go` — implement `shrinkray probe <file>` subcommand. Run `Probe()` on the given file and display all `VideoInfo` fields in a formatted output: file name, format, duration, file size, video codec + resolution + framerate + bitrate, audio codec + channels + sample rate + bitrate, HDR status, stream count. Use Lip Gloss styling for terminal output (or plain text if `--no-color`). Exit with error if file doesn't exist or isn't a valid video.

### Phase Testing

- [x] **Task 4.12:** Write unit tests for Phase 4 packages. **engine/**: test `MapQuality()` returns correct mappings for all encoder/CRF combinations; test `BuildArgs()` with HW encoder set produces correct FFmpeg flags (e.g., `hevc_nvenc` instead of `libx265`); test `BuildArgs()` with AV1 codec produces `libsvtav1` args; test `BuildArgs()` with VP9 produces `libvpx-vp9` args with `-b:v 0` and opus audio; test portrait handling produces swapped scale dimensions; test `ProbeEncoder()` handles timeout correctly (mock). **Note:** actual HW encoder availability varies by machine — tests should verify arg building logic, not real GPU availability. Target: all new tests pass.

---

## Acceptance Criteria

- `shrinkray --no-tui -i video.mp4 --preset balanced --hw nvenc` uses NVENC hardware encoding
- Hardware encoder auto-detection runs on startup and reports available encoders
- Software fallback works when no HW encoders available
- `shrinkray --no-tui -i video.mp4 --codec av1 --crf 32` produces AV1-encoded output
- `shrinkray --no-tui -i video.mp4 --codec vp9` produces WebM with VP9+Opus
- Portrait video compressed with TikTok preset produces 9:16 output with correct padding
- TUI advanced options form allows changing codec, CRF, resolution, audio, and encoder
- Advanced form dynamically shows/hides options based on codec selection
- `shrinkray probe video.mp4` displays formatted video metadata
- Splash screen shows detected hardware encoders
- `go test ./internal/engine/...` passes with new HW/codec tests

---

## Notes

- HW encoder detection results are cached per session — no re-probing after splash screen
- NVENC two-pass is not well-supported — fall back to single-pass CQ mode for NVENC with target-size presets
- AV1 software encoding (SVT-AV1) is significantly slower than h264/h265 — consider adding speed warning in preview
- VP9 is primarily for WebM output — most users will use h264/h265 for mp4
- The advanced options form should feel optional — most users will use presets directly

---

## Phase Completion Summary

| Field | Value |
|-------|-------|
| Date completed | 2026-03-21 |
| Implementer | Claude Opus 4.6 |
| What was done | All 12 tasks: HW encoder detection (NVENC/VTB/QSV/AMF/VAAPI), test-encode probing with 5s timeout and session caching, quality mapping tables for all HW encoders, BuildArgs HW encoder integration, AV1 support (libsvtav1 + CRF mapping + HW encoders), VP9 support (libvpx-vp9 + opus + row-mt), portrait/social handling with padding for Instagram/TikTok, advanced options TUI screen, splash screen HW encoder display, preview screen encoder info + efficiency hints, `shrinkray probe` CLI subcommand, 22 unit tests |
| Files changed | New: internal/engine/hwaccel.go, internal/engine/hwaccel_probe.go, internal/engine/phase4_test.go, internal/tui/screens/advanced.go, internal/tui/style/theme.go, internal/tui/style/styles.go, internal/tui/style/screen.go, internal/cli/probe.go. Modified: internal/engine/encode_args.go, internal/tui/screens/splash.go, internal/tui/screens/preview.go, internal/tui/app.go, internal/tui/messages/messages.go, internal/cli/root.go, internal/tui/styles.go, internal/tui/theme.go, internal/tui/screen.go, internal/tui/theme_test.go, internal/tui/screens/presets.go, internal/tui/screens/complete.go, internal/tui/screens/info.go, internal/tui/screens/encoding.go, internal/tui/screens/filepicker.go |
| Issues encountered | Pre-existing import cycle between tui and tui/screens packages resolved by extracting styles/theme/ScreenModel into new tui/style leaf package. Pre-existing bubbles v2 API incompatibilities fixed (progress.WithDefaultGradient -> WithDefaultBlend, filepicker.Model.Update return type change). lipgloss v2 changed Color from type to function (returns color.Color). |
