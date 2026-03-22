package engine

import (
	"github.com/jparkerweb/shrinkray/internal/presets"
)

// bppTable maps codec+CRF to approximate bits-per-pixel-per-frame values.
// These are empirical estimates for typical video content.
var bppTable = map[string]map[int]float64{
	"h264": {
		0:  0.80, // lossless
		16: 0.25,
		18: 0.20,
		20: 0.15,
		22: 0.11,
		23: 0.10,
		24: 0.085,
		26: 0.065,
		28: 0.05,
		30: 0.035,
		32: 0.025,
		34: 0.018,
		36: 0.012,
	},
	"h265": {
		0:  0.60,
		16: 0.18,
		18: 0.14,
		20: 0.10,
		22: 0.08,
		23: 0.07,
		24: 0.06,
		26: 0.045,
		28: 0.035,
		30: 0.025,
		32: 0.018,
		34: 0.012,
		36: 0.008,
	},
	"av1": {
		0:  0.50,
		18: 0.12,
		20: 0.09,
		23: 0.06,
		26: 0.04,
		28: 0.03,
		30: 0.02,
		32: 0.015,
		36: 0.008,
	},
	"vp9": {
		0:  0.55,
		18: 0.13,
		20: 0.10,
		23: 0.07,
		26: 0.045,
		28: 0.035,
		30: 0.025,
		32: 0.017,
		36: 0.009,
	},
}

// lookupBPP returns an interpolated bits-per-pixel value for a given codec and CRF.
func lookupBPP(codec string, crf int) float64 {
	table, ok := bppTable[codec]
	if !ok {
		// Fallback to h264 table
		table = bppTable["h264"]
	}

	// Exact match
	if bpp, ok := table[crf]; ok {
		return bpp
	}

	// Find surrounding values for linear interpolation
	var lowerCRF, upperCRF int
	var lowerBPP, upperBPP float64
	foundLower, foundUpper := false, false

	for c, b := range table {
		if c <= crf && (!foundLower || c > lowerCRF) {
			lowerCRF = c
			lowerBPP = b
			foundLower = true
		}
		if c >= crf && (!foundUpper || c < upperCRF) {
			upperCRF = c
			upperBPP = b
			foundUpper = true
		}
	}

	if !foundLower && !foundUpper {
		return 0.10 // fallback
	}
	if !foundLower {
		return upperBPP
	}
	if !foundUpper {
		return lowerBPP
	}
	if lowerCRF == upperCRF {
		return lowerBPP
	}

	// Linear interpolation
	ratio := float64(crf-lowerCRF) / float64(upperCRF-lowerCRF)
	return lowerBPP + ratio*(upperBPP-lowerBPP)
}

// EstimateSize estimates the output file size in bytes for a given video and preset.
// For target-size presets, it returns TargetSizeMB * 1024 * 1024.
// For CRF-based presets, it calculates based on bits-per-pixel lookup.
func EstimateSize(info *VideoInfo, preset presets.Preset) int64 {
	if info == nil {
		return 0
	}

	// Target-size presets return the target directly
	if preset.TargetSizeMB > 0 {
		return int64(preset.TargetSizeMB * 1024 * 1024)
	}

	// Determine effective resolution
	width := info.Width
	height := info.Height
	if preset.Resolution != "" {
		pw, ph := parseResolution(preset.Resolution)
		if pw > 0 && ph > 0 {
			// Scale down only
			if pw < width || ph < height {
				// Maintain aspect ratio — scale to fit within preset resolution
				scaleW := float64(pw) / float64(width)
				scaleH := float64(ph) / float64(height)
				scale := scaleW
				if scaleH < scaleW {
					scale = scaleH
				}
				width = int(float64(width) * scale)
				height = int(float64(height) * scale)
				// Ensure divisible by 2
				width = (width / 2) * 2
				height = (height / 2) * 2
			}
		}
	}

	if width <= 0 || height <= 0 {
		return 0
	}

	fps := info.Framerate
	if fps <= 0 {
		fps = 30 // reasonable default
	}
	if preset.MaxFPS > 0 && fps > float64(preset.MaxFPS) {
		fps = float64(preset.MaxFPS)
	}

	duration := info.Duration
	if duration <= 0 {
		return 0
	}

	// Lookup bits per pixel
	bpp := lookupBPP(preset.Codec, preset.CRF)

	// Estimated video bitrate (bits per second)
	videoBitrate := bpp * float64(width) * float64(height) * fps

	// Estimated audio size
	var audioBitsPerSecond float64
	switch preset.AudioCodec {
	case "copy":
		audioBitsPerSecond = float64(info.AudioBitrate)
	case "none":
		audioBitsPerSecond = 0
	default:
		audioBitsPerSecond = parseAudioBitrate(preset.AudioBitrate)
	}

	// Total estimated bytes
	totalBits := (videoBitrate + audioBitsPerSecond) * duration
	estimatedBytes := int64(totalBits / 8)

	if estimatedBytes < 1024 {
		estimatedBytes = 1024
	}

	return estimatedBytes
}

// parseAudioBitrate converts a string like "128k" or "192k" to bits per second.
func parseAudioBitrate(s string) float64 {
	if s == "" {
		return 128000 // default
	}

	var val float64
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		val = val*10 + float64(s[i]-'0')
		i++
	}

	if i == 0 {
		return 128000 // couldn't parse, use default
	}

	suffix := s[i:]
	switch suffix {
	case "k", "K":
		return val * 1000
	case "m", "M":
		return val * 1000000
	default:
		return val * 1000 // assume kbps
	}
}
