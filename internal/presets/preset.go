package presets

// Category represents the type of preset.
type Category string

const (
	CategoryQuality  Category = "quality"
	CategoryPurpose  Category = "purpose"
	CategoryPlatform Category = "platform"
)

// Preset defines a complete set of encoding parameters with human-friendly metadata.
type Preset struct {
	Key          string   `json:"key"`          // unique identifier (e.g., "balanced")
	Name         string   `json:"name"`         // display name (e.g., "Balanced")
	Description  string   `json:"description"`  // short description
	Category     Category `json:"category"`     // quality, purpose, or platform
	Codec        string   `json:"codec"`        // h264, h265, av1, vp9
	Container    string   `json:"container"`    // mp4, mkv, webm
	CRF          int      `json:"crf"`          // constant rate factor
	MaxBitrate   string   `json:"maxBitrate"`   // e.g., "8M" (empty = no limit)
	AudioCodec   string   `json:"audioCodec"`   // aac, opus, copy
	AudioBitrate string   `json:"audioBitrate"` // e.g., "128k"
	AudioChannels int     `json:"audioChannels"` // 0 = keep source
	Resolution   string   `json:"resolution"`   // e.g., "1280x720" (empty = keep source)
	MaxFPS       int      `json:"maxFps"`       // 0 = keep source
	TargetSizeMB float64  `json:"targetSizeMb"` // 0 = no target
	TwoPass      bool     `json:"twoPass"`
	SpeedPreset  string   `json:"speedPreset"`  // ultrafast, veryfast, fast, medium, slow, veryslow
	ExtraArgs    []string `json:"extraArgs"`
	Tags         []string `json:"tags"`
	Icon         string   `json:"icon"`
}
