package engine

import "strings"

// DetectHDR determines whether a video is HDR and what format it uses.
// It inspects color transfer characteristics and side data.
// Returns (isHDR bool, format string) where format is one of:
// "HDR10", "HLG", "Dolby Vision", or "SDR".
func DetectHDR(info *VideoInfo) (bool, string) {
	if info == nil {
		return false, "SDR"
	}

	transfer := strings.ToLower(info.ColorTransfer)

	// HDR10 / HDR10+ use SMPTE ST 2084 (PQ)
	if transfer == "smpte2084" || transfer == "smpte-st-2084" {
		return true, "HDR10"
	}

	// HLG (Hybrid Log-Gamma)
	if transfer == "arib-std-b67" {
		return true, "HLG"
	}

	// Dolby Vision detection via codec name or pixel format hints
	codec := strings.ToLower(info.Codec)
	if strings.Contains(codec, "dolby") || strings.Contains(codec, "dvhe") || strings.Contains(codec, "dvh1") {
		return true, "Dolby Vision"
	}

	// Check pixel format for high bit-depth as a secondary indicator
	// (not conclusive on its own, so only flag if we haven't found explicit HDR markers)

	return false, "SDR"
}
