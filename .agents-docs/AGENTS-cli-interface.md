# CLI Interface
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Commands

```
shrinkray                           # Launch TUI (or show help if no TTY)
shrinkray [files...]                # Launch TUI with file(s) pre-loaded
shrinkray run [files...]            # Explicit run command (same as default)
shrinkray presets                   # List all available presets
shrinkray presets show <name>       # Show details for a specific preset
shrinkray probe <file>              # Display video metadata (non-interactive)
shrinkray version                   # Show version, build info, FFmpeg version
shrinkray help                      # Show help
```

## Global Flags

```
--no-tui              Disable interactive TUI, use headless mode
--config <path>       Path to config file (default: ~/.config/shrinkray/config.yaml)
--ffmpeg-path <path>  Path to FFmpeg binary (overrides PATH)
--verbose             Enable verbose logging (debug level)
--quiet               Suppress all non-error output (headless)
```

## Encoding Flags

```
-i, --input <path>        Input file or directory
-o, --output <path>       Output file path (single file) or directory
--preset <name>           Preset name (see `shrinkray presets`)
--codec <name>            Video codec: h264, h265, av1, vp9 (default: h265)
--crf <number>            CRF value (0-51 for h264/h265, 0-63 for av1/vp9)
--resolution <WxH>        Output resolution (e.g., 1920x1080)
--fps <number>            Output framerate
--speed-preset <name>     Encoder speed: ultrafast..veryslow (default: medium)
--audio-codec <name>      Audio codec: aac, opus, copy, none (default: aac)
--audio-bitrate <rate>    Audio bitrate: 64k-256k (default: 128k)
--target-size <size>      Target output file size: 10mb, 25mb, 50mb, 1gb
--two-pass                Enable two-pass encoding
--hw-accel                Force hardware acceleration
--no-hw-accel             Disable hardware acceleration
--strip-metadata          Remove EXIF/GPS metadata
--suffix <string>         Output filename suffix (default: _shrunk)
--output-dir <path>       Write outputs to this directory
--in-place                Replace source files (safe: encode → verify → replace)
--overwrite               Overwrite existing output files
--skip-existing           Skip files with existing output
--auto-rename             Auto-rename on conflict: video_shrunk(1).mp4
--audio-channels <mode>   Audio channels: stereo, mono, source
--extra-args <arg>        Extra FFmpeg arguments (repeatable)
--recursive               Recurse into subdirectories
--jobs <n>                Concurrent encoding jobs (default: 2)
--dry-run                 Show what would be done without encoding
--stdin                   Read input paths from stdin (one per line)
```

## Shell Completions

Fang/Cobra generates completions for bash, zsh, fish, and PowerShell:
```bash
shrinkray completion bash > /etc/bash_completion.d/shrinkray
shrinkray completion zsh > ~/.zsh/completions/_shrinkray
```

## Dual Mode

- **TUI mode** (default): Full interactive wizard with all screens
- **Headless mode** (`--no-tui` or non-TTY): Progress line to stderr, suitable for scripts/CI
