# AGENTS.md

This file provides guidance to AI coding agents like Claude Code (claude.ai/code), Cursor AI, Codex, Gemini CLI, GitHub Copilot, and other AI coding assistants when working with code in this repository.

---

## Project Overview

**shrinkray** is a cross-platform CLI video compression tool powered by FFmpeg and built with Go using Charm's TUI libraries (Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh v2). It provides a wizard-style TUI that walks users through video compression with smart presets, hardware acceleration auto-detection, and batch processing. Tagline: "Less bytes, same vibes."

- **Language:** Go 1.23.0+
- **External runtime dependency:** FFmpeg/FFprobe (not bundled — detected at runtime)
- **Platforms:** Windows, macOS, Linux (single static binary via GoReleaser)

---

## Development Commands

Common commands for building, testing, and running the project.

Details: [Development Commands](./.agents-docs/AGENTS-development-commands.md)

---

## Architecture

The codebase follows strict separation of concerns: `engine/` (FFmpeg logic, zero TUI deps), `tui/` (presentation only), `presets/` (data + recommendation, no deps on engine or TUI). The TUI uses Bubble Tea's Model-View-Update pattern with screen routing and inter-screen communication via custom `tea.Msg` types.

Details: [Architecture](./.agents-docs/AGENTS-architecture.md)

---

## Technology Stack & Dependencies

Go with the Charm ecosystem (Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh v2, Fang), plus go-ffprobe v2 and YAML v3. Build tooling: GoReleaser, Make, golangci-lint.

Details: [Technology Stack](./.agents-docs/AGENTS-technology-stack.md)

---

## CLI Interface & Flags

shrinkray supports interactive TUI mode (default) and headless mode (`--no-tui`). Subcommands: `presets`, `probe`, `version`, `help`. Extensive encoding flags for headless/scripting use.

Details: [CLI Interface](./.agents-docs/AGENTS-cli-interface.md)

---

## Preset System

18 built-in presets across 3 categories: quality-tier (6), purpose-driven (5), and platform-specific (7). Each preset maps to FFmpeg settings. A smart recommendation engine analyzes source video metadata and suggests optimal presets.

Details: [Preset System](./.agents-docs/AGENTS-preset-system.md)

---

## TUI Screens & Styling

11 screens in the wizard flow (splash → filepicker → info → presets → advanced → preview → encoding → complete, plus 3 batch screens). Two color themes: "Neon Dusk" (default) and "Electric Sunset". Switchable at runtime with Ctrl+T.

Details: [TUI Screens & Styling](./.agents-docs/AGENTS-tui-screens.md)

---

## How to Use This File

The sections above contain brief summaries of each topic. For full details, follow the markdown links to the `.agents-docs/` directory. Only read the detail files that are relevant to your current task — there's no need to read everything upfront. The detail files contain implementation guidance, code patterns, and architectural decisions that will help you write code consistent with the project's design.
