package presets

// purposePresets returns the 5 purpose-driven presets.
func purposePresets() []Preset {
	return []Preset{
		{
			Key:          "web",
			Name:         "Web Delivery",
			Description:  "Optimized for web streaming and downloads — fast start, wide compatibility",
			Category:     CategoryPurpose,
			Codec:        "h264",
			Container:    "mp4",
			CRF:          23,
			Resolution:   "1920x1080",
			AudioCodec:   "aac",
			AudioBitrate: "128k",
			SpeedPreset:  "medium",
			Tags:         []string{"web", "streaming", "html5", "browser", "online"},
			Icon:         "\U0001f310", // globe
		},
		{
			Key:          "email",
			Name:         "Email Friendly",
			Description:  "Small enough for email attachments — targets ~20MB output",
			Category:     CategoryPurpose,
			Codec:        "h264",
			Container:    "mp4",
			CRF:          28,
			Resolution:   "1280x720",
			AudioCodec:   "aac",
			AudioBitrate: "96k",
			SpeedPreset:  "medium",
			TargetSizeMB: 20,
			TwoPass:      true,
			Tags:         []string{"email", "attachment", "small", "20mb"},
			Icon:         "\U0001f4e7", // envelope
		},
		{
			Key:          "archive",
			Name:         "Archive",
			Description:  "Maximum quality retention with H.265 — preserve your originals",
			Category:     CategoryPurpose,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          18,
			AudioCodec:   "copy",
			AudioBitrate: "",
			SpeedPreset:  "slow",
			Tags:         []string{"archive", "preserve", "quality", "backup", "lossless-ish"},
			Icon:         "\U0001f4be", // floppy disk
		},
		{
			Key:          "slideshow",
			Name:         "Slideshow / Screencast",
			Description:  "Low-motion content like photo slideshows and screen recordings",
			Category:     CategoryPurpose,
			Codec:        "h264",
			Container:    "mp4",
			CRF:          32,
			Resolution:   "1280x720",
			AudioCodec:   "aac",
			AudioBitrate: "64k",
			SpeedPreset:  "fast",
			Tags:         []string{"slideshow", "screencast", "presentation", "low-motion", "static"},
			Icon:         "\U0001f4fd\ufe0f", // film projector
		},
		{
			Key:          "4k-to-1080",
			Name:         "4K to 1080p",
			Description:  "Downscale 4K footage to 1080p with excellent quality",
			Category:     CategoryPurpose,
			Codec:        "h265",
			Container:    "mp4",
			CRF:          20,
			Resolution:   "1920x1080",
			AudioCodec:   "aac",
			AudioBitrate: "192k",
			SpeedPreset:  "medium",
			Tags:         []string{"4k", "downscale", "1080p", "uhd", "resize"},
			Icon:         "\U0001f4f0", // newspaper (repurposed as resize icon)
		},
	}
}
