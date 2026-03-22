package engine

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

// MetadataMode controls metadata handling for the output file.
type MetadataMode struct {
	StripAll bool   // add -map_metadata -1
	Title    string // set -metadata title="value"
}

// MetadataFromFlags builds MetadataMode from CLI flag values.
func MetadataFromFlags(strip, keep bool, title string) MetadataMode {
	m := MetadataMode{}
	if strip {
		m.StripAll = true
	}
	if title != "" {
		m.Title = title
	}
	return m
}

// EncodeOptions holds all parameters needed to build an FFmpeg command.
type EncodeOptions struct {
	Input              string
	Output             string
	Preset             presets.Preset
	HWEncoder          string // empty = software encoding
	CRFOverride        *int
	ResolutionOverride string
	Pass               int    // 0 = single-pass, 1 or 2 for two-pass
	PassLogFile        string
	VideoBitrate       int64 // target video bitrate in bps (used for target-size two-pass)
	ExtraArgs          []string

	// Source video info for portrait/social handling
	SourceInfo *VideoInfo

	// Metadata handling
	MetadataMode MetadataMode
}

// BuildArgs constructs the FFmpeg argument array from the given EncodeOptions.
func BuildArgs(opts EncodeOptions) []string {
	var args []string

	// Hardware acceleration device flags (before input)
	if opts.HWEncoder != "" {
		args = append(args, hwAccelFlags(opts.HWEncoder)...)
	}

	// Input
	args = append(args, "-i", opts.Input)

	// Video codec
	videoCodecLib := codecToLib(opts.Preset.Codec)
	if opts.HWEncoder != "" {
		videoCodecLib = opts.HWEncoder
	}
	args = append(args, "-c:v", videoCodecLib)

	// CRF / bitrate / quality handling
	crf := opts.Preset.CRF
	if opts.CRFOverride != nil {
		crf = *opts.CRFOverride
	}

	// AV1 CRF mapping: software AV1 uses 0-63 range
	if opts.Preset.Codec == "av1" && opts.HWEncoder == "" {
		crf = mapAV1CRF(crf)
	}

	if opts.HWEncoder != "" {
		// Hardware encoder quality mapping
		args = append(args, buildHWQualityArgs(opts, crf)...)
	} else {
		// Software encoding quality
		args = append(args, buildSoftwareQualityArgs(opts, crf)...)
	}

	// Speed preset (x264/x265 only, software encoding only)
	if opts.HWEncoder == "" {
		speedPreset := opts.Preset.SpeedPreset
		if speedPreset == "" {
			speedPreset = "medium"
		}
		if opts.Preset.Codec == "h264" || opts.Preset.Codec == "h265" {
			args = append(args, "-preset", speedPreset)
		}
	} else if isNVENC(opts.HWEncoder) {
		// NVENC has its own preset system
		args = append(args, "-preset", "p4")
	} else if isQSV(opts.HWEncoder) {
		args = append(args, "-preset", "medium")
	}

	// SVT-AV1 specific params (software only)
	if opts.Preset.Codec == "av1" && opts.HWEncoder == "" {
		args = append(args, "-svtav1-params", "tune=0")
	}

	// VP9 specific: row-based multithreading
	if opts.Preset.Codec == "vp9" && opts.HWEncoder == "" {
		args = append(args, "-row-mt", "1")
	}

	// Max bitrate
	if opts.Preset.MaxBitrate != "" {
		args = append(args, "-maxrate", opts.Preset.MaxBitrate, "-bufsize", opts.Preset.MaxBitrate)
	}

	// Resolution scaling with portrait/social handling
	args = append(args, buildScaleArgs(opts)...)

	// MaxFPS
	if opts.Preset.MaxFPS > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", opts.Preset.MaxFPS))
	}

	// Audio settings
	args = append(args, buildAudioArgs(opts)...)

	// Two-pass handling
	if opts.Pass > 0 {
		// Check if HW encoder supports two-pass
		if opts.HWEncoder != "" && !SupportsHWTwoPass(opts.HWEncoder) {
			slog.Warn("HW encoder does not support two-pass, using single-pass CQ mode",
				"encoder", opts.HWEncoder)
			// Don't add pass flags — fall through to single-pass
		} else {
			args = append(args, "-pass", fmt.Sprintf("%d", opts.Pass))
			if opts.PassLogFile != "" {
				args = append(args, "-passlogfile", opts.PassLogFile)
			}
			if opts.Pass == 1 {
				// First pass: skip output, use null muxer
				args = append(args, "-f", "null")
			}
		}
	}

	// Container-specific flags
	if opts.Pass != 1 { // Don't add these for null output
		container := opts.Preset.Container
		if container == "" {
			container = "mp4"
		}
		switch container {
		case "mp4":
			args = append(args, "-movflags", "+faststart")
		case "webm":
			// webm-specific flags if needed
		}
	}

	// Metadata handling
	if opts.MetadataMode.StripAll {
		args = append(args, "-map_metadata", "-1")
	} else {
		args = append(args, "-map_metadata", "0")
	}
	if opts.MetadataMode.Title != "" {
		args = append(args, "-metadata", fmt.Sprintf("title=%s", opts.MetadataMode.Title))
	}

	// Extra args from preset
	args = append(args, opts.Preset.ExtraArgs...)

	// Extra args from options
	args = append(args, opts.ExtraArgs...)

	// Progress output for parsing
	args = append(args, "-progress", "pipe:1", "-stats_period", "0.5")

	// Overwrite output
	args = append(args, "-y")

	// Output path
	if opts.Pass == 1 {
		// First pass writes to null
		if isWindows() {
			args = append(args, "NUL")
		} else {
			args = append(args, "/dev/null")
		}
	} else {
		args = append(args, opts.Output)
	}

	return args
}

// hwAccelFlags returns platform-specific hardware acceleration device flags.
func hwAccelFlags(encoder string) []string {
	switch {
	case isNVENC(encoder):
		return []string{"-hwaccel", "cuda"}
	case isVideoToolbox(encoder):
		return []string{"-hwaccel", "videotoolbox"}
	case isQSV(encoder):
		return []string{"-hwaccel", "qsv"}
	case isAMF(encoder):
		// AMF doesn't need a -hwaccel flag typically
		return nil
	default:
		return nil
	}
}

// buildHWQualityArgs builds quality-related FFmpeg args for hardware encoders.
func buildHWQualityArgs(opts EncodeOptions, crf int) []string {
	var args []string

	// For target-size two-pass encoding, use bitrate mode
	if opts.VideoBitrate > 0 && opts.Pass > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%d", opts.VideoBitrate))
		return args
	}

	// Map software CRF to hardware quality params
	qualityMap := MapQuality(crf, opts.HWEncoder)
	if qualityMap != nil {
		for k, v := range qualityMap {
			args = append(args, k, v)
		}
	} else {
		// Fallback: use CRF directly (some encoders might support it)
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
	}

	return args
}

// buildSoftwareQualityArgs builds quality-related FFmpeg args for software encoders.
func buildSoftwareQualityArgs(opts EncodeOptions, crf int) []string {
	var args []string

	// For target-size two-pass encoding, use bitrate mode instead of CRF
	if opts.VideoBitrate > 0 && opts.Pass > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%d", opts.VideoBitrate))
		return args
	}

	if opts.Pass != 2 || opts.Preset.MaxBitrate == "" {
		// VP9 and AV1 require -b:v 0 for CRF-only mode
		if opts.Preset.Codec == "vp9" || opts.Preset.Codec == "av1" {
			args = append(args, "-crf", fmt.Sprintf("%d", crf), "-b:v", "0")
		} else {
			args = append(args, "-crf", fmt.Sprintf("%d", crf))
		}
	}

	return args
}

// mapAV1CRF converts h264/h265-style CRF values (0-51) to SVT-AV1 range (0-63).
// Known mapping: 18->24, 23->32, 28->38
func mapAV1CRF(crf int) int {
	return interpolateQuality(crf, []qualityPoint{
		{18, 24},
		{23, 32},
		{28, 38},
	})
}

// buildScaleArgs builds the video filter args for resolution scaling,
// including portrait/social media handling.
func buildScaleArgs(opts EncodeOptions) []string {
	resolution := opts.Preset.Resolution
	if opts.ResolutionOverride != "" {
		resolution = opts.ResolutionOverride
	}
	if resolution == "" {
		return nil
	}

	w, h := parseResolution(resolution)
	if w <= 0 || h <= 0 {
		return nil
	}

	// Portrait/social media handling
	if opts.SourceInfo != nil && opts.SourceInfo.IsPortrait() && isSocialPreset(opts.Preset) {
		return buildPortraitScaleArgs(opts, w, h)
	}

	// Standard landscape scaling: scale down only, preserve aspect ratio, ensure divisible by 2
	filter := fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", w, h)
	return []string{"-vf", filter}
}

// buildPortraitScaleArgs builds scale filter for portrait videos targeting social platforms.
func buildPortraitScaleArgs(opts EncodeOptions, w, h int) []string {
	// Swap width/height for portrait: 1920x1080 -> 1080x1920
	pw, ph := h, w
	if pw > ph {
		pw, ph = ph, pw
	}
	// For Instagram/TikTok, add padding to ensure 9:16 aspect ratio
	if isInstagramOrTikTok(opts.Preset) {
		filter := fmt.Sprintf(
			"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black",
			pw, ph, pw, ph,
		)
		return []string{"-vf", filter}
	}

	// Standard portrait swap: just swap dimensions
	filter := fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", pw, ph)
	return []string{"-vf", filter}
}

// isSocialPreset returns true if the preset targets a social media platform.
func isSocialPreset(p presets.Preset) bool {
	if p.Category == presets.CategoryPlatform {
		return true
	}
	for _, tag := range p.Tags {
		if strings.ToLower(tag) == "social" {
			return true
		}
	}
	return false
}

// isInstagramOrTikTok returns true for Instagram or TikTok presets.
func isInstagramOrTikTok(p presets.Preset) bool {
	key := strings.ToLower(p.Key)
	return key == "instagram" || key == "tiktok"
}

// buildAudioArgs builds audio-related FFmpeg arguments.
func buildAudioArgs(opts EncodeOptions) []string {
	var args []string

	audioCodec := opts.Preset.AudioCodec

	// VP9 requires opus audio for webm container
	if opts.Preset.Codec == "vp9" && audioCodec != "copy" && audioCodec != "none" {
		audioCodec = "opus"
	}

	switch audioCodec {
	case "copy":
		args = append(args, "-c:a", "copy")
	case "none":
		args = append(args, "-an")
	case "opus":
		args = append(args, "-c:a", "libopus")
		if opts.Preset.AudioBitrate != "" {
			args = append(args, "-b:a", opts.Preset.AudioBitrate)
		}
	default: // "aac" or empty
		args = append(args, "-c:a", "aac")
		if opts.Preset.AudioBitrate != "" {
			args = append(args, "-b:a", opts.Preset.AudioBitrate)
		}
	}

	if opts.Preset.AudioChannels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", opts.Preset.AudioChannels))
	}

	return args
}

func codecToLib(codec string) string {
	switch strings.ToLower(codec) {
	case "h264":
		return "libx264"
	case "h265", "hevc":
		return "libx265"
	case "av1":
		return "libsvtav1"
	case "vp9":
		return "libvpx-vp9"
	default:
		return "libx265"
	}
}

func parseResolution(res string) (int, int) {
	res = strings.ToLower(strings.TrimSpace(res))
	var w, h int
	if n, _ := fmt.Sscanf(res, "%dx%d", &w, &h); n == 2 {
		return w, h
	}
	return 0, 0
}

func isWindows() bool {
	// We check at runtime rather than build tag for simplicity
	return strings.Contains(strings.ToLower(runtimeGOOS()), "windows")
}
