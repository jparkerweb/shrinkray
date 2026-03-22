package engine

import "time"

// ProgressUpdate represents a single progress update from an FFmpeg encode.
type ProgressUpdate struct {
	Percent     float64       `json:"percent"`     // 0-100
	Frame       int64         `json:"frame"`
	FPS         float64       `json:"fps"`
	Speed       float64       `json:"speed"`       // e.g., 1.5 means 1.5x realtime
	Bitrate     string        `json:"bitrate"`     // e.g., "2500kbits/s"
	Size        int64         `json:"size"`        // bytes written so far
	TimeElapsed time.Duration `json:"timeElapsed"`
	ETA         time.Duration `json:"eta"`
	Pass        int           `json:"pass"`        // 0 for single-pass, 1 or 2 for two-pass
	Done        bool          `json:"done"`
	Error       error         `json:"-"`
	Stderr      string        `json:"-"` // FFmpeg stderr output (populated on completion)
}
