# TUI Screens & Styling
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Screen Flow

```
Splash → FilePicker → Info → Presets → Advanced → Preview → Encoding → Complete
                        ↓ (batch)
                   BatchQueue → BatchProgress → BatchComplete
```

## Screen Inventory

| # | Screen | File | Purpose |
|---|--------|------|---------|
| 1 | Splash | `splash.go` | ASCII logo, FFmpeg version, HW capabilities |
| 2 | FilePicker | `filepicker.go` | File browser + path input, batch multi-select |
| 3 | Info | `info.go` | Source video metadata card (resolution, codec, duration, etc.) |
| 4 | Presets | `presets.go` | Preset grid with smart recommendations highlighted |
| 5 | Advanced | `advanced.go` | Full options form (codec, CRF, resolution, audio, etc.) |
| 6 | Preview | `preview.go` | Before/after comparison with estimated output size |
| 7 | Encoding | `encoding.go` | Real-time progress bar, ETA, speed, FPS, bitrate |
| 8 | Complete | `complete.go` | Results summary with bar chart, savings percentage |
| 9a | BatchQueue | `batch_queue.go` | File list with per-file status, preset assignment |
| 9b | BatchProgress | `batch_progress.go` | Multi-file progress with current + overall bars |
| 9c | BatchComplete | `batch_complete.go` | Batch results table with totals |
| — | Help Overlay | `help.go` | Global/context-sensitive key bindings (triggered by `?`, composited on current screen) |

## Persistent UI Elements

- **Header bar**: App name (left), step indicator + help/quit hints (right)
- **Footer bar**: Context-sensitive keyboard shortcuts for the active screen

## Color Themes

### Neon Dusk (default)
| Role | Color | Hex |
|------|-------|-----|
| Primary | Electric violet | `#7B2FF7` |
| Secondary | Hot pink | `#FF2D95` |
| Tertiary | Cyan | `#00F0FF` |
| Success | Neon green | `#39FF14` |
| Warning | Amber | `#FFAB00` |
| Muted | Soft gray | `#6C6C8A` |
| Text | Near-white | `#E8E8F0` |

### Electric Sunset (alt)
| Role | Color | Hex |
|------|-------|-----|
| Primary | Coral red | `#FF6B6B` |
| Secondary | Golden yellow | `#FFD93D` |
| Accent | Warm orange | `#FF8E53` |
| Muted | Gray | `#6C6C6C` |

Theme switching: `Ctrl+T` at runtime. Persisted in config.

## Global Key Bindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit (with confirmation if encoding) |
| `?` | Toggle help overlay |
| `Ctrl+T` | Switch theme |
| `Esc` | Go back one screen |

## Styling Notes

- All styles defined in `tui/styles.go` using Lip Gloss v2
- Theme struct in `tui/theme.go` — all colors referenced through the theme, never hardcoded
- Progress bars use gradient fills (primary → secondary)
- ASCII logo rendered with gradient coloring
