# Phase 2: TUI Core

- **Status:** In Progress
- **Estimated Tasks:** 14
- **Goal:** Full single-file TUI workflow (splash → pick → info → presets → preview → encode → complete)

---

## Overview

Build the Bubble Tea v2 TUI application with 8 screens for the single-file compression workflow. Establish the theme system, screen routing architecture, inter-screen messaging, and responsive layout system. After this phase, a user can launch shrinkray (no flags) and walk through the full wizard: splash → file picker → video info → preset selection → preview/confirm → encoding progress → completion.

---

## Prerequisites

- Phase 1 complete (engine, presets, config, CLI packages)
- Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh v2 added to go.mod
- Familiarity with Bubble Tea Model-View-Update pattern and `tea.Msg` types

---

## Tasks

### Theme & Styling

- [x] **Task 2.1:** Create `internal/tui/theme.go` — define `Theme` struct with color fields: `Primary`, `Secondary`, `Accent`, `Success`, `Warning`, `Error`, `Muted`, `Background`, `Surface`, `Text`, `TextDim`, `Border`, `Highlight`, `GradientStart`, `GradientEnd` (all `lipgloss.Color`). Implement `NeonDusk() Theme` with the project's default palette (deep purples, neon cyan/magenta accents — refer to IDEA.md section 12 for exact hex values). Export `var ActiveTheme Theme` initialized to Neon Dusk. Theme switching logic (Ctrl+T) will be added in Phase 6.

- [x] **Task 2.2:** Create `internal/tui/styles.go` — define reusable Lip Gloss styles using `ActiveTheme` colors: `HeaderStyle`, `FooterStyle`, `TitleStyle`, `SubtitleStyle`, `CardStyle` (bordered box), `CardActiveStyle` (highlighted border), `ButtonStyle`, `ButtonActiveStyle`, `StatBoxStyle` (small bordered box for stats), `ProgressBarStyle`, `ErrorStyle`, `MutedStyle`, `KeyHintStyle`. All styles must reference theme colors (not hardcoded) so theme switching works later. Include `TerminalWidth() int` and `TerminalHeight() int` helpers.

### Framework

- [x] **Task 2.3:** Create `internal/tui/keys.go` — define `KeyMap` struct using `bubbles/key` with bindings: `Quit` (q, ctrl+c), `Back` (esc), `Enter` (enter), `Help` (?), `Tab` (tab), `Up/Down/Left/Right` (arrows), `ThemeToggle` (ctrl+t). Export `DefaultKeyMap() KeyMap`. Each binding includes short help text for the help overlay (Phase 6).

- [x] **Task 2.4:** Create `internal/tui/messages/messages.go` — define all inter-screen `tea.Msg` types: `FileSelectedMsg` (path string), `VideoProbeCompleteMsg` (info *engine.VideoInfo, err error), `PresetSelectedMsg` (preset presets.Preset), `EncodeStartMsg` (opts engine.EncodeOptions), `EncodeProgressMsg` (update engine.ProgressUpdate), `EncodeCompleteMsg` (outputPath string, inputSize int64, outputSize int64), `EncodeErrorMsg` (err error), `EncodeCancelMsg`, `NavigateMsg` (screen Screen), `BackMsg`, `ThemeToggleMsg`, `WindowSizeMsg` (width, height int). Define `Screen` type with constants for each screen.

- [x] **Task 2.5:** Create `internal/tui/app.go` — define `App` struct implementing `tea.Model` with fields: `currentScreen Screen`, `screens map[Screen]tea.Model`, `width int`, `height int`, `keyMap KeyMap`, `videoInfo *engine.VideoInfo`, `selectedPreset *presets.Preset`, `encodeOpts *engine.EncodeOptions`. Implement `Init()` returning splash screen init. Implement `Update()` handling: `tea.WindowSizeMsg` (propagate to all screens), `NavigateMsg` (switch screen, call new screen's `Init()`), `BackMsg` (navigate to previous screen), global `KeyMap` bindings (quit, theme toggle), and delegate remaining messages to current screen. Implement `View()` rendering: header bar (app name + current screen title), current screen view, footer bar (key hints for current screen). Header/footer use `HeaderStyle`/`FooterStyle`.

### Screens

- [x] **Task 2.6:** Create `internal/tui/screens/splash.go` — implement splash screen model. Display ASCII art logo for "shrinkray" (styled with gradient colors), tagline "Less bytes, same vibes.", system info: FFmpeg version, FFprobe version, detected HW encoders (placeholder "detecting..." — actual HW detection in Phase 4). Auto-advance to file picker after 2 seconds or on any keypress. Use `tea.Tick` for auto-advance timer.

- [x] **Task 2.7:** Create `internal/tui/screens/filepicker.go` — implement file picker screen using `bubbles/filepicker` component. Configure to show video file extensions (.mp4, .mkv, .webm, .avi, .mov, .wmv, .flv, .ts, .m4v). Include text input field for direct path entry (Tab to toggle between filepicker and text input). On file selection, send `FileSelectedMsg` and trigger async `Probe()` via `tea.Cmd`. Show file size and name in status bar. Handle invalid file selection with error message.

- [x] **Task 2.8:** Create `internal/tui/screens/info.go` — implement video info screen. Display probed `VideoInfo` in a styled card layout: video thumbnail area (placeholder — no actual thumbnail), codec badge, resolution, framerate, bitrate, duration, file size, audio info, HDR badge (if applicable). Show quality heuristic assessment (bitrate-per-pixel calculation → "overcompressed" / "well-compressed" / "could shrink significantly"). Display mini stat boxes for key metrics. Navigation: Enter to continue to presets, Esc to go back to file picker.

- [x] **Task 2.9:** Create `internal/tui/screens/presets.go` — implement preset selection screen. Display presets in a 2-column grid of styled cards. Each card shows: icon, name, description, estimated output size (placeholder calculation — real estimation in Phase 3), category badge. Arrow keys navigate between cards, highlighting active card with `CardActiveStyle`. Category filter tabs at top (All / Quality / Purpose / Platform) — only show Quality category in this phase (6 presets). Enter to select preset and navigate to preview. Show selected preset detail panel on the right when terminal is wide enough (80+ cols).

- [x] **Task 2.10:** Create `internal/tui/screens/preview.go` — implement preview/confirmation screen. Display side-by-side "Before" and "After" cards. Before card: source file info (name, size, codec, resolution, bitrate). After card: selected preset name, estimated output codec, estimated resolution, estimated size (placeholder). Show estimated savings percentage and size reduction. Display output file path. Navigation: Enter to start encoding (send `EncodeStartMsg`), Esc to go back to presets, `e` to edit advanced options (placeholder — Phase 4).

- [x] **Task 2.11:** Create `internal/tui/screens/encoding.go` — implement encoding progress screen. Display animated progress bar (gradient fill using `bubbles/progress`). Show 6 stat boxes: Percent, Speed, FPS, Bitrate, Size, ETA. Below stats, show scrolling log panel (last 5 progress updates as text lines). Cancel button/hint (Esc or `c` to cancel — triggers `Cancel()` then sends `EncodeCancelMsg`). On receiving `EncodeProgressMsg`, update all displays. On `EncodeCompleteMsg`, navigate to completion screen. Subscribe to progress channel via `tea.Cmd` that reads from channel and returns `EncodeProgressMsg`.

- [x] **Task 2.12:** Create `internal/tui/screens/complete.go` — implement completion screen. Display bar chart comparing input vs output file size (horizontal bars with labels). Show savings: absolute size reduction, percentage, and compression ratio. Show output file path. Action buttons/hints: `o` to open containing folder (placeholder — Phase 6), `r` to re-encode with different preset (navigate back to presets), `n` for new file (navigate to file picker), `q` to quit. Display encoding time elapsed.

### Layout & Integration

- [x] **Task 2.13:** Implement responsive layout in `internal/tui/app.go` and all screen `View()` methods. Define 3 breakpoints: wide (80+ cols, full layout with side panels), medium (60-79 cols, single column, condensed cards), narrow (<60 cols, minimal layout, abbreviated labels). Pass terminal dimensions to all screens via `tea.WindowSizeMsg`. Each screen's `View()` must adapt: preset grid switches from 2-col to 1-col below 80, stat boxes wrap from horizontal to vertical below 60, cards reduce padding in narrow mode.

- [x] **Task 2.14:** Integrate TUI launch in `internal/cli/run.go` — when `--no-tui` is NOT set and input is a terminal (check `term.IsTerminal()`), create `App` model with probed video info (if `--input` provided) or empty (user picks file in TUI). Start Bubble Tea program with `tea.WithAltScreen()`, `tea.WithMouseCellMotion()`. Pass config, detected FFmpeg info, and initial file path (if any) to App constructor. Handle Bubble Tea program exit and return appropriate exit code.

### Phase Testing

- [x] **Task 2.15:** Write tests for Phase 2 non-TUI components. **messages/**: test all message types can be instantiated with expected fields. **theme.go**: test `NeonDusk()` returns theme with all color fields non-empty. **keys.go**: test `DefaultKeyMap()` returns keymap with all bindings defined. **Note:** Bubble Tea screen models are not directly unit tested (per testing strategy — low ROI). Manual testing checklist: launch TUI, navigate all 7 screens, verify key bindings work, verify responsive layout at different terminal widths (resize terminal to test breakpoints).

---

## Acceptance Criteria

- `shrinkray` (no flags) launches TUI in alternate screen
- Splash screen displays and auto-advances after 2 seconds
- File picker allows navigating filesystem and selecting a video file
- Info screen displays probed video metadata correctly
- Preset selection grid shows 6 quality-tier presets in styled cards
- Preview screen shows before/after comparison
- Encoding screen shows progress bar and stats updating in real-time
- Completion screen shows file size comparison and savings
- Esc navigates back through screens
- Ctrl+C exits cleanly from any screen
- Layout adapts when terminal is resized
- `shrinkray -i video.mp4` launches TUI with file pre-selected (skips file picker)

---

## Notes

- Only 6 quality-tier presets shown in this phase — remaining 12 presets added in Phase 3
- Size estimation is placeholder (simple percentage guess) — real estimation engine in Phase 3
- Advanced options screen (Screen 5) is skipped in this phase — added in Phase 4
- Theme switching (Ctrl+T) is registered but only Neon Dusk available — Electric Sunset in Phase 6
- Help overlay (?) is a no-op in this phase — implemented in Phase 6
- Batch screens (9a, 9b, 9c) are not built in this phase — added in Phase 5

---

## Phase Completion Summary

| Field | Value |
|-------|-------|
| Date completed | -- |
| Implementer | -- |
| What was done | -- |
| Files changed | -- |
| Issues encountered | -- |
