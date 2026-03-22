package messages

import (
	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
)

// Screen identifies a TUI screen.
type Screen int

const (
	ScreenSplash    Screen = iota
	ScreenFilePicker
	ScreenInfo
	ScreenPresets
	ScreenAdvanced
	ScreenPreview
	ScreenEncoding
	ScreenComplete
	ScreenBatchQueue
	ScreenBatchProgress
	ScreenBatchComplete
)

// String returns a human-readable name for the screen.
func (s Screen) String() string {
	switch s {
	case ScreenSplash:
		return "Splash"
	case ScreenFilePicker:
		return "File Picker"
	case ScreenInfo:
		return "Video Info"
	case ScreenPresets:
		return "Presets"
	case ScreenAdvanced:
		return "Advanced Options"
	case ScreenPreview:
		return "Preview"
	case ScreenEncoding:
		return "Encoding"
	case ScreenComplete:
		return "Complete"
	case ScreenBatchQueue:
		return "Batch Queue"
	case ScreenBatchProgress:
		return "Batch Progress"
	case ScreenBatchComplete:
		return "Batch Complete"
	default:
		return "Unknown"
	}
}

// FileSelectedMsg is sent when the user selects a video file.
type FileSelectedMsg struct {
	Path string
}

// VideoProbeCompleteMsg is sent when probing finishes.
type VideoProbeCompleteMsg struct {
	Info *engine.VideoInfo
	Err  error
}

// PresetSelectedMsg is sent when the user selects a preset.
type PresetSelectedMsg struct {
	Preset presets.Preset
}

// EncodeStartMsg signals the encoding screen to begin encoding.
type EncodeStartMsg struct {
	Opts engine.EncodeOptions
}

// EncodeProgressMsg wraps a progress update from the encoder.
type EncodeProgressMsg struct {
	Update engine.ProgressUpdate
}

// EncodeCompleteMsg is sent when encoding finishes successfully.
type EncodeCompleteMsg struct {
	OutputPath string
	InputSize  int64
	OutputSize int64
}

// EncodeErrorMsg is sent when encoding fails.
type EncodeErrorMsg struct {
	Err error
}

// EncodeCancelMsg is sent when the user cancels encoding.
type EncodeCancelMsg struct{}

// NavigateMsg tells the app to switch to a different screen.
type NavigateMsg struct {
	Screen Screen
}

// BackMsg tells the app to go back one screen.
type BackMsg struct{}

// AdvancedOptions holds the user's choices from the advanced form.
type AdvancedOptions struct {
	Codec         string
	CRF           int
	Resolution    string
	MaxFPS        int
	AudioCodec    string
	AudioBitrate  string
	AudioChannels int
	HWEncoderName string
	Suffix        string
	ConflictMode  string
}

// AdvancedOptionsMsg carries user choices from the advanced options form.
type AdvancedOptionsMsg struct {
	Opts AdvancedOptions
}

// HWDetectedMsg carries hardware encoder detection results.
type HWDetectedMsg struct {
	Encoders []engine.HWEncoder
}

// ThemeToggleMsg toggles the active theme.
type ThemeToggleMsg struct{}

// WindowSizeMsg carries the terminal dimensions.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// FilesSelectedMsg is sent when the user selects multiple files for batch processing.
type FilesSelectedMsg struct {
	Paths []string
}

// BatchQueueReadyMsg signals that the batch queue is set up and ready.
type BatchQueueReadyMsg struct {
	Queue *engine.JobQueue
}

// BatchEventMsg wraps a batch event for the TUI.
type BatchEventMsg struct {
	Event engine.BatchEvent
}

// BatchStartMsg signals the batch progress screen to begin encoding.
type BatchStartMsg struct {
	Queue *engine.JobQueue
}

// BatchCancelMsg is sent when the user cancels batch processing.
type BatchCancelMsg struct {
	CancelAll bool // true = cancel entire batch, false = cancel current file only
}
