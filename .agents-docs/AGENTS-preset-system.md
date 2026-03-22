# Preset System
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Overview

18 built-in presets organized into 3 categories. Each preset maps human-friendly names to FFmpeg encoding parameters.

## Quality-Tier Presets (6)

| Key | Name | Codec | CRF | Speed | Resolution | Audio |
|-----|------|-------|-----|-------|------------|-------|
| lossless | Lossless | H.265 | 0 | slow | original | copy |
| high | High Quality | H.265 | 20 | medium | original | AAC 192k |
| balanced | Balanced | H.265 | 26 | medium | original | AAC 128k |
| low | Low Quality | H.265 | 32 | fast | original | AAC 96k |
| tiny | Tiny | H.265 | 36 | veryfast | 720p cap | AAC 64k |
| potato | Potato | H.264 | 40 | ultrafast | 480p cap | AAC 48k |

## Purpose-Driven Presets (5)

| Key | Name | Use Case |
|-----|------|----------|
| web | Web Optimized | Fast-start MP4, balanced quality |
| email | Email Friendly | ≤25 MB target, smaller resolution |
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
| instagram | Instagram | 250 MB | 1080x1920 / 1080x1080 | AAC mono |
| tiktok | TikTok | 287 MB | 1080x1920 | H.264 baseline |
| youtube | YouTube | — | original | High quality upload |

## Preset Struct

Each preset is a Go struct containing: key, name, description, category, codec, CRF, speed preset, resolution cap, audio codec/bitrate, container, target size, and additional FFmpeg flags.

## Smart Recommendation Engine

Located in `presets/recommend.go`. Analyzes source video metadata (resolution, duration, bitrate, codec, file size) and suggests the top 3 presets ranked by relevance. Displayed in the TUI with estimated output sizes and compression ratios.

## Custom Presets

Users can define custom presets in their config file (`~/.config/shrinkray/config.yaml`). Custom presets can extend built-in presets by overriding specific fields.
