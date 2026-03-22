package engine

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	ffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

// Probe extracts video metadata from the file at the given path.
func Probe(ctx context.Context, path string) (*VideoInfo, error) {
	// Validate file exists
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	data, err := ffprobe.ProbeURL(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	info := &VideoInfo{
		Path:        path,
		Size:        stat.Size(),
		StreamCount: len(data.Streams),
	}

	// Format info
	if data.Format != nil {
		info.Format = data.Format.FormatName
		info.Duration = data.Format.DurationSeconds
		if br, err := strconv.ParseInt(data.Format.BitRate, 10, 64); err == nil {
			info.Bitrate = br
		}
	}

	// Find video stream
	videoStream := data.FirstVideoStream()
	if videoStream != nil {
		info.Codec = videoStream.CodecName
		info.CodecLong = videoStream.CodecLongName
		info.Width = videoStream.Width
		info.Height = videoStream.Height
		info.PixelFormat = videoStream.PixFmt
		info.ColorSpace = videoStream.ColorSpace
		info.ColorTransfer = videoStream.ColorTransfer
		info.ColorPrimaries = videoStream.ColorPrimaries

		// Parse framerate from r_frame_rate (e.g., "30000/1001")
		info.Framerate = parseFramerate(videoStream.RFrameRate)

		// If overall bitrate is missing, try video stream bitrate
		if info.Bitrate == 0 {
			if br, err := strconv.ParseInt(videoStream.BitRate, 10, 64); err == nil {
				info.Bitrate = br
			}
		}
	}

	// Find audio stream
	audioStream := data.FirstAudioStream()
	if audioStream != nil {
		info.AudioCodec = audioStream.CodecName
		if br, err := strconv.ParseInt(audioStream.BitRate, 10, 64); err == nil {
			info.AudioBitrate = br
		}
		info.AudioChannels = audioStream.Channels
		if sr, err := strconv.Atoi(audioStream.SampleRate); err == nil {
			info.AudioSampleRate = sr
		}
	}

	// Count subtitle streams
	for _, s := range data.Streams {
		if s.CodecType == "subtitle" {
			info.SubtitleCount++
		}
	}

	// Detect HDR
	info.IsHDR, info.HDRFormat = DetectHDR(info)

	return info, nil
}

func parseFramerate(rate string) float64 {
	if rate == "" {
		return 0
	}
	parts := strings.SplitN(rate, "/", 2)
	if len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den != 0 {
			return math.Round(num/den*1000) / 1000
		}
	}
	if f, err := strconv.ParseFloat(rate, 64); err == nil {
		return f
	}
	return 0
}
