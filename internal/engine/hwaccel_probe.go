package engine

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// probeCache caches encoder probe results for the session.
var (
	probeCacheMu sync.Mutex
	probeCache   = make(map[string]bool)
)

// ProbeEncoder tests whether a specific FFmpeg encoder is available by running
// a minimal test encode. It generates a tiny black frame and attempts to encode
// it with the given encoder. Results are cached in memory for the session.
func ProbeEncoder(ctx context.Context, encoder string, codec string) (bool, error) {
	cacheKey := encoder + ":" + codec

	probeCacheMu.Lock()
	if result, ok := probeCache[cacheKey]; ok {
		probeCacheMu.Unlock()
		slog.Debug("encoder probe cache hit", "encoder", encoder, "codec", codec, "available", result)
		return result, nil
	}
	probeCacheMu.Unlock()

	// Run test encode with 5-second timeout
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Generate a tiny black frame and encode it
	args := []string{
		"-f", "lavfi",
		"-i", "color=c=black:s=64x64:d=0.1",
		"-c:v", encoder,
		"-f", "null",
		"-",
	}

	slog.Debug("probing encoder", "encoder", encoder, "codec", codec, "args", args)

	cmd := exec.CommandContext(probeCtx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	available := err == nil

	// Check for common failure messages in stderr
	if !available {
		stderrStr := stderr.String()
		if containsFailureIndicator(stderrStr) {
			slog.Debug("encoder not available",
				"encoder", encoder,
				"codec", codec,
				"reason", extractFailureReason(stderrStr),
			)
		} else if probeCtx.Err() == context.DeadlineExceeded {
			slog.Debug("encoder probe timed out", "encoder", encoder, "codec", codec)
		} else {
			slog.Debug("encoder probe failed",
				"encoder", encoder,
				"codec", codec,
				"error", err,
			)
		}
	} else {
		slog.Debug("encoder available", "encoder", encoder, "codec", codec)
	}

	// Cache the result
	probeCacheMu.Lock()
	probeCache[cacheKey] = available
	probeCacheMu.Unlock()

	return available, nil
}

// ResetProbeCache clears the encoder probe cache. Intended for testing.
func ResetProbeCache() {
	probeCacheMu.Lock()
	probeCache = make(map[string]bool)
	probeCacheMu.Unlock()
}

// containsFailureIndicator checks stderr for known hardware encoder failure messages.
func containsFailureIndicator(stderr string) bool {
	indicators := []string{
		"No NVENC capable devices",
		"Unknown encoder",
		"Encoder not found",
		"Cannot load",
		"No capable devices",
		"driver does not support",
		"Failed to",
		"not available",
		"not supported",
		"Cannot open",
	}
	lower := strings.ToLower(stderr)
	for _, ind := range indicators {
		if strings.Contains(lower, strings.ToLower(ind)) {
			return true
		}
	}
	return false
}

// extractFailureReason extracts a brief failure reason from FFmpeg stderr.
func extractFailureReason(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "no ") ||
			strings.Contains(lower, "cannot") ||
			strings.Contains(lower, "unknown") ||
			strings.Contains(lower, "not found") {
			return strings.TrimSpace(line)
		}
	}
	if len(stderr) > 200 {
		return stderr[:200]
	}
	return stderr
}
