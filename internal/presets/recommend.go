package presets

import (
	"sort"
)

// VideoMetadata holds the video properties needed for recommendations.
// This is a lightweight struct that avoids importing the engine package to prevent
// circular dependencies (engine already imports presets).
type VideoMetadata struct {
	Width     int
	Height    int
	Framerate float64
	Bitrate   int64   // bits per second
	Duration  float64 // seconds
	Size      int64   // bytes
	Codec     string  // e.g., "h264", "hevc", "av1"
}

// IsPortrait returns true if the video is taller than it is wide.
func (v *VideoMetadata) IsPortrait() bool {
	return v.Height > v.Width
}

// Recommendation represents a preset recommendation with a score and reason.
type Recommendation struct {
	Preset Preset
	Score  int      // 0-100, higher is better match
	Reason string
	Tags   []string // e.g., "best-quality", "smallest", "fastest"
}

// Recommend analyzes a video's metadata and returns the top preset recommendations.
// Uses rule-based heuristics to score presets based on source characteristics.
func Recommend(info *VideoMetadata) []Recommendation {
	if info == nil {
		return nil
	}

	var recommendations []Recommendation

	bitrateKbps := float64(info.Bitrate) / 1000
	is4K := info.Width >= 3840 || info.Height >= 2160
	isHighBitrate := bitrateKbps > 10000 // > 10 Mbps
	isShort := info.Duration > 0 && info.Duration < 120
	isLargeFile := info.Size > 500*1024*1024 // > 500 MB
	isAlreadyCompressed := isWellCompressed(info)
	isPortrait := info.IsPortrait()

	// Score each preset based on rules
	allPresets := All()
	for _, p := range allPresets {
		score := 0
		var reason string
		var tags []string

		switch p.Key {
		case "balanced":
			score = 70 // good default
			reason = "Great all-around choice for most videos"
			tags = append(tags, "recommended")
			if isHighBitrate {
				score += 15
				reason = "High bitrate source — good compression potential"
				tags = append(tags, "high-savings")
			}

		case "compact":
			score = 40
			reason = "Smaller file with 720p cap"
			if isHighBitrate {
				score += 20
				reason = "High bitrate source — significant size reduction possible"
				tags = append(tags, "smallest")
			}
			if isLargeFile {
				score += 15
				reason = "Large file — compact preset can save substantial space"
				tags = append(tags, "high-savings")
			}

		case "4k-to-1080":
			if is4K {
				score = 90
				reason = "4K source detected — downscale to 1080p with excellent quality"
				tags = append(tags, "best-match")
			}

		case "high":
			score = 55
			reason = "High quality with good compression"
			if !isHighBitrate {
				score += 10
				reason = "Already moderate bitrate — high quality preserves detail"
				tags = append(tags, "best-quality")
			}

		case "web":
			score = 50
			reason = "Optimized for web delivery and streaming"
			if isHighBitrate {
				score += 10
			}

		case "email":
			score = 30
			reason = "Small enough for email attachments (~20MB)"
			if isLargeFile {
				score += 25
				reason = "Large file needs aggressive compression for email"
			}

		case "discord":
			score = 25
			reason = "Fits Discord's 10MB free-tier limit"
			if isShort {
				score += 30
				reason = "Short clip — perfect for Discord sharing"
				tags = append(tags, "best-match")
			}

		case "discord-nitro":
			score = 20
			reason = "Fits Discord Nitro's 50MB limit at 1080p"
			if isShort {
				score += 20
				reason = "Short clip fits well in 50MB"
			}

		case "whatsapp":
			score = 20
			reason = "Fits WhatsApp's 16MB limit"
			if isShort {
				score += 25
				reason = "Short clip — perfect for WhatsApp sharing"
			}

		case "twitter":
			if isShort {
				score = 40
				reason = "Short clip — fits Twitter's 140s recommendation"
			} else {
				score = 20
				reason = "Optimized for Twitter upload"
			}

		case "instagram":
			if isShort && info.Duration <= 60 {
				score = 45
				reason = "Within Instagram's 60s recommendation"
			} else {
				score = 15
				reason = "Optimized for Instagram upload"
			}
			if isPortrait {
				score += 10
				reason += " (portrait content detected)"
			}

		case "tiktok":
			if isShort {
				score = 35
				reason = "Short clip — good for TikTok"
			} else {
				score = 15
				reason = "Optimized for TikTok upload"
			}
			if isPortrait {
				score += 15
				reason = "Portrait content detected — ideal for TikTok"
				tags = append(tags, "portrait")
			}

		case "youtube":
			score = 40
			reason = "Upload-optimized for YouTube"
			if is4K {
				score += 10
				reason = "4K source — YouTube upload preset preserves resolution"
			}

		case "archive":
			score = 35
			reason = "Maximum quality retention for archival"
			tags = append(tags, "best-quality")

		case "ultra":
			score = 30
			reason = "Near-lossless quality"
			tags = append(tags, "best-quality")

		case "lossless":
			score = 10
			reason = "No quality loss — large output"

		case "tiny":
			score = 20
			reason = "Maximum compression at 480p"
			if isLargeFile {
				score += 10
				tags = append(tags, "smallest")
			}

		case "slideshow":
			score = 15
			reason = "Optimized for low-motion content"
		}

		// Penalty if already well-compressed
		if isAlreadyCompressed && score > 0 {
			score -= 20
			if score < 5 {
				score = 5
			}
			reason += " (source already well-compressed — savings may be minimal)"
		}

		if score > 0 {
			recommendations = append(recommendations, Recommendation{
				Preset: p,
				Score:  score,
				Reason: reason,
				Tags:   tags,
			})
		}
	}

	// Sort by score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Return top 5
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations
}

// isWellCompressed returns true if the video appears to be already efficiently compressed.
func isWellCompressed(info *VideoMetadata) bool {
	if info.Width == 0 || info.Height == 0 || info.Bitrate == 0 || info.Framerate == 0 {
		return false
	}

	pixels := float64(info.Width * info.Height)
	bpp := float64(info.Bitrate) / (pixels * info.Framerate)

	// Check if already using an efficient codec with low bpp
	isEfficientCodec := info.Codec == "hevc" || info.Codec == "h265" || info.Codec == "av1" || info.Codec == "vp9"
	return isEfficientCodec && bpp < 0.05
}
