package engine

import (
	"testing"
	"time"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

// ---- EstimateSize tests ----

func TestEstimateSize_TargetSizePreset(t *testing.T) {
	info := &VideoInfo{
		Width:     1920,
		Height:    1080,
		Framerate: 30,
		Duration:  60,
		Size:      100000000,
		Bitrate:   8000000,
	}

	preset := presets.Preset{
		Codec:        "h264",
		TargetSizeMB: 10,
	}

	est := EstimateSize(info, preset)
	expected := int64(10 * 1024 * 1024)
	if est != expected {
		t.Errorf("expected %d bytes, got %d", expected, est)
	}
}

func TestEstimateSize_CRFBasedPreset_1080p60s(t *testing.T) {
	info := &VideoInfo{
		Width:        1920,
		Height:       1080,
		Framerate:    30,
		Duration:     60,
		Size:         100000000,
		Bitrate:      10000000,
		AudioBitrate: 128000,
	}

	preset := presets.Preset{
		Key:          "balanced",
		Codec:        "h265",
		CRF:          23,
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	est := EstimateSize(info, preset)
	// 1080p 60s balanced should estimate between 5-50MB
	fiveMB := int64(5 * 1024 * 1024)
	fiftyMB := int64(50 * 1024 * 1024)
	if est < fiveMB || est > fiftyMB {
		t.Errorf("estimate %d bytes (%.1f MB) outside expected range 5-50MB",
			est, float64(est)/1024/1024)
	}
}

func TestEstimateSize_ResolutionScaling(t *testing.T) {
	info := &VideoInfo{
		Width:        3840,
		Height:       2160,
		Framerate:    30,
		Duration:     60,
		Size:         500000000,
		Bitrate:      50000000,
		AudioBitrate: 192000,
	}

	presetKeep := presets.Preset{
		Codec:        "h265",
		CRF:          23,
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	presetScale := presets.Preset{
		Codec:        "h265",
		CRF:          23,
		Resolution:   "1920x1080",
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	estKeep := EstimateSize(info, presetKeep)
	estScale := EstimateSize(info, presetScale)

	if estScale >= estKeep {
		t.Errorf("scaled estimate (%d) should be less than full-res estimate (%d)",
			estScale, estKeep)
	}
}

func TestEstimateSize_NilInfo(t *testing.T) {
	preset := presets.Preset{Codec: "h264", CRF: 23}
	est := EstimateSize(nil, preset)
	if est != 0 {
		t.Errorf("expected 0 for nil info, got %d", est)
	}
}

func TestEstimateSize_MaxFPS(t *testing.T) {
	info := &VideoInfo{
		Width:        1920,
		Height:       1080,
		Framerate:    60,
		Duration:     60,
		Size:         200000000,
		Bitrate:      20000000,
		AudioBitrate: 128000,
	}

	presetNoLimit := presets.Preset{
		Codec:        "h264",
		CRF:          23,
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	presetFPSLimit := presets.Preset{
		Codec:        "h264",
		CRF:          23,
		MaxFPS:       30,
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	estNo := EstimateSize(info, presetNoLimit)
	estLimit := EstimateSize(info, presetFPSLimit)

	if estLimit >= estNo {
		t.Errorf("FPS-limited estimate (%d) should be less than unlimited (%d)", estLimit, estNo)
	}
}

// ---- CalculateBitrate tests ----

func TestCalculateBitrate_KnownValues(t *testing.T) {
	// 10MB target, 60 seconds, 128kbps audio
	targetBytes := int64(10 * 1024 * 1024)
	duration := 60 * time.Second
	audioBitrate := int64(128000)

	bitrate := CalculateBitrate(targetBytes, duration, audioBitrate)

	// Expected: ((10*1024*1024*8) / 60 - 128000) * 0.97
	// = (83886080 / 60 - 128000) * 0.97
	// = (1398101 - 128000) * 0.97
	// = 1270101 * 0.97
	// = 1231998
	if bitrate < 1000000 || bitrate > 1500000 {
		t.Errorf("bitrate %d outside expected range 1.0-1.5 Mbps", bitrate)
	}
}

func TestCalculateBitrate_ZeroDuration(t *testing.T) {
	bitrate := CalculateBitrate(10*1024*1024, 0, 128000)
	if bitrate != 0 {
		t.Errorf("expected 0 for zero duration, got %d", bitrate)
	}
}

func TestCalculateBitrate_AudioExceedsTotal(t *testing.T) {
	// Very small target with high audio bitrate
	bitrate := CalculateBitrate(100, time.Second, 10000000)
	if bitrate != 0 {
		t.Errorf("expected 0 when audio exceeds total, got %d", bitrate)
	}
}

// ---- FeasibilityCheck tests ----

func TestFeasibilityCheck_Acceptable(t *testing.T) {
	// 50MB target, 30s, 720p = plenty of room
	target := int64(50 * 1024 * 1024)
	duration := 30 * time.Second

	result := FeasibilityCheck(target, duration, 1280, 720)
	if result.Status != FeasibilityAcceptable {
		t.Errorf("expected acceptable, got %s: %s", result.Status, result.Message)
	}
}

func TestFeasibilityCheck_Warning(t *testing.T) {
	// 10MB target, 120s, 1080p = tight
	target := int64(10 * 1024 * 1024)
	duration := 120 * time.Second

	result := FeasibilityCheck(target, duration, 1920, 1080)
	if result.Status == FeasibilityAcceptable {
		t.Errorf("expected warning or impossible for tight target, got acceptable (bpp=%.4f)",
			result.BitsPerPixel)
	}
}

func TestFeasibilityCheck_Impossible(t *testing.T) {
	// 1MB target, 300s, 4K = impossible
	target := int64(1 * 1024 * 1024)
	duration := 300 * time.Second

	result := FeasibilityCheck(target, duration, 3840, 2160)
	if result.Status != FeasibilityImpossible {
		t.Errorf("expected impossible, got %s (bpp=%.4f)", result.Status, result.BitsPerPixel)
	}
}

func TestFeasibilityCheck_InvalidDimensions(t *testing.T) {
	result := FeasibilityCheck(10*1024*1024, 60*time.Second, 0, 0)
	if result.Status != FeasibilityImpossible {
		t.Errorf("expected impossible for 0 dimensions, got %s", result.Status)
	}
}

// ---- AdaptiveResolution tests ----

func TestAdaptiveResolution_AcceptableAtSource(t *testing.T) {
	// Big target, short video, low res = no need to downscale
	target := int64(100 * 1024 * 1024)
	duration := 30 * time.Second

	w, h := AdaptiveResolution(target, duration, 1280, 720, 30)
	if w != 1280 || h != 720 {
		t.Errorf("expected source resolution 1280x720, got %dx%d", w, h)
	}
}

func TestAdaptiveResolution_ScalesDown(t *testing.T) {
	// Tight target, long video, 4K = must downscale
	target := int64(10 * 1024 * 1024)
	duration := 300 * time.Second

	w, h := AdaptiveResolution(target, duration, 3840, 2160, 30)
	if w >= 3840 || h >= 2160 {
		t.Errorf("expected downscaled resolution, got %dx%d", w, h)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("resolution should be positive, got %dx%d", w, h)
	}
	// Should be divisible by 2
	if w%2 != 0 || h%2 != 0 {
		t.Errorf("resolution should be divisible by 2, got %dx%d", w, h)
	}
}

func TestAdaptiveResolution_PreservesAspectRatio(t *testing.T) {
	target := int64(5 * 1024 * 1024)
	duration := 120 * time.Second

	w, h := AdaptiveResolution(target, duration, 1920, 1080, 30)
	if w > 0 && h > 0 {
		// Aspect ratio should be approximately 16:9
		ratio := float64(w) / float64(h)
		expected := 1920.0 / 1080.0
		diff := ratio - expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.1 {
			t.Errorf("aspect ratio %.2f differs too much from source %.2f (%dx%d)",
				ratio, expected, w, h)
		}
	}
}

// ---- BPP lookup tests ----

func TestLookupBPP_ExactMatch(t *testing.T) {
	bpp := lookupBPP("h264", 23)
	if bpp <= 0 {
		t.Errorf("expected positive bpp, got %f", bpp)
	}
	if bpp != 0.10 {
		t.Errorf("expected 0.10 for h264 CRF 23, got %f", bpp)
	}
}

func TestLookupBPP_Interpolation(t *testing.T) {
	// CRF 21 should interpolate between CRF 20 (0.15) and CRF 22 (0.11)
	bpp := lookupBPP("h264", 21)
	if bpp <= 0.11 || bpp >= 0.15 {
		t.Errorf("interpolated bpp %f should be between 0.11 and 0.15", bpp)
	}
}

func TestLookupBPP_UnknownCodec(t *testing.T) {
	bpp := lookupBPP("unknown-codec", 23)
	if bpp <= 0 {
		t.Errorf("expected positive bpp for unknown codec (should fallback), got %f", bpp)
	}
}

// ---- parseAudioBitrate tests ----

func TestParseAudioBitrate(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"128k", 128000},
		{"192k", 192000},
		{"96k", 96000},
		{"256K", 256000},
		{"", 128000}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseAudioBitrate(tt.input)
			if result != tt.expected {
				t.Errorf("parseAudioBitrate(%q) = %f, expected %f", tt.input, result, tt.expected)
			}
		})
	}
}
