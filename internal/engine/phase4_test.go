package engine

import (
	"strings"
	"testing"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

// --- Task 4.3: MapQuality tests ---

func TestMapQuality_NVENC(t *testing.T) {
	tests := []struct {
		crf      int
		expected string // expected -cq value
	}{
		{18, "20"},
		{23, "25"},
		{28, "30"},
	}

	for _, tt := range tests {
		result := MapQuality(tt.crf, "h264_nvenc")
		if result == nil {
			t.Fatalf("MapQuality(%d, h264_nvenc) returned nil", tt.crf)
		}
		if got := result["-cq"]; got != tt.expected {
			t.Errorf("MapQuality(%d, h264_nvenc)[-cq] = %s, want %s", tt.crf, got, tt.expected)
		}
	}
}

func TestMapQuality_VideoToolbox(t *testing.T) {
	tests := []struct {
		crf      int
		expected string
	}{
		{18, "45"},
		{23, "55"},
		{28, "65"},
	}

	for _, tt := range tests {
		result := MapQuality(tt.crf, "h264_videotoolbox")
		if result == nil {
			t.Fatalf("MapQuality(%d, h264_videotoolbox) returned nil", tt.crf)
		}
		if got := result["-q:v"]; got != tt.expected {
			t.Errorf("MapQuality(%d, h264_videotoolbox)[-q:v] = %s, want %s", tt.crf, got, tt.expected)
		}
	}
}

func TestMapQuality_QSV(t *testing.T) {
	tests := []struct {
		crf      int
		expected string
	}{
		{18, "22"},
		{23, "27"},
		{28, "32"},
	}

	for _, tt := range tests {
		result := MapQuality(tt.crf, "h264_qsv")
		if result == nil {
			t.Fatalf("MapQuality(%d, h264_qsv) returned nil", tt.crf)
		}
		if got := result["-global_quality"]; got != tt.expected {
			t.Errorf("MapQuality(%d, h264_qsv)[-global_quality] = %s, want %s", tt.crf, got, tt.expected)
		}
	}
}

func TestMapQuality_AMF(t *testing.T) {
	tests := []struct {
		crf      int
		expected string
	}{
		{18, "20"},
		{23, "25"},
		{28, "30"},
	}

	for _, tt := range tests {
		result := MapQuality(tt.crf, "h264_amf")
		if result == nil {
			t.Fatalf("MapQuality(%d, h264_amf) returned nil", tt.crf)
		}
		if got := result["-rc"]; got != "cqp" {
			t.Errorf("MapQuality(%d, h264_amf)[-rc] = %s, want cqp", tt.crf, got)
		}
		if got := result["-qp_i"]; got != tt.expected {
			t.Errorf("MapQuality(%d, h264_amf)[-qp_i] = %s, want %s", tt.crf, got, tt.expected)
		}
		if got := result["-qp_p"]; got != tt.expected {
			t.Errorf("MapQuality(%d, h264_amf)[-qp_p] = %s, want %s", tt.crf, got, tt.expected)
		}
	}
}

func TestMapQuality_Interpolation(t *testing.T) {
	// Test a value between known points (CRF 20 is between 18 and 23)
	result := MapQuality(20, "h264_nvenc")
	if result == nil {
		t.Fatal("MapQuality(20, h264_nvenc) returned nil")
	}
	cq := result["-cq"]
	// CRF 18->CQ 20, CRF 23->CQ 25, so CRF 20 should be ~22
	if cq != "22" {
		t.Errorf("MapQuality(20, h264_nvenc)[-cq] = %s, want 22", cq)
	}
}

func TestMapQuality_UnknownEncoder(t *testing.T) {
	result := MapQuality(23, "unknown_encoder")
	if result != nil {
		t.Errorf("MapQuality for unknown encoder should return nil, got %v", result)
	}
}

// --- Task 4.4: BuildArgs with HW encoder ---

func TestBuildArgs_HWEncoder(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:        "h265",
			Container:    "mp4",
			CRF:          23,
			AudioCodec:   "aac",
			AudioBitrate: "128k",
			SpeedPreset:  "medium",
		},
		HWEncoder: "hevc_nvenc",
	}

	args := BuildArgs(opts)

	// Should use HW encoder instead of libx265
	assertContains(t, args, "-c:v", "hevc_nvenc")

	// Should have -hwaccel cuda
	assertContains(t, args, "-hwaccel", "cuda")

	// Should NOT have -crf (NVENC uses -cq)
	for i, arg := range args {
		if arg == "-crf" {
			t.Errorf("HW encoder should not have -crf flag, found at index %d", i)
		}
	}

	// Should have -cq quality flag
	hasCQ := false
	for _, arg := range args {
		if arg == "-cq" {
			hasCQ = true
			break
		}
	}
	if !hasCQ {
		t.Error("NVENC should have -cq quality flag")
	}

	// Should still have -movflags +faststart for mp4
	assertContains(t, args, "-movflags", "+faststart")

	// Should NOT have -preset medium (x265 preset), but may have -preset p4 (nvenc)
	for i, arg := range args {
		if arg == "-preset" && i+1 < len(args) && args[i+1] == "medium" {
			t.Error("NVENC should not use x265 'medium' preset")
		}
	}
}

func TestBuildArgs_HWEncoder_VideoToolbox(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:        "h264",
			Container:    "mp4",
			CRF:          23,
			AudioCodec:   "aac",
			AudioBitrate: "128k",
		},
		HWEncoder: "h264_videotoolbox",
	}

	args := BuildArgs(opts)

	assertContains(t, args, "-c:v", "h264_videotoolbox")
	assertContains(t, args, "-hwaccel", "videotoolbox")

	// Should have -q:v quality flag
	hasQV := false
	for _, arg := range args {
		if arg == "-q:v" {
			hasQV = true
			break
		}
	}
	if !hasQV {
		t.Error("VideoToolbox should have -q:v quality flag")
	}
}

// --- Task 4.5: AV1 support ---

func TestBuildArgs_AV1Software(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:        "av1",
			Container:    "mp4",
			CRF:          23,
			AudioCodec:   "aac",
			AudioBitrate: "128k",
		},
	}

	args := BuildArgs(opts)

	// Should use libsvtav1
	assertContains(t, args, "-c:v", "libsvtav1")

	// CRF should be mapped: 23 -> 32
	assertContains(t, args, "-crf", "32")

	// Should have -b:v 0 for CRF mode
	assertContains(t, args, "-b:v", "0")

	// Should have SVT-AV1 tune param
	assertContains(t, args, "-svtav1-params", "tune=0")
}

func TestBuildArgs_AV1_CRFMapping(t *testing.T) {
	tests := []struct {
		inputCRF    int
		expectedCRF string
	}{
		{18, "24"},
		{23, "32"},
		{28, "38"},
	}

	for _, tt := range tests {
		opts := EncodeOptions{
			Input:  "input.mp4",
			Output: "output.mp4",
			Preset: presets.Preset{
				Codec:      "av1",
				Container:  "mp4",
				CRF:        tt.inputCRF,
				AudioCodec: "aac",
			},
		}

		args := BuildArgs(opts)
		assertContains(t, args, "-crf", tt.expectedCRF)
	}
}

func TestBuildArgs_AV1_HWEncoder(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:      "av1",
			Container:  "mp4",
			CRF:        23,
			AudioCodec: "aac",
		},
		HWEncoder: "av1_nvenc",
	}

	args := BuildArgs(opts)

	// Should use HW encoder
	assertContains(t, args, "-c:v", "av1_nvenc")

	// Should NOT have svtav1-params (that's for software only)
	for _, arg := range args {
		if arg == "-svtav1-params" {
			t.Error("HW AV1 should not have -svtav1-params")
		}
	}
}

// --- Task 4.6: VP9 support ---

func TestBuildArgs_VP9(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.webm",
		Preset: presets.Preset{
			Codec:        "vp9",
			Container:    "webm",
			CRF:          30,
			AudioCodec:   "aac", // Should be forced to opus
			AudioBitrate: "128k",
		},
	}

	args := BuildArgs(opts)

	// Should use libvpx-vp9
	assertContains(t, args, "-c:v", "libvpx-vp9")

	// Should have -b:v 0 (required for CRF mode in VP9)
	assertContains(t, args, "-b:v", "0")

	// Should force opus audio
	assertContains(t, args, "-c:a", "libopus")

	// Should have -row-mt 1
	assertContains(t, args, "-row-mt", "1")

	// Should NOT have -movflags for webm
	for _, arg := range args {
		if arg == "-movflags" {
			t.Error("webm container should not have -movflags")
		}
	}
}

func TestBuildArgs_VP9_OpusForced(t *testing.T) {
	// Even when preset says AAC, VP9 forces opus
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.webm",
		Preset: presets.Preset{
			Codec:        "vp9",
			Container:    "webm",
			CRF:          30,
			AudioCodec:   "aac",
			AudioBitrate: "192k",
		},
	}

	args := BuildArgs(opts)
	assertContains(t, args, "-c:a", "libopus")
}

func TestBuildArgs_VP9_AudioCopy_NotForced(t *testing.T) {
	// When audio is "copy", it should stay as copy
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.webm",
		Preset: presets.Preset{
			Codec:      "vp9",
			Container:  "webm",
			CRF:        30,
			AudioCodec: "copy",
		},
	}

	args := BuildArgs(opts)
	assertContains(t, args, "-c:a", "copy")
}

// --- Task 4.7: Portrait handling ---

func TestBuildArgs_PortraitSocial(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:      "h264",
			Container:  "mp4",
			CRF:        23,
			Resolution: "1920x1080",
			AudioCodec: "aac",
			Category:   presets.CategoryPlatform,
			Tags:       []string{"social"},
		},
		SourceInfo: &VideoInfo{
			Width:  1080,
			Height: 1920, // portrait
		},
	}

	args := BuildArgs(opts)

	// Should have -vf with swapped dimensions (portrait)
	hasVF := false
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			filter := args[i+1]
			hasVF = true
			// For portrait social, width should be smaller than height in the scale
			if !strings.Contains(filter, "1080") || !strings.Contains(filter, "1920") {
				t.Errorf("portrait social scale should contain 1080 and 1920, got: %s", filter)
			}
			break
		}
	}
	if !hasVF {
		t.Error("portrait social video should have -vf scale filter")
	}
}

func TestBuildArgs_PortraitTikTok_Padding(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Key:        "tiktok",
			Codec:      "h264",
			Container:  "mp4",
			CRF:        23,
			Resolution: "1920x1080",
			AudioCodec: "aac",
			Category:   presets.CategoryPlatform,
			Tags:       []string{"tiktok", "social", "portrait"},
		},
		SourceInfo: &VideoInfo{
			Width:  720,
			Height: 1280, // portrait
		},
	}

	args := BuildArgs(opts)

	// Should have padding filter for TikTok
	hasVF := false
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			filter := args[i+1]
			hasVF = true
			if !strings.Contains(filter, "pad") {
				t.Errorf("TikTok portrait should have pad filter, got: %s", filter)
			}
			break
		}
	}
	if !hasVF {
		t.Error("TikTok portrait video should have -vf filter with padding")
	}
}

func TestBuildArgs_PortraitInstagram_Padding(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Key:        "instagram",
			Codec:      "h264",
			Container:  "mp4",
			CRF:        23,
			Resolution: "1920x1080",
			AudioCodec: "aac",
			Category:   presets.CategoryPlatform,
			Tags:       []string{"instagram", "social"},
		},
		SourceInfo: &VideoInfo{
			Width:  720,
			Height: 1280, // portrait
		},
	}

	args := BuildArgs(opts)

	// Should have padding filter for Instagram
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			filter := args[i+1]
			if !strings.Contains(filter, "pad") {
				t.Errorf("Instagram portrait should have pad filter, got: %s", filter)
			}
			return
		}
	}
	t.Error("Instagram portrait video should have -vf filter with padding")
}

func TestBuildArgs_LandscapeNoSwap(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:      "h264",
			Container:  "mp4",
			CRF:        23,
			Resolution: "1920x1080",
			AudioCodec: "aac",
			Category:   presets.CategoryPlatform,
			Tags:       []string{"social"},
		},
		SourceInfo: &VideoInfo{
			Width:  1920,
			Height: 1080, // landscape
		},
	}

	args := BuildArgs(opts)

	// Should have normal scale filter (no swap, no padding)
	for i, arg := range args {
		if arg == "-vf" && i+1 < len(args) {
			filter := args[i+1]
			if strings.Contains(filter, "pad") {
				t.Errorf("landscape video should not have pad filter, got: %s", filter)
			}
			return
		}
	}
}

// --- HWEncoder helper tests ---

func TestHWEncoderName(t *testing.T) {
	tests := []struct {
		hwName   string
		codec    string
		expected string
	}{
		{"nvenc", "h264", "h264_nvenc"},
		{"nvenc", "h265", "hevc_nvenc"},
		{"nvenc", "av1", "av1_nvenc"},
		{"qsv", "h264", "h264_qsv"},
		{"qsv", "h265", "hevc_qsv"},
		{"amf", "h264", "h264_amf"},
		{"nvenc", "unknown", ""},
		{"unknown", "h264", ""},
	}

	for _, tt := range tests {
		got := HWEncoderName(tt.hwName, tt.codec)
		if got != tt.expected {
			t.Errorf("HWEncoderName(%q, %q) = %q, want %q", tt.hwName, tt.codec, got, tt.expected)
		}
	}
}

func TestSupportsHWTwoPass(t *testing.T) {
	tests := []struct {
		encoder  string
		expected bool
	}{
		{"hevc_nvenc", false},
		{"h264_nvenc", false},
		{"h264_qsv", true},
		{"hevc_qsv", true},
		{"h264_videotoolbox", true},
		{"h264_amf", false},
		{"libx265", true},
	}

	for _, tt := range tests {
		got := SupportsHWTwoPass(tt.encoder)
		if got != tt.expected {
			t.Errorf("SupportsHWTwoPass(%q) = %v, want %v", tt.encoder, got, tt.expected)
		}
	}
}

// --- AV1 CRF mapping ---

func TestMapAV1CRF(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{18, 24},
		{23, 32},
		{28, 38},
	}

	for _, tt := range tests {
		got := mapAV1CRF(tt.input)
		if got != tt.expected {
			t.Errorf("mapAV1CRF(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

// --- Codec to lib mapping ---

func TestCodecToLib(t *testing.T) {
	tests := []struct {
		codec    string
		expected string
	}{
		{"h264", "libx264"},
		{"h265", "libx265"},
		{"hevc", "libx265"},
		{"av1", "libsvtav1"},
		{"vp9", "libvpx-vp9"},
		{"unknown", "libx265"},
	}

	for _, tt := range tests {
		got := codecToLib(tt.codec)
		if got != tt.expected {
			t.Errorf("codecToLib(%q) = %q, want %q", tt.codec, got, tt.expected)
		}
	}
}
