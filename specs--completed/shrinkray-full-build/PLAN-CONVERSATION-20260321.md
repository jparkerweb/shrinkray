# Plan Conversation Log — shrinkray Full Build

**Date:** 2026-03-21
**Feature:** shrinkray-full-build
**Participants:** Justin Parker, Claude (AI Planning Agent)

---

## Phase 1: Requirements Analysis

### Context
- **Source document:** IDEA.md — exhaustive product specification (~2400 lines, 22 sections)
- **Project state:** Greenfield — no Go source code exists, only documentation (AGENTS.md, .agents-docs/, IDEA.md)
- **Git:** Initialized during planning session with initial commit (85f6201)

### Questions Asked
1. Additional files/folders to examine? → IDEA.md is authoritative (docs built from it)
2. Reference materials? → None beyond IDEA.md
3. External systems/APIs? → FFmpeg/FFprobe only (confirmed)
4. Scope? → Full project (all 7 phases)
5. Testing preferences? → Delegated to planner

### Requirements Extracted
- 31 functional requirements (FR-1 through FR-31)
- 8 non-functional requirements (NFR-1 through NFR-8)
- Full list in PLAN-DRAFT sections 2.1 and 2.2

### Testing Decision
- **Types:** Unit + Integration (no E2E on TUI — brittle, low ROI)
- **Phase testing:** Yes, after each phase
- **Coverage:** Moderate (~60-80%) focused on engine/, presets/, config/ packages
- **Rationale:** CLI wrapping external process + TUI layer. Engine/presets are pure logic (highly testable). TUI models are hard to unit test meaningfully.

### Sign-off
User approved requirements as presented.

---

## Phase 2: System Context Examination

### Findings
- Greenfield project — no existing codebase
- External systems: FFmpeg (child process, stdin/stdout pipes), FFprobe (JSON output), file system, system file manager
- System boundary diagram created showing layered architecture
- No technical debt

---

## Phase 3: Scope Assessment

### Assessment: LARGE
| Metric | Value | Threshold |
|--------|-------|-----------|
| Phases | 7 | Large >= 6 |
| Requirements | 31 | Large >= 15 |
| Components | 7 | Large >= 7 |
| Integrations | 1 | Small <= 1 |

Hit 3 of 4 large-project thresholds. User opted to continue in single conversation rather than checkpoint split.

---

## Phase 4: Tech Stack

### Decision: doublestar library
**Issue raised:** IDEA.md mentions `**` glob expansion for user-provided patterns like `~/Videos/**/*.mp4`. Go's standard library `filepath.Glob` does not support `**` patterns.

**Research conducted:** Investigated doublestar (v4.10.0, actively maintained), gobwas/glob (stale since 2018), and manual WalkDir approach.

**Decision:** Skip doublestar. Use `filepath.WalkDir()` + extension filtering everywhere. The `--recursive` flag + directory args + stdin piping covers all practical use cases. `**` glob is a niche power-user scenario. Can add doublestar later if requested.

**User confirmed:** Option 2 (skip doublestar).

### Final Stack
All technologies from IDEA.md confirmed: Go 1.23+, Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh v2, Fang, Charm Log, go-ffprobe v2, YAML v3, GoReleaser, golangci-lint, Make.

---

## Phase 5: Architecture Design

### Pattern Selected
Layered Architecture with Message-Passing UI (Elm Architecture via Bubble Tea).

### Devil's Advocate
Considered flat single-package structure. Rejected: spec requires headless mode reusing engine, custom presets independent of both layers, and potential future GUI/web frontends. 3-package separation is essential.

### Key Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Package separation | engine/ (no TUI), tui/ (no engine logic), presets/ (no deps) | Testability, reusability, headless mode |
| Dependency direction | Downward only: cli → tui/headless → engine → presets/config | Clean layering, no circular deps |
| Screen communication | Typed tea.Msg values, not direct function calls | Bubble Tea pattern, decoupled screens |
| Progress channel | Buffered size 1 with non-blocking send | Prevents TUI stalls |
| Error strategy | Return values with custom error types, no panics | Go idiom, wraps FFmpeg exit codes |

### User approved architecture.

---

## Phase 6: Technical Specification

### Implementation Phases
7 phases with 80 total tasks. See PLAN-DRAFT section 5 for full breakdown.

### Risk Assessment
7 risks identified, all with mitigations. Highest concern: Windows signal handling differences (mitigated by stdin `q` as primary cross-platform cancel method).

### Final Confidence: 92%
- Requirements Clarity: 24/25
- Technical Feasibility: 22/25
- Integration Points: 23/25
- Risk Assessment: 23/25
