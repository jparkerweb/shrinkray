# Preset System
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Overview

18 built-in presets organized into 3 categories. Each preset maps human-friendly names to FFmpeg encoding parameters.

## Quality-Tier Presets (6)

| Key | Name | Codec | CRF | Speed | Resolution | Audio |
|-----|------|-------|-----|-------|------------|-------|
| lossless | Lossless | H.264 | 0 | slow | original | copy |
| ultra | Ultra Quality | H.265 | 16 | slow | original | AAC 256k |
| high | High Quality | H.265 | 20 | medium | original | AAC 192k |
| balanced | Balanced | H.265 | 23 | medium | original | AAC 128k |
| compact | Compact | H.265 | 28 | medium | 720p cap | AAC 96k |
| tiny | Tiny | H.264 | 32 | fast | 480p cap | AAC 64k |

## Purpose-Driven Presets (5)

| Key | Name | Use Case |
|-----|------|----------|
| web | Web Delivery | Fast-start MP4, balanced quality |
| email | Email Friendly | 20 MB target, smaller resolution |
| archive | Archive | Near-lossless long-term storage |
| slideshow | Slideshow/Screencast | Low-motion optimization |
| 4k-to-1080 | 4K → 1080p | Downscale with quality preservation |

## Platform-Specific Presets (7)

| Key | Platform | Max Size | Max Resolution | Notes |
|-----|----------|----------|----------------|-------|
| discord | Discord (Free) | 10 MB | 720p | Target size mode |
| discord-nitro | Discord Nitro | 50 MB | 1080p | Target size mode |
| whatsapp | WhatsApp | 16 MB | 720p | H.264 required |
| twitter | Twitter/X | 512 MB | 1920x1200 | AAC stereo |
| instagram | Instagram | 250 MB | 1920x1080 | AAC 128k (source channels) |
| tiktok | TikTok | 287 MB | 1920x1080 | H.264, 60fps cap |
| youtube | YouTube | — | original | High quality upload |

## Preset Struct

Each preset is a Go struct containing: key, name, description, category, codec, CRF, speed preset, resolution cap, audio codec/bitrate, container, target size, and additional FFmpeg flags.

## Smart Recommendation Engine

Located in `presets/recommend.go`. Analyzes source video metadata (resolution, duration, bitrate, codec, file size) and suggests the top 5 presets ranked by relevance. Displayed in the TUI with estimated output sizes and compression ratios.

## Custom Presets

Users can define custom presets in a separate file (`~/.config/shrinkray/custom_presets.yaml`). Custom presets use the same YAML structure as the Preset struct (camelCase fields). Requires key, name, and codec; must specify either CRF or targetSizeMb. Defaults: container "mp4", category "quality".
