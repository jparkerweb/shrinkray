package engine

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sort"
	"sync"
)

// HWEncoder describes a hardware-accelerated video encoder.
type HWEncoder struct {
	Name        string   // short name: "nvenc", "videotoolbox", "qsv", "amf", "vaapi"
	DisplayName string   // human-friendly: "NVIDIA NVENC"
	GPU         string   // GPU name from detection (if available)
	Codecs      []string // supported codecs: "h264", "h265", "av1"
	Available   bool     // true if test-encode succeeded
}

// hwEncoderCandidate defines a candidate encoder to probe.
type hwEncoderCandidate struct {
	Name        string
	DisplayName string
	// map of codec -> FFmpeg encoder name
	Encoders map[string]string
}

// platformCandidates returns the ordered encoder candidates for the current platform.
func platformCandidates() []hwEncoderCandidate {
	switch runtime.GOOS {
	case "darwin":
		return []hwEncoderCandidate{
			{
				Name:        "videotoolbox",
				DisplayName: "Apple VideoToolbox",
				Encoders: map[string]string{
					"h264": "h264_videotoolbox",
					"h265": "hevc_videotoolbox",
				},
			},
		}
	default: // windows, linux
		return []hwEncoderCandidate{
			{
				Name:        "nvenc",
				DisplayName: "NVIDIA NVENC",
				Encoders: map[string]string{
					"h264": "h264_nvenc",
					"h265": "hevc_nvenc",
					"av1":  "av1_nvenc",
				},
			},
			{
				Name:        "qsv",
				DisplayName: "Intel Quick Sync",
				Encoders: map[string]string{
					"h264": "h264_qsv",
					"h265": "hevc_qsv",
					"av1":  "av1_qsv",
				},
			},
			{
				Name:        "amf",
				DisplayName: "AMD AMF",
				Encoders: map[string]string{
					"h264": "h264_amf",
					"h265": "hevc_amf",
					"av1":  "av1_amf",
				},
			},
			{
				Name:        "vaapi",
				DisplayName: "VAAPI",
				Encoders: map[string]string{
					"h264": "h264_vaapi",
					"h265": "hevc_vaapi",
					"av1":  "av1_vaapi",
				},
			},
		}
	}
}

// hwDetectCache stores session-level detection results.
var (
	hwDetectOnce   sync.Once
	hwDetectResult []HWEncoder
	hwDetectErr    error
)

// DetectHardware probes for available hardware encoders on the current platform.
// Results are cached for the session — subsequent calls return the cached result.
func DetectHardware(ctx context.Context) ([]HWEncoder, error) {
	hwDetectOnce.Do(func() {
		hwDetectResult, hwDetectErr = detectHardwareImpl(ctx)
	})
	return hwDetectResult, hwDetectErr
}

// ResetHWDetectCache clears the cached detection results. Intended for testing.
func ResetHWDetectCache() {
	hwDetectOnce = sync.Once{}
	hwDetectResult = nil
	hwDetectErr = nil
}

func detectHardwareImpl(ctx context.Context) ([]HWEncoder, error) {
	candidates := platformCandidates()
	var encoders []HWEncoder

	for _, cand := range candidates {
		enc := HWEncoder{
			Name:        cand.Name,
			DisplayName: cand.DisplayName,
			Available:   false,
		}

		var codecs []string
		for codec, ffmpegEncoder := range cand.Encoders {
			ok, err := ProbeEncoder(ctx, ffmpegEncoder, codec)
			if err != nil {
				slog.Debug("encoder probe error",
					"encoder", ffmpegEncoder,
					"codec", codec,
					"error", err,
				)
				continue
			}
			if ok {
				codecs = append(codecs, codec)
				enc.Available = true
			}
		}
		// Sort codecs for deterministic output
		sort.Strings(codecs)
		enc.Codecs = codecs

		encoders = append(encoders, enc)
	}

	// Sort: available encoders first, maintaining platform priority otherwise
	sort.SliceStable(encoders, func(i, j int) bool {
		if encoders[i].Available != encoders[j].Available {
			return encoders[i].Available
		}
		return false
	})

	return encoders, nil
}

// HWEncoderName returns the FFmpeg encoder name for a given HW accelerator and codec.
// For example, ("nvenc", "h265") returns "hevc_nvenc".
func HWEncoderName(hwName string, codec string) string {
	candidates := platformCandidates()
	for _, cand := range candidates {
		if cand.Name == hwName {
			if enc, ok := cand.Encoders[codec]; ok {
				return enc
			}
		}
	}
	return ""
}

// MapQuality converts a software CRF value to hardware encoder quality parameters.
// It returns a map of FFmpeg flag key-value pairs (e.g., {"-cq": "25"} for NVENC).
func MapQuality(softwareCRF int, encoder string) map[string]string {
	// Interpolate between known mapping points
	switch {
	case isNVENC(encoder):
		return mapNVENCQuality(softwareCRF)
	case isVideoToolbox(encoder):
		return mapVTBQuality(softwareCRF)
	case isQSV(encoder):
		return mapQSVQuality(softwareCRF)
	case isAMF(encoder):
		return mapAMFQuality(softwareCRF)
	default:
		return nil
	}
}

func isNVENC(encoder string) bool {
	return containsAny(encoder, "nvenc")
}

func isVideoToolbox(encoder string) bool {
	return containsAny(encoder, "videotoolbox")
}

func isQSV(encoder string) bool {
	return containsAny(encoder, "qsv")
}

func isAMF(encoder string) bool {
	return containsAny(encoder, "amf")
}

func containsAny(s string, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// mapNVENCQuality maps software CRF to NVENC -cq values.
// Known points: CRF 18->CQ 20, CRF 23->CQ 25, CRF 28->CQ 30
func mapNVENCQuality(crf int) map[string]string {
	cq := interpolateQuality(crf, []qualityPoint{
		{18, 20},
		{23, 25},
		{28, 30},
	})
	return map[string]string{
		"-cq": intToStr(cq),
	}
}

// mapVTBQuality maps software CRF to VideoToolbox -q:v values.
// Known points: CRF 18->45, CRF 23->55, CRF 28->65
func mapVTBQuality(crf int) map[string]string {
	qv := interpolateQuality(crf, []qualityPoint{
		{18, 45},
		{23, 55},
		{28, 65},
	})
	return map[string]string{
		"-q:v": intToStr(qv),
	}
}

// mapQSVQuality maps software CRF to QSV -global_quality values.
// Known points: CRF 18->22, CRF 23->27, CRF 28->32
func mapQSVQuality(crf int) map[string]string {
	gq := interpolateQuality(crf, []qualityPoint{
		{18, 22},
		{23, 27},
		{28, 32},
	})
	return map[string]string{
		"-global_quality": intToStr(gq),
	}
}

// mapAMFQuality maps software CRF to AMF constant QP values.
// Known points: CRF 18->20, CRF 23->25, CRF 28->30
func mapAMFQuality(crf int) map[string]string {
	qp := interpolateQuality(crf, []qualityPoint{
		{18, 20},
		{23, 25},
		{28, 30},
	})
	return map[string]string{
		"-rc":   "cqp",
		"-qp_i": intToStr(qp),
		"-qp_p": intToStr(qp),
	}
}

type qualityPoint struct {
	softwareCRF int
	hwValue     int
}

// interpolateQuality linearly interpolates between known quality mapping points.
func interpolateQuality(crf int, points []qualityPoint) int {
	if len(points) == 0 {
		return crf
	}

	// Below the first point — extrapolate
	if crf <= points[0].softwareCRF {
		if len(points) < 2 {
			return points[0].hwValue
		}
		slope := float64(points[1].hwValue-points[0].hwValue) / float64(points[1].softwareCRF-points[0].softwareCRF)
		val := float64(points[0].hwValue) + slope*float64(crf-points[0].softwareCRF)
		return int(val + 0.5)
	}

	// Above the last point — extrapolate
	if crf >= points[len(points)-1].softwareCRF {
		if len(points) < 2 {
			return points[len(points)-1].hwValue
		}
		last := len(points) - 1
		slope := float64(points[last].hwValue-points[last-1].hwValue) / float64(points[last].softwareCRF-points[last-1].softwareCRF)
		val := float64(points[last].hwValue) + slope*float64(crf-points[last].softwareCRF)
		return int(val + 0.5)
	}

	// Between two points — interpolate
	for i := 0; i < len(points)-1; i++ {
		if crf >= points[i].softwareCRF && crf <= points[i+1].softwareCRF {
			t := float64(crf-points[i].softwareCRF) / float64(points[i+1].softwareCRF-points[i].softwareCRF)
			val := float64(points[i].hwValue) + t*float64(points[i+1].hwValue-points[i].hwValue)
			return int(val + 0.5)
		}
	}

	return crf
}

func intToStr(v int) string {
	return fmt.Sprintf("%d", v)
}

// SupportsHWTwoPass returns whether a hardware encoder supports two-pass encoding.
// NVENC does not support two-pass well — returns false.
func SupportsHWTwoPass(encoder string) bool {
	if isNVENC(encoder) {
		return false
	}
	if isAMF(encoder) {
		return false
	}
	return true
}
