# CLI Interface
> Part of [AGENTS.md](../AGENTS.md) — project guidance for AI coding agents.

## Commands

```
shrinkray                           # Launch TUI (or show help if no TTY)
shrinkray [files...]                # Launch TUI with file(s) pre-loaded
shrinkray presets                   # List all available presets
shrinkray presets show <name>       # Show details for a specific preset
shrinkray probe <file>              # Display video metadata (non-interactive)
shrinkray version                   # Show version, build info, FFmpeg version
shrinkray completion [shell]        # Generate shell completions (bash/zsh/fish/powershell)
shrinkray help                      # Show help (built-in via Cobra)
```

Note: There is no explicit `run` subcommand — the encode workflow is the root command's default behavior.

## Global Flags

```
--no-tui              Disable interactive TUI, use headless mode
--config <path>       Path to config file (default: auto-detected platform config dir)
--log-level <level>   Log level: debug, info, warn, error (default: info)
--no-color            Disable color output
```

## Encoding Flags

```
-i, --input <path>        Input file or directory
    --inputs <path>       Additional input video file paths (repeatable)
-o, --output <path>       Output file path (single file) or directory
-p, --preset <name>       Preset name (see `shrinkray presets`)
    --codec <name>        Video codec: h264, h265, av1, vp9
    --crf <number>        CRF value override (0 = use preset default)
    --resolution <WxH>    Output resolution (e.g., 1920x1080)
    --fps <number>        Maximum output framerate
    --speed-preset <name> Encoder speed preset (e.g., fast, medium, slow)
    --target-size <size>  Target output size (e.g., 25mb, 1gb) — forces two-pass
    --two-pass            Force two-pass encoding
    --hw-accel            Auto-detect and use hardware encoder
    --no-hw-accel         Force software encoding
    --audio-codec <name>  Audio codec: aac, opus, copy, none
    --audio-bitrate <rate> Audio bitrate: 64k, 96k, 128k, 192k, 256k
    --audio-channels <mode> Audio channels: stereo, mono, source
    --strip-metadata      Remove all metadata from output
    --keep-metadata       Preserve source metadata (default: true)
    --metadata-title <s>  Set output title metadata
    --suffix <string>     Output filename suffix (default: _shrunk)
    --output-dir <path>   Write outputs to this directory
    --overwrite           Overwrite existing output files
    --auto-rename         Auto-rename on conflict: video_shrunk(1).mp4
    --in-place            Replace source files (safe: encode → verify → replace)
    --skip-existing       Skip files with existing output
    --skip-optimal        Skip files already compressed with target codec
    --extra-args <arg>    Extra FFmpeg arguments (repeatable)
-r, --recursive           Recurse into subdirectories
-j, --jobs <n>            Concurrent encoding jobs (default: 1)
    --retry-failed        Retry failed jobs from persisted queue
    --max-retries <n>     Maximum retry attempts per file (default: 2)
    --dry-run             Show what would be done without encoding
    --stdin               Read input paths from stdin (one per line)
    --open                Open output folder after completion
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
