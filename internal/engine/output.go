package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OutputMode defines how output file paths are determined.
type OutputMode string

const (
	OutputModeSuffix    OutputMode = "suffix"
	OutputModeDirectory OutputMode = "directory"
	OutputModeExplicit  OutputMode = "explicit"
	OutputModeInplace   OutputMode = "inplace"
)

// ConflictMode defines behavior when an output file already exists.
type ConflictMode string

const (
	ConflictSkip       ConflictMode = "skip"
	ConflictOverwrite  ConflictMode = "overwrite"
	ConflictAutorename ConflictMode = "autorename"
)

// OutputOptions configures how the output path is resolved.
type OutputOptions struct {
	Mode         OutputMode
	Suffix       string // default "_shrunk"
	Directory    string // output directory for directory mode
	ExplicitPath string // explicit output path
	ConflictMode ConflictMode
	BaseDir      string // base input directory for path mirroring in directory mode
}

// ErrSkipExisting is returned when an output file exists and ConflictMode is skip.
var ErrSkipExisting = fmt.Errorf("output file already exists, skipping")

// ResolveOutput determines the final output path for a given input file.
func ResolveOutput(input string, opts OutputOptions) (string, error) {
	var output string

	switch opts.Mode {
	case OutputModeExplicit:
		if opts.ExplicitPath == "" {
			return "", fmt.Errorf("explicit output mode requires a path")
		}
		output = opts.ExplicitPath

	case OutputModeDirectory:
		if opts.Directory == "" {
			return "", fmt.Errorf("directory output mode requires a directory path")
		}
		// If a base directory is set, mirror the relative path structure
		if opts.BaseDir != "" {
			relPath, err := filepath.Rel(opts.BaseDir, input)
			if err != nil {
				// Fallback to just the filename
				relPath = filepath.Base(input)
			}
			output = filepath.Join(opts.Directory, relPath)
		} else {
			output = filepath.Join(opts.Directory, filepath.Base(input))
		}
		// Ensure the output directory (including subdirectories) exists
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}

	case OutputModeInplace:
		output = input

	default: // suffix mode
		suffix := opts.Suffix
		if suffix == "" {
			suffix = "_shrunk"
		}
		ext := filepath.Ext(input)
		base := strings.TrimSuffix(input, ext)
		output = base + suffix + ext
	}

	// Handle conflicts
	if output != input { // Don't check for inplace mode
		if _, err := os.Stat(output); err == nil {
			switch opts.ConflictMode {
			case ConflictSkip:
				return "", ErrSkipExisting
			case ConflictOverwrite:
				// Allow overwrite — do nothing
			case ConflictAutorename:
				output = autorename(output)
			default:
				output = autorename(output)
			}
		}
	}

	return output, nil
}

// TempPath returns a temporary file path in the same directory as output.
// Uses a short name to avoid Windows 260-char path limit while keeping
// the extension to help FFmpeg choose the right muxer.
// The caller should use defer os.Remove(tempPath) and os.Rename(temp, output) on success.
func TempPath(output string) string {
	dir := filepath.Dir(output)
	ext := filepath.Ext(output)
	base := filepath.Base(output)
	base = strings.TrimSuffix(base, ext)

	// Truncate base to keep total path short on Windows
	const maxBase = 20
	if len(base) > maxBase {
		base = base[:maxBase]
	}

	return filepath.Join(dir, base+".shrinkray.tmp"+ext)
}

// ErrAutorenameExhausted is returned when all 99 auto-rename slots are taken.
var ErrAutorenameExhausted = fmt.Errorf("all auto-rename slots (1-99) are taken")

// autorename adds a numeric suffix to avoid conflicts: video_shrunk(1).mp4
// It tries numbers 1 through 99 and returns an error string if all are taken.
func autorename(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; i <= 99; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}

	// All 99 slots taken — return the path with (99) as fallback
	// The caller should check for this condition
	return fmt.Sprintf("%s(99)%s", base, ext)
}

// Autorename is the exported version of autorename for testing.
func Autorename(path string) string {
	return autorename(path)
}
