# Phase 5: Batch Processing

- **Status:** Complete
- **Estimated Tasks:** 15
- **Goal:** Multi-file workflow with queue management, parallelism, and resume

---

## Overview

Add multi-file batch processing with a job queue, parallel encoding via goroutine pool, queue persistence for resume across sessions, skip-if-already-compressed logic, and 3 new TUI batch screens (queue overview, batch progress, batch completion). Extend the file picker for multi-select and add all batch-related CLI flags for headless mode.

---

## Prerequisites

- Phase 4 complete (hardware acceleration, advanced options, full codec support)
- Understanding of Go concurrency patterns (goroutine pools, sync.WaitGroup, channels)
- IDEA.md section 14 (Batch Processing) for queue management and persistence design

---

## Tasks

### Job Queue

- [x] **Task 5.1:** Create `internal/engine/job.go` — define `Job` struct: `ID` (string, UUID), `InputPath` (string), `OutputPath` (string), `PresetKey` (string), `Status` (enum: pending/encoding/complete/failed/skipped), `Progress` (float64, 0-100), `InputSize` (int64), `OutputSize` (int64), `Error` (string), `Attempts` (int), `StartedAt` (time.Time), `CompletedAt` (time.Time), `Pass` (int, for two-pass tracking). Define `JobQueue` struct with `Jobs []Job`, `mutex sync.RWMutex` for thread-safe access. Export methods: `Add(job Job)`, `Next() (*Job, bool)`, `Update(id string, fn func(*Job))`, `ByStatus(status Status) []Job`, `Stats() QueueStats` (total, pending, encoding, complete, failed, skipped counts + total input/output sizes).

- [x] **Task 5.2:** Implement queue ordering in `internal/engine/job.go` — export `Sort(jobs []Job, mode string)`. Sort modes: `size-asc` (default — smallest first for quick wins), `size-desc` (largest first), `name` (alphabetical), `duration` (shortest first). Configurable via config file `batch.sort` and `--sort` CLI flag.

- [x] **Task 5.3:** Implement skip-if-compressed logic in `internal/engine/job.go` — export `ShouldSkip(job Job, opts SkipOptions) (bool, string)`. Define `SkipOptions`: `SkipExisting` (bool — output file already exists), `SkipOptimal` (bool — source already uses target codec and bitrate is below threshold), `SizeThreshold` (float64 — skip if source is already below this MB). Return bool + reason string. Check output exists, probe source codec/bitrate vs preset target, check file size against threshold. Used before each job starts.

### TUI Multi-Select

- [x] **Task 5.4:** Update `internal/tui/screens/filepicker.go` — add multi-file selection mode. Space bar toggles file selection (visual checkmark). Show selected count in status bar (e.g., "3 files selected"). Enter with multiple files selected navigates to batch queue screen (Screen 9a) instead of info screen. Tab still toggles to text input — in text input mode, support comma-separated paths or directory path (auto-discover video files in directory). Add `--recursive` behavior: when a directory is entered, walk it with `filepath.WalkDir()` finding all video files by extension.

### Batch TUI Screens

- [x] **Task 5.5:** Create `internal/tui/screens/batch_queue.go` — implement Screen 9a (batch queue overview). Display table of queued files: filename, size, preset, estimated output size, status. Show totals row: total input size, total estimated output size, estimated savings. Allow reordering (move up/down with Shift+arrows), removing files (Delete/d key), changing preset per-file (p key opens preset picker modal). Navigation: Enter to start batch encoding, Esc to go back to file picker. Show estimated total encoding time based on source durations and speed estimate.

- [x] **Task 5.6:** Create `internal/tui/screens/batch_progress.go` — implement Screen 9b (batch progress). Display overall progress bar (based on total bytes processed / total bytes). Below, show per-file rows: filename, individual progress bar (mini), status badge (pending/encoding/complete/failed/skipped), output size. Highlight currently encoding file(s). Show current file detail panel: same stat boxes as single-file encoding screen (percent, speed, FPS, bitrate, size, ETA). Show overall ETA and elapsed time. Cancel button (Esc) — prompts "Cancel current file only or entire batch?" with `c` for current, `a` for all.

- [x] **Task 5.7:** Create `internal/tui/screens/batch_complete.go` — implement Screen 9c (batch completion summary). Display table: filename, input size, output size, savings %, status (with color: green=complete, yellow=skipped, red=failed). Show totals: total input, total output, total savings (absolute and %), total encoding time. Show failed files with error messages and retry option (r key). Navigation: `o` to open output folder, `n` for new batch, `q` to quit.

### Parallel Encoding

- [x] **Task 5.8:** Create `internal/engine/batch.go` — export `RunBatch(ctx context.Context, queue *JobQueue, opts BatchOptions) <-chan BatchEvent`. Define `BatchOptions`: `Jobs` (int, parallel workers, default 1), `Preset` (presets.Preset), `HWEncoder` (string), `OutputOpts` (OutputOptions), `SkipOpts` (SkipOptions). Define `BatchEvent` interface with types: `JobStartedEvent`, `JobProgressEvent`, `JobCompleteEvent`, `JobFailedEvent`, `JobSkippedEvent`, `BatchCompleteEvent`. Use goroutine worker pool (`--jobs N`): launch N workers reading from job channel. Each worker: check `ShouldSkip()`, resolve output path, run `Encode()` or `EncodeTwoPass()`, update job status via `queue.Update()`, send events on channel. Respect context cancellation. Default `--jobs 1` for safety; warn if `--jobs > 1` that file order is not guaranteed.

### Persistence & Resume

- [x] **Task 5.9:** Create `internal/engine/batch_persist.go` — export `SaveQueue(path string, queue *JobQueue) error` and `LoadQueue(path string) (*JobQueue, error)`. Serialize queue to JSON at `CacheDir()/batch_queue.json`. Save after each job status change (debounced — max once per second). On startup, check for existing queue file: if found with pending/failed jobs, offer to resume. Export `CleanQueue(path string) error` to delete the queue file after batch completes successfully.

- [x] **Task 5.10:** Implement retry logic in `internal/engine/batch.go` — when a job fails and `Attempts < MaxAttempts` (default 2), re-queue with incremented attempt count. Retry on next available worker. If retried job fails again, mark as permanently failed. Add `--retry-failed` CLI flag that re-queues only failed jobs from a persisted queue. Add `--max-retries` flag (default 2).

### Output Modes

- [x] **Task 5.11:** Extend `internal/engine/output.go` — implement directory output mode: `ResolveOutput()` with `Mode: "directory"` mirrors the input directory structure under `opts.Directory`. Example: input `/home/user/Videos/vacation/clip.mp4` with output dir `/home/user/compressed/` → output `/home/user/compressed/vacation/clip.mp4`. Create intermediate directories with `os.MkdirAll()`. Handle conflict modes for existing files.

- [x] **Task 5.12:** Implement in-place mode in `internal/engine/output.go` — when `Mode: "inplace"`: encode to temp file, probe temp file with `Probe()` to verify it's a valid video (has video stream, duration within 5% of source), then `os.Rename(temp, source)` to atomically replace. If verification fails, keep source unchanged and return error. Log verification results. This mode is inherently destructive — require explicit `--in-place` flag, warn in TUI with confirmation prompt.

### Headless Batch

- [x] **Task 5.13:** Implement headless batch mode in `internal/cli/run.go` — when `--no-tui` with multiple `--input`/`-i` flags or a directory input: create `JobQueue` from inputs, run `RunBatch()`, print progress lines to stderr (one line per file: `[2/10] encoding video.mp4 ... 45% 1.2x ETA 0:32`). Print summary table to stdout on completion. Support `--recursive` flag for directory inputs. Support `--output-dir` for directory output mode.

- [x] **Task 5.14:** Add batch-related CLI flags to `internal/cli/run.go`: `--jobs`/`-j` (int, parallel workers), `--recursive`/`-r` (bool), `--sort` (string: size-asc/size-desc/name/duration), `--skip-existing` (bool), `--skip-optimal` (bool), `--in-place` (bool), `--output-dir` (string), `--retry-failed` (bool), `--max-retries` (int). All flags work in both TUI and headless modes (TUI uses them as initial values for batch queue setup).

### Phase Testing

- [x] **Task 5.15:** Write unit tests for Phase 5 packages. **engine/**: test `JobQueue` thread-safety (concurrent `Add`/`Update`/`Next` from multiple goroutines); test `Sort()` produces correct ordering for each mode; test `ShouldSkip()` returns correct results for existing output, already-compressed source, and size threshold; test `SaveQueue`/`LoadQueue` round-trip with various job states; test `ResolveOutput()` directory mode mirrors path structure correctly; test in-place mode verification logic (mock probe results). **Batch integration:** test `RunBatch()` with 3 mock jobs processes all in correct order and sends expected events (mock the actual FFmpeg call). Target: all new tests pass.

---

## Acceptance Criteria

- TUI file picker supports multi-file selection with Space bar
- Batch queue screen shows all queued files with estimated sizes
- Batch progress screen shows overall + per-file progress with real-time updates
- Batch completion screen shows per-file results and totals
- `shrinkray --no-tui -i a.mp4 -i b.mp4 --preset balanced` batch-encodes both files
- `shrinkray --no-tui -i ./videos/ --recursive --preset compact` finds and encodes all videos in directory
- `--jobs 2` runs 2 files in parallel
- `--skip-existing` skips files that already have output
- Queue persists to disk and resumes on next run with pending/failed jobs
- Failed jobs retry up to 2 times automatically
- `--output-dir ./compressed/` mirrors input directory structure
- `--in-place` replaces source after verification
- `go test ./internal/engine/...` passes with new batch tests

---

## Notes

- Default `--jobs 1` is conservative — HW encoders may not benefit from parallelism (GPU is the bottleneck)
- Queue persistence file is in cache dir, not config dir — it's ephemeral state
- In-place mode is the most dangerous feature — triple-verify before replacing source files
- Batch cancel in TUI should cleanly stop current encoding and preserve queue state for resume
- The `--recursive` flag does NOT support glob patterns — it walks directories finding video files by extension

---

## Phase Completion Summary

| Field | Value |
|-------|-------|
| Date completed | 2026-03-21 |
| Implementer | Claude Opus 4.6 |
| What was done | Implemented all 15 tasks: Job/JobQueue with thread-safe mutex access, Sort() with 4 modes, ShouldSkip()/ShouldSkipOptimal() skip logic, file picker multi-select + directory discovery, 3 batch TUI screens (queue/progress/complete), RunBatch() with goroutine worker pool, queue persistence with debounced JSON save/load/resume, retry logic with MaxAttempts, directory output mode with path mirroring via BaseDir, in-place mode with probe verification, headless batch mode with progress lines, all batch CLI flags (--jobs, --recursive, --sort, --skip-existing, --skip-optimal, --in-place, --output-dir, --retry-failed, --max-retries), and comprehensive unit tests |
| Files changed | internal/engine/job.go (new), internal/engine/batch.go (new), internal/engine/batch_persist.go (new), internal/engine/output.go (modified - directory mirroring with BaseDir), internal/engine/job_test.go (new), internal/engine/batch_persist_test.go (new), internal/engine/output_test.go (extended), internal/tui/messages/messages.go (modified - batch screens + messages), internal/tui/screens/filepicker.go (modified - multi-select + directory discovery), internal/tui/screens/batch_queue.go (new), internal/tui/screens/batch_progress.go (new), internal/tui/screens/batch_complete.go (new), internal/tui/app.go (modified - batch screen registration + message handling), internal/cli/run.go (rewritten - batch flags, headless batch, retry-failed) |
| Issues encountered | Bubble Tea v2 KeyPressMsg struct differs from v1 (no Key field for direct construction). Bubbles v2 filepicker does not expose cursor path for space-toggle; adapted multi-select to use Enter for adding files to batch selection and "b" key to proceed to batch queue. |
