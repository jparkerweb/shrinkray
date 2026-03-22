package messages

import (
	"errors"
	"testing"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
)

func TestScreenConstants(t *testing.T) {
	screens := []Screen{
		ScreenSplash,
		ScreenFilePicker,
		ScreenInfo,
		ScreenPresets,
		ScreenPreview,
		ScreenEncoding,
		ScreenComplete,
	}

	// All screens should have unique values
	seen := make(map[Screen]bool)
	for _, s := range screens {
		if seen[s] {
			t.Errorf("duplicate Screen constant value: %d", s)
		}
		seen[s] = true
	}

	// All screens should have non-empty string representations
	for _, s := range screens {
		if s.String() == "Unknown" {
			t.Errorf("Screen %d has Unknown string representation", s)
		}
		if s.String() == "" {
			t.Errorf("Screen %d has empty string representation", s)
		}
	}
}

func TestFileSelectedMsg(t *testing.T) {
	msg := FileSelectedMsg{Path: "/path/to/video.mp4"}
	if msg.Path != "/path/to/video.mp4" {
		t.Errorf("FileSelectedMsg.Path = %q, want %q", msg.Path, "/path/to/video.mp4")
	}
}

func TestVideoProbeCompleteMsg(t *testing.T) {
	info := &engine.VideoInfo{
		Path:  "/path/to/video.mp4",
		Codec: "h264",
		Width: 1920,
	}
	msg := VideoProbeCompleteMsg{Info: info, Err: nil}
	if msg.Info.Codec != "h264" {
		t.Errorf("VideoProbeCompleteMsg.Info.Codec = %q, want %q", msg.Info.Codec, "h264")
	}
	if msg.Err != nil {
		t.Errorf("VideoProbeCompleteMsg.Err should be nil")
	}

	// With error
	errMsg := VideoProbeCompleteMsg{Info: nil, Err: errors.New("probe failed")}
	if errMsg.Err == nil {
		t.Error("VideoProbeCompleteMsg.Err should not be nil")
	}
}

func TestPresetSelectedMsg(t *testing.T) {
	p := presets.Preset{Key: "balanced", Name: "Balanced"}
	msg := PresetSelectedMsg{Preset: p}
	if msg.Preset.Key != "balanced" {
		t.Errorf("PresetSelectedMsg.Preset.Key = %q, want %q", msg.Preset.Key, "balanced")
	}
}

func TestEncodeStartMsg(t *testing.T) {
	opts := engine.EncodeOptions{
		Input:  "/input.mp4",
		Output: "/output.mp4",
	}
	msg := EncodeStartMsg{Opts: opts}
	if msg.Opts.Input != "/input.mp4" {
		t.Errorf("EncodeStartMsg.Opts.Input = %q, want %q", msg.Opts.Input, "/input.mp4")
	}
}

func TestEncodeProgressMsg(t *testing.T) {
	update := engine.ProgressUpdate{
		Percent: 50.5,
		FPS:     30,
		Speed:   1.5,
	}
	msg := EncodeProgressMsg{Update: update}
	if msg.Update.Percent != 50.5 {
		t.Errorf("EncodeProgressMsg.Update.Percent = %f, want %f", msg.Update.Percent, 50.5)
	}
}

func TestEncodeCompleteMsg(t *testing.T) {
	msg := EncodeCompleteMsg{
		OutputPath: "/output.mp4",
		InputSize:  1000000,
		OutputSize: 500000,
	}
	if msg.OutputPath != "/output.mp4" {
		t.Errorf("EncodeCompleteMsg.OutputPath = %q, want %q", msg.OutputPath, "/output.mp4")
	}
	if msg.InputSize != 1000000 {
		t.Errorf("EncodeCompleteMsg.InputSize = %d, want %d", msg.InputSize, 1000000)
	}
}

func TestEncodeErrorMsg(t *testing.T) {
	msg := EncodeErrorMsg{Err: errors.New("encode failed")}
	if msg.Err == nil {
		t.Error("EncodeErrorMsg.Err should not be nil")
	}
	if msg.Err.Error() != "encode failed" {
		t.Errorf("EncodeErrorMsg.Err = %q, want %q", msg.Err.Error(), "encode failed")
	}
}

func TestNavigateMsg(t *testing.T) {
	msg := NavigateMsg{Screen: ScreenPresets}
	if msg.Screen != ScreenPresets {
		t.Errorf("NavigateMsg.Screen = %d, want %d", msg.Screen, ScreenPresets)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	msg := WindowSizeMsg{Width: 120, Height: 40}
	if msg.Width != 120 || msg.Height != 40 {
		t.Errorf("WindowSizeMsg = {%d, %d}, want {120, 40}", msg.Width, msg.Height)
	}
}
