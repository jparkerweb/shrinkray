# AGENTS.md

This file provides guidance to AI coding agents like Claude Code (claude.ai/code), Cursor AI, Codex, Gemini CLI, GitHub Copilot, and other AI coding assistants when working with code in this repository.

---

## Project Overview

**shrinkray** is a cross-platform CLI video compression tool powered by FFmpeg and built with Go using Charm's TUI libraries (Bubble Tea v2, Bubbles v2, Lip Gloss v2) and Fang/Cobra for CLI. It provides a wizard-style TUI that walks users through video compression with smart presets, hardware acceleration auto-detection, and batch processing. Tagline: "Less bytes, same vibes."

- **Language:** Go 1.25.0+
- **External runtime dependency:** FFmpeg/FFprobe (not bundled — detected at runtime)
- **Platforms:** Windows, macOS, Linux (single static binary via GoReleaser)

---

## Development Commands

Common commands for building, testing, and running the project. Use `make ci` to run the full lint → test → build pipeline locally before pushing.

Details: [Development Commands](./.agents-docs/AGENTS-development-commands.md)

---

## Architecture

The codebase follows strict separation of concerns: `engine/` (FFmpeg logic, zero TUI deps), `tui/` (presentation only), `presets/` (data + recommendation, no deps on engine or TUI). The TUI uses Bubble Tea's Model-View-Update pattern with screen routing and inter-screen communication via custom `tea.Msg` types.

Details: [Architecture](./.agents-docs/AGENTS-architecture.md)

---

## Technology Stack & Dependencies

Go with the Charm ecosystem (Bubble Tea v2, Bubbles v2, Lip Gloss v2, Fang, Charm Log), Cobra, plus go-ffprobe v2 and YAML v3. Build tooling: GoReleaser, Make, golangci-lint.

Details: [Technology Stack](./.agents-docs/AGENTS-technology-stack.md)

---

## CLI Interface & Flags

shrinkray supports interactive TUI mode (default) and headless mode (`--no-tui`). Subcommands: `presets`, `probe`, `version`, `completion`. The encode workflow is the root command's default behavior. Extensive encoding flags for headless/scripting use including HW acceleration, audio control, target size, and batch processing.

Details: [CLI Interface](./.agents-docs/AGENTS-cli-interface.md)

---

## Preset System

18 built-in presets across 3 categories: quality-tier (6), purpose-driven (5), and platform-specific (7). Each preset maps to FFmpeg settings. A smart recommendation engine analyzes source video metadata and suggests optimal presets.

Details: [Preset System](./.agents-docs/AGENTS-preset-system.md)

---

## TUI Screens & Styling

11 screens in the wizard flow (splash → filepicker → info → presets → advanced → preview → encoding → complete, plus 3 batch screens), plus a help overlay (`?` key). Two color themes: "Neon Dusk" (default) and "Electric Sunset". Switchable at runtime with Ctrl+T.

Details: [TUI Screens & Styling](./.agents-docs/AGENTS-tui-screens.md)

---

## Versioning & Changelog

This project uses [Semantic Versioning](https://semver.org/). Git tags and `CHANGELOG.md` must stay aligned.

### Rules

- **CHANGELOG.md** follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format with sections: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.
- All notable changes go under `[Unreleased]` as they are made.
- When cutting a release, rename `[Unreleased]` to `[X.Y.Z] - YYYY-MM-DD`, add a fresh `[Unreleased]` section above it, and update the comparison links at the bottom of the file.
- Create a matching git tag: `git tag vX.Y.Z && git push origin vX.Y.Z`. GoReleaser uses the tag to build release binaries.
- **Version bumping**: breaking changes → major, new features → minor, bug fixes → patch.
- The `install-local.ps1` script derives its version string from `git describe --tags`, so the tag is the source of truth.

### Checklist for Releases

1. Move `[Unreleased]` entries to `[X.Y.Z] - YYYY-MM-DD` in `CHANGELOG.md`
2. Add new `[Unreleased]` section header
3. Update comparison links at the bottom of `CHANGELOG.md`
4. Commit: `Release vX.Y.Z`
5. Tag: `git tag vX.Y.Z`
6. Push: `git push origin main --tags`

---

## Git Commit Messages

Use this format for all commits:

```
<description>

<JIRA-Ticket-ID>
AI Assisted
```

- **JIRA ticket ID**: Derive from the branch name format `<prefix>/<TICKET-ID>-description` (uppercase project key + hyphen + integer, e.g., `PCWEB-10968`). Include the ticket ID on its own line after a blank line.
- **AI Assisted**: Always the last line, on its own line directly after the ticket ID.
- If no JIRA ticket is identifiable from the branch name, omit the ticket ID line but still include `AI Assisted` as the last line.

---

## How to Use This File

The sections above contain brief summaries of each topic. For full details, follow the markdown links to the `.agents-docs/` directory. Only read the detail files that are relevant to your current task — there's no need to read everything upfront. The detail files contain implementation guidance, code patterns, and architectural decisions that will help you write code consistent with the project's design.
