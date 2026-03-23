---
name: shrinkray-update-docs
description: Audit and update all project documentation (README, AGENTS.md, and .agents-docs/ files) to match the current codebase. Use when the user says "update docs", "sync docs", "audit docs", "docs are stale", "refresh documentation", or after significant code changes. Also use when the user types "/shrinkray-update-docs".
user-invocable: true
---

# Documentation Audit & Update

This skill systematically audits every documentation file against the actual codebase and fixes all discrepancies. It uses parallel subagents to maximize speed.

## Documentation Files to Audit

| File | Checked Against |
|------|----------------|
| `README.md` | CLI flags, presets, config struct, Go version, install scripts |
| `AGENTS.md` | All summary claims (preset count, screen count, dependency names, Go version) |
| `.agents-docs/AGENTS-architecture.md` | File tree under `internal/`, package dependencies |
| `.agents-docs/AGENTS-cli-interface.md` | Flags in `internal/cli/root.go` + `run.go`, subcommands, defaults |
| `.agents-docs/AGENTS-technology-stack.md` | `go.mod` deps, `Makefile`, `.goreleaser.yaml`, CI workflows |
| `.agents-docs/AGENTS-preset-system.md` | `internal/presets/*.go`, recommendation count, custom preset path |
| `.agents-docs/AGENTS-tui-screens.md` | `internal/tui/screens/*.go`, `internal/tui/style/`, themes, key bindings |
| `.agents-docs/AGENTS-development-commands.md` | `Makefile` targets, `scripts/` directory |

## Execution

Launch **up to 10 parallel Explore subagents**, each responsible for one audit area. Each agent should read the documentation AND the relevant source files, then report every discrepancy found. The audit areas are:

### Agent 1: README — CLI Flags
Read `README.md` flags table and compare against every flag registered in `internal/cli/root.go` (global flags) and `internal/cli/run.go` (`registerRunFlags` function). Report flags missing from README, flags in README not in code, wrong descriptions, and wrong defaults.

### Agent 2: README — Presets, Config, Versions
Read `README.md` presets tables, config example, and version requirements. Compare against `internal/presets/quality.go`, `internal/presets/purpose.go`, `internal/presets/platform.go` (preset names, keys, descriptions, counts), `internal/config/config.go` (struct field names, YAML tags, defaults), and `go.mod` (Go version). Check that YAML field names in README match the `yaml:` struct tags in code (camelCase).

### Agent 3: README — Install Scripts & Release Workflow
Verify install scripts referenced in README exist in `scripts/` directory. Verify release workflow description against `.github/workflows/release.yml`.

### Agent 4: AGENTS.md — Summary Claims
Check every factual claim: Go version vs `go.mod`, preset count (6+5+7=18) vs actual preset source files, screen count vs `internal/tui/screens/*.go` file list, dependency names (Bubble Tea, Bubbles, Lip Gloss, Fang, etc.) vs `go.mod`, theme names vs `internal/tui/style/theme.go`.

### Agent 5: Architecture Doc
Compare `AGENTS-architecture.md` file tree against actual files under `internal/`. Check for added/removed/renamed files, misplaced entries (e.g. file listed in wrong package), missing subdirectories (like `tui/style/`). Verify dependency rules.

### Agent 6: CLI Interface Doc
Compare `AGENTS-cli-interface.md` against `internal/cli/root.go` and `internal/cli/run.go`. Check every flag, default value, subcommand, and documented behavior. Ensure no phantom flags (documented but not implemented) and no missing flags.

### Agent 7: Technology Stack Doc
Compare `AGENTS-technology-stack.md` against `go.mod` (all direct dependencies), `Makefile`, `.goreleaser.yaml`, and `.github/workflows/`. Check for missing deps, removed deps, wrong versions, and stale build tool references.

### Agent 8: Preset System Doc
Compare `AGENTS-preset-system.md` against all files in `internal/presets/`. Verify preset names, categories, counts, FFmpeg parameter mappings, recommendation engine count (top N), and custom preset file path/format (check `custom_presets.go`).

### Agent 9: TUI Screens Doc
Compare `AGENTS-tui-screens.md` against `internal/tui/screens/*.go`, `internal/tui/style/*.go`, `internal/tui/app.go`, `internal/tui/keys.go`, `internal/tui/theme.go`. Verify screen count, screen names, key bindings (especially footer hints per screen), theme colors, and style file locations.

### Agent 10: Development Commands Doc
Compare `AGENTS-development-commands.md` against `Makefile` (all targets) and `scripts/` directory (all scripts). Report missing targets, missing scripts, wrong command names, and outdated instructions.

## After Agents Complete

1. **Compile all findings** into a categorized list
2. **Apply fixes** to each file — prefer `Edit` for targeted changes, `Write` only for full rewrites
3. **Organize README flags** logically by category (input, encoding, audio, HW accel, output, batch, metadata, global)
4. **Preserve formatting conventions** — match existing table styles, heading levels, and code block formats
5. **Show the user a summary** of all changes made, grouped by file

## Key Things That Frequently Drift

These are the most common sources of doc staleness — pay extra attention:

- **New CLI flags** added to `run.go` but not to README or CLI doc
- **Preset descriptions** that get more detail in code but README stays short
- **Config YAML tags** — code uses camelCase (`skipExisting`) but docs may drift to snake_case
- **Recommendation engine "top N"** count changing in `recommend.go`
- **New files in `engine/`** not added to architecture doc file tree
- **`tui/style/` subdirectory** files not reflected in architecture doc
- **Makefile targets** or new scripts not documented in dev commands doc
- **Go version** bumps in `go.mod` not reflected in README requirements
- **Dependency additions/removals** in `go.mod` not reflected in tech stack doc

## Validation

After all edits, run `make ci` to ensure no code was accidentally broken. If it fails, the doc changes likely introduced no code issues — investigate and fix.
