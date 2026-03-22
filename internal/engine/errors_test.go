package engine

import (
	"strings"
	"testing"
)

func TestParseFFmpegError_NoSuchFile(t *testing.T) {
	stderr := "/path/to/video.mp4: No such file or directory"
	result := ParseFFmpegError(stderr)
	if result != "Input file not found" {
		t.Errorf("expected 'Input file not found', got %q", result)
	}
}

func TestParseFFmpegError_InvalidData(t *testing.T) {
	stderr := "Invalid data found when processing input"
	result := ParseFFmpegError(stderr)
	if result != "File appears corrupted or is not a valid video" {
		t.Errorf("expected corruption message, got %q", result)
	}
}

func TestParseFFmpegError_EncoderNotFound(t *testing.T) {
	stderr := "Encoder hevc_nvenc not found"
	result := ParseFFmpegError(stderr)
	if result != "Codec is not available in your FFmpeg build" {
		t.Errorf("expected codec not available message, got %q", result)
	}
}

func TestParseFFmpegError_UnknownEncoder(t *testing.T) {
	stderr := "Unknown encoder 'libx265special'"
	result := ParseFFmpegError(stderr)
	if !strings.Contains(result, "libx265special") {
		t.Errorf("expected message to contain encoder name, got %q", result)
	}
}

func TestParseFFmpegError_PermissionDenied(t *testing.T) {
	stderr := "Permission denied"
	result := ParseFFmpegError(stderr)
	if !strings.Contains(result, "permissions") {
		t.Errorf("expected permission message, got %q", result)
	}
}

func TestParseFFmpegError_DiskFull(t *testing.T) {
	stderr := "No space left on device"
	result := ParseFFmpegError(stderr)
	if !strings.Contains(result, "Disk full") {
		t.Errorf("expected disk full message, got %q", result)
	}
}

func TestParseFFmpegError_EmptyStderr(t *testing.T) {
	result := ParseFFmpegError("")
	if result != "Unknown FFmpeg error" {
		t.Errorf("expected 'Unknown FFmpeg error', got %q", result)
	}
}

func TestParseFFmpegError_UnknownPattern(t *testing.T) {
	stderr := "Some completely unknown error message from ffmpeg"
	result := ParseFFmpegError(stderr)
	if result != stderr {
		t.Errorf("expected original stderr for unknown pattern, got %q", result)
	}
}

func TestParseFFmpegError_LongStderrTruncated(t *testing.T) {
	stderr := strings.Repeat("a", 300)
	result := ParseFFmpegError(stderr)
	if len(result) > 210 { // 200 + "..."
		t.Errorf("expected truncated output, got length %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("expected truncated output to end with '...'")
	}
}
