package engine

import (
	"regexp"
	"strings"
)

// ffmpegErrorPattern maps an FFmpeg stderr pattern to a user-friendly message template.
type ffmpegErrorPattern struct {
	pattern *regexp.Regexp
	message string
}

// ffmpegErrorPatterns defines the known FFmpeg error patterns and their friendly messages.
var ffmpegErrorPatterns = []ffmpegErrorPattern{
	{
		pattern: regexp.MustCompile(`(?i)no such file or directory`),
		message: "Input file not found",
	},
	{
		pattern: regexp.MustCompile(`(?i)invalid data found when processing input`),
		message: "File appears corrupted or is not a valid video",
	},
	{
		pattern: regexp.MustCompile(`(?i)encoder[^\n]*not found`),
		message: "Codec is not available in your FFmpeg build",
	},
	{
		pattern: regexp.MustCompile(`(?i)unknown encoder '([^']*)'`),
		message: "Codec '$1' is not available in your FFmpeg build",
	},
	{
		pattern: regexp.MustCompile(`(?i)output file.*already exists`),
		message: "Output file already exists (handled by conflict resolution)",
	},
	{
		pattern: regexp.MustCompile(`(?i)permission denied`),
		message: "Cannot write to output location -- check permissions",
	},
	{
		pattern: regexp.MustCompile(`(?i)no space left on device`),
		message: "Disk full -- free up space and try again",
	},
	{
		pattern: regexp.MustCompile(`(?i)could not find codec parameters`),
		message: "File appears corrupted or has unsupported codec parameters",
	},
	{
		pattern: regexp.MustCompile(`(?i)does not contain any stream`),
		message: "File does not contain any video or audio streams",
	},
	{
		pattern: regexp.MustCompile(`(?i)device or resource busy`),
		message: "File is in use by another program",
	},
	{
		pattern: regexp.MustCompile(`(?i)protocol not found`),
		message: "Unsupported file protocol or path format",
	},
}

// ParseFFmpegError takes raw FFmpeg stderr output and returns a user-friendly
// error message. If no known pattern matches, it returns the original stderr
// (truncated to 200 characters if longer).
func ParseFFmpegError(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return "Unknown FFmpeg error"
	}

	for _, ep := range ffmpegErrorPatterns {
		if ep.pattern.MatchString(stderr) {
			// If the message template contains $1, expand the submatch
			if strings.Contains(ep.message, "$1") {
				matches := ep.pattern.FindStringSubmatch(stderr)
				if len(matches) > 1 {
					return strings.ReplaceAll(ep.message, "$1", matches[1])
				}
			}
			return ep.message
		}
	}

	// No pattern matched — return last meaningful lines (the banner is at
	// the top and the actual error is usually near the bottom).
	lines := strings.Split(stderr, "\n")
	var tail []string
	for i := len(lines) - 1; i >= 0 && len(tail) < 6; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			tail = append([]string{trimmed}, tail...)
		}
	}
	result := strings.Join(tail, "\n")
	if len(result) > 500 {
		result = result[:500] + "..."
	}
	return result
}
