package engine

import (
	"strings"
	"testing"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

func TestBuildArgs_BalancedPreset(t *testing.T) {
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
	}

	args := BuildArgs(opts)

	assertContains(t, args, "-i", "input.mp4")
	assertContains(t, args, "-c:v", "libx265")
	assertContains(t, args, "-crf", "23")
	assertContains(t, args, "-preset", "medium")
	assertContains(t, args, "-c:a", "aac")
	assertContains(t, args, "-b:a", "128k")
	assertContains(t, args, "-movflags", "+faststart")
	assertContains(t, args, "-progress", "pipe:1")
	assertHasFlag(t, args, "-y")

	// Output should be last
	if args[len(args)-1] != "output.mp4" {
		t.Errorf("expected output.mp4 as last arg, got %s", args[len(args)-1])
	}
}

func TestBuildArgs_H264Preset(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:        "h264",
			Container:    "mp4",
			CRF:          0,
			AudioCodec:   "copy",
			SpeedPreset:  "slow",
		},
	}

	args := BuildArgs(opts)

	assertContains(t, args, "-c:v", "libx264")
	assertContains(t, args, "-crf", "0")
	assertContains(t, args, "-preset", "slow")
	assertContains(t, args, "-c:a", "copy")
}

func TestBuildArgs_WithResolution(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:        "h265",
			Container:    "mp4",
			CRF:          28,
			Resolution:   "1280x720",
			AudioCodec:   "aac",
			AudioBitrate: "96k",
			SpeedPreset:  "medium",
		},
	}

	args := BuildArgs(opts)

	// Should contain -vf with scale filter
	hasVF := false
	for _, arg := range args {
		if strings.Contains(arg, "scale") {
			hasVF = true
			break
		}
	}
	if !hasVF {
		t.Error("expected -vf with scale filter for resolution preset")
	}
}

func TestBuildArgs_CRFOverride(t *testing.T) {
	crf := 18
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
		CRFOverride: &crf,
	}

	args := BuildArgs(opts)
	assertContains(t, args, "-crf", "18")
}

func TestBuildArgs_AudioCopy(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:      "h264",
			Container:  "mp4",
			CRF:        0,
			AudioCodec: "copy",
		},
	}

	args := BuildArgs(opts)
	assertContains(t, args, "-c:a", "copy")

	// Should NOT have -b:a when copying
	for i, arg := range args {
		if arg == "-b:a" {
			t.Errorf("should not have -b:a with audio copy, found at index %d", i)
		}
	}
}

func TestBuildArgs_AudioNone(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		Preset: presets.Preset{
			Codec:      "h265",
			Container:  "mp4",
			CRF:        23,
			AudioCodec: "none",
		},
	}

	args := BuildArgs(opts)

	found := false
	for _, arg := range args {
		if arg == "-an" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected -an flag for audio=none")
	}
}

func TestBuildArgs_WebMContainer(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.webm",
		Preset: presets.Preset{
			Codec:        "vp9",
			Container:    "webm",
			CRF:          30,
			AudioCodec:   "opus",
			AudioBitrate: "128k",
		},
	}

	args := BuildArgs(opts)
	assertContains(t, args, "-c:v", "libvpx-vp9")
	assertContains(t, args, "-c:a", "libopus")

	// Should NOT have -movflags for webm
	for _, arg := range args {
		if arg == "-movflags" {
			t.Error("should not have -movflags for webm container")
		}
	}
}

// assertContains checks that args contains the key-value pair.
func assertContains(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i, arg := range args {
		if arg == key && i+1 < len(args) && args[i+1] == value {
			return
		}
	}
	t.Errorf("expected args to contain %s %s, got: %v", key, value, args)
}

// assertHasFlag checks that a single flag (no value) exists in args.
func assertHasFlag(t *testing.T, args []string, flag string) {
	t.Helper()
	for _, arg := range args {
		if arg == flag {
			return
		}
	}
	t.Errorf("expected args to contain flag %s, got: %v", flag, args)
}
