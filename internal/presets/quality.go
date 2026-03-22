package presets

// qualityPresets returns the 6 quality-tier presets.
func qualityPresets() []Preset {
	return []Preset{
		{
			Key:          "lossless",
			Name:         "Lossless",
			Description:  "Mathematically identical output — no quality loss whatsoever",
			Category:     CategoryQuality,
			Codec:        "h264",
			Container:    "mp4",
			CRF:          0,
			AudioCodec:   "copy",
			AudioBitrate: "",
			SpeedPreset:  "slow",
			Tags:         []string{"lossless", "archive", "no-loss", "quality"},
			Icon:         "\U0001f48e", // diamond
		},
		{
			Key:          "ultra",
			Name:         "Ultra Quality",
			Description:  "Near-lossless encoding with H.265 — ideal for archival or mastering",
			Category:     CategoryQuality,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          16,
			AudioCodec:   "aac",
			AudioBitrate: "256k",
			SpeedPreset:  "slow",
			Tags:         []string{"ultra", "high-quality", "archive", "mastering"},
			Icon:         "\u2b50", // star
		},
		{
			Key:          "high",
			Name:         "High Quality",
			Description:  "Excellent quality with good compression — great for personal collections",
			Category:     CategoryQuality,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          20,
			AudioCodec:   "aac",
			AudioBitrate: "192k",
			SpeedPreset:  "medium",
			Tags:         []string{"high", "quality", "collection"},
			Icon:         "\U0001f525", // fire
		},
		{
			Key:          "balanced",
			Name:         "Balanced",
			Description:  "Best trade-off between quality and file size — recommended for most uses",
			Category:     CategoryQuality,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          23,
			AudioCodec:   "aac",
			AudioBitrate: "128k",
			SpeedPreset:  "medium",
			Tags:         []string{"balanced", "default", "recommended", "general"},
			Icon:         "\u2696\ufe0f", // balance scale
		},
		{
			Key:          "compact",
			Name:         "Compact",
			Description:  "Smaller files capped at 720p — good for sharing and storage savings",
			Category:     CategoryQuality,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          28,
			Resolution:   "1280x720",
			AudioCodec:   "aac",
			AudioBitrate: "96k",
			SpeedPreset:  "medium",
			Tags:         []string{"compact", "small", "720p", "sharing"},
			Icon:         "\U0001f4e6", // package
		},
		{
			Key:          "tiny",
			Name:         "Tiny",
			Description:  "Maximum compression at 480p — for when file size matters most",
			Category:     CategoryQuality,
			Codec:        "h264",
			Container:    "mp4",
			CRF:          32,
			Resolution:   "854x480",
			AudioCodec:   "aac",
			AudioBitrate: "64k",
			SpeedPreset:  "fast",
			Tags:         []string{"tiny", "smallest", "480p", "mobile", "low"},
			Icon:         "\U0001f41c", // ant
		},
	}
}
