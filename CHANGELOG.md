# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- "Delete original" option (`d` key) on complete screen — deletes source file and renames output to original filename
- Hardware encoder automatic fallback — retries with software encoding when HW encoder fails (TUI and batch modes)
- FFmpeg stderr output included in error messages for better diagnostics
- Empty output file validation — catches silent FFmpeg failures before reporting success
- Temp file cleanup on all encoding error paths

### Fixed
- `explorer /select` not opening correct folder on Windows — Go's argument escaping mangled quotes through `cmd /c`; now uses `SysProcAttr.CmdLine` for exact command line control
- Error messages truncated in TUI — errors now wrap to terminal width using lipgloss
- FFmpeg error display showing banner/copyright instead of actual error — now shows last meaningful lines of stderr
- Duplicate key hints on complete screen removed (app status bar already shows them)
- Encoding reported "Complete" with 0-byte output — now validates temp file size before moving
- Race condition where FFmpeg stderr was dropped — `Done` update from progress parser was forwarded before `cmd.Wait()` captured stderr; relay now defers `Done` until after process exit

### Changed
- `ParseFFmpegError` fallback returns last 6 non-empty lines of stderr (up to 500 chars) instead of first 200 chars
- `OpenFolder`/`OpenFolderCommand` split into platform-specific files (`open_windows.go`, `open_unix.go`) for correct Windows syscall handling
- Verbose error messages for rename failures now show source path, destination path, and underlying cause on separate lines

## [0.1.0] - 2025-05-01

### Added
- Initial release
- Interactive TUI wizard (splash → file picker → info → presets → advanced → preview → encoding → complete)
- Batch processing with visual queue management
- 18 built-in presets across quality-tier, purpose-driven, and platform-specific categories
- Smart preset recommendation engine based on source video analysis
- Hardware acceleration auto-detection (NVENC, QSV, AMF, VideoToolbox, VAAPI)
- Two-pass encoding support with target file size mode
- Headless CLI mode (`--no-tui`) for scripting
- Cross-platform support (Windows, macOS, Linux)
- Drive selector in file picker for Windows
- Two color themes ("Neon Dusk" and "Electric Sunset") switchable at runtime
- Help overlay (`?` key) on all screens
- GoReleaser build pipeline with GitHub Actions CI/CD

[Unreleased]: https://github.com/jparkerweb/shrinkray/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jparkerweb/shrinkray/releases/tag/v0.1.0
