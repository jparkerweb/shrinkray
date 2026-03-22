package engine

import (
	"fmt"
	"time"
)

// FeasibilityStatus represents the quality feasibility of a target size.
type FeasibilityStatus string

const (
	FeasibilityImpossible  FeasibilityStatus = "impossible"
	FeasibilityWarning     FeasibilityStatus = "warning"
	FeasibilityAcceptable  FeasibilityStatus = "acceptable"
)

// FeasibilityResult holds the result of a feasibility analysis.
type FeasibilityResult struct {
	Status       FeasibilityStatus
	BitsPerPixel float64
	Message      string
}

// CalculateBitrate computes the video bitrate needed to hit a target file size.
// targetBytes: desired output size in bytes
// duration: video duration
// audioBitrate: audio bitrate in bits per second
// Returns: video bitrate in bits per second
func CalculateBitrate(targetBytes int64, duration time.Duration, audioBitrate int64) int64 {
	durationSeconds := duration.Seconds()
	if durationSeconds <= 0 {
		return 0
	}

	// Total available bitrate
	totalBitrate := float64(targetBytes*8) / durationSeconds

	// Subtract audio
	videoBitrate := totalBitrate - float64(audioBitrate)

	// Apply safety factor for container overhead
	videoBitrate *= 0.97

	if videoBitrate < 0 {
		return 0
	}

	return int64(videoBitrate)
}

// FeasibilityCheck evaluates whether a target file size is achievable for a given
// video resolution and duration without unacceptable quality loss.
func FeasibilityCheck(targetBytes int64, duration time.Duration, width int, height int) FeasibilityResult {
	durationSeconds := duration.Seconds()
	if durationSeconds <= 0 || width <= 0 || height <= 0 {
		return FeasibilityResult{
			Status:  FeasibilityImpossible,
			Message: "Invalid video dimensions or duration",
		}
	}

	// Approximate video bitrate (assuming ~128kbps audio)
	audioBits := 128000.0
	totalBitrate := float64(targetBytes*8) / durationSeconds
	videoBitrate := totalBitrate - audioBits
	if videoBitrate < 0 {
		videoBitrate = 0
	}

	// Estimate fps (assume 30 if unknown)
	fps := 30.0
	pixels := float64(width * height)
	bpp := videoBitrate / (pixels * fps)

	switch {
	case bpp < 0.01:
		return FeasibilityResult{
			Status:       FeasibilityImpossible,
			BitsPerPixel: bpp,
			Message:      fmt.Sprintf("Target too small for %dx%d — would need %.4f bpp (minimum ~0.01)", width, height, bpp),
		}
	case bpp < 0.04:
		return FeasibilityResult{
			Status:       FeasibilityWarning,
			BitsPerPixel: bpp,
			Message:      fmt.Sprintf("Tight fit at %dx%d — quality loss likely (%.3f bpp)", width, height, bpp),
		}
	default:
		return FeasibilityResult{
			Status:       FeasibilityAcceptable,
			BitsPerPixel: bpp,
			Message:      fmt.Sprintf("Target achievable at %dx%d (%.3f bpp)", width, height, bpp),
		}
	}
}

// Standard resolution steps for adaptive scaling (width x height).
var resolutionSteps = [][2]int{
	{1920, 1080},
	{1280, 720},
	{854, 480},
	{640, 360},
}

// AdaptiveResolution determines the best resolution to achieve a target file size
// with acceptable quality. It iteratively tries lower resolutions until the
// bits-per-pixel reaches an acceptable range.
// Returns the recommended width and height, preserving the source aspect ratio.
func AdaptiveResolution(targetBytes int64, duration time.Duration, width int, height int, fps float64) (int, int) {
	if width <= 0 || height <= 0 || duration <= 0 {
		return width, height
	}

	if fps <= 0 {
		fps = 30
	}

	// First check if source resolution is acceptable
	result := FeasibilityCheck(targetBytes, duration, width, height)
	if result.Status == FeasibilityAcceptable {
		return width, height
	}

	aspectRatio := float64(width) / float64(height)
	isPortrait := height > width

	// Try each resolution step
	for _, step := range resolutionSteps {
		stepW, stepH := step[0], step[1]

		// Skip if this step is larger than source
		if stepW >= width && stepH >= height {
			continue
		}

		// Adjust for portrait orientation
		var targetW, targetH int
		if isPortrait {
			targetH = stepW // swap for portrait
			targetW = int(float64(targetH) * aspectRatio)
		} else {
			targetW = stepW
			targetH = int(float64(targetW) / aspectRatio)
		}

		// Ensure divisible by 2
		targetW = (targetW / 2) * 2
		targetH = (targetH / 2) * 2

		if targetW <= 0 || targetH <= 0 {
			continue
		}

		result = FeasibilityCheck(targetBytes, duration, targetW, targetH)
		if result.Status == FeasibilityAcceptable {
			return targetW, targetH
		}
	}

	// If nothing works, return the smallest step adjusted for aspect ratio
	lastStep := resolutionSteps[len(resolutionSteps)-1]
	if isPortrait {
		h := lastStep[0]
		w := int(float64(h) * aspectRatio)
		return (w / 2) * 2, (h / 2) * 2
	}
	w := lastStep[0]
	h := int(float64(w) / aspectRatio)
	return (w / 2) * 2, (h / 2) * 2
}
