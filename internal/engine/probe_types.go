package engine

import "fmt"

// VideoInfo holds all metadata extracted from a video file via FFprobe.
type VideoInfo struct {
	Path            string  `json:"path"`
	Format          string  `json:"format"`
	Duration        float64 `json:"duration"`         // seconds
	Size            int64   `json:"size"`              // bytes
	Codec           string  `json:"codec"`             // e.g. "h264", "hevc"
	CodecLong       string  `json:"codecLong"`         // e.g. "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10"
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Framerate       float64 `json:"framerate"`
	Bitrate         int64   `json:"bitrate"`           // bits per second
	PixelFormat     string  `json:"pixelFormat"`
	ColorSpace      string  `json:"colorSpace"`
	ColorTransfer   string  `json:"colorTransfer"`
	ColorPrimaries  string  `json:"colorPrimaries"`
	IsHDR           bool    `json:"isHDR"`
	HDRFormat       string  `json:"hdrFormat"`         // "HDR10", "HLG", "Dolby Vision", "SDR"
	AudioCodec      string  `json:"audioCodec"`
	AudioBitrate    int64   `json:"audioBitrate"`      // bits per second
	AudioChannels   int     `json:"audioChannels"`
	AudioSampleRate int     `json:"audioSampleRate"`
	StreamCount     int     `json:"streamCount"`
	SubtitleCount   int     `json:"subtitleCount"`
}

// Resolution returns the video resolution as "WxH" (e.g., "1920x1080").
func (v *VideoInfo) Resolution() string {
	if v.Width == 0 && v.Height == 0 {
		return "unknown"
	}
	return fmt.Sprintf("%dx%d", v.Width, v.Height)
}

// AspectRatio returns a simplified aspect ratio string (e.g., "16:9").
func (v *VideoInfo) AspectRatio() string {
	if v.Width == 0 || v.Height == 0 {
		return "unknown"
	}
	g := gcd(v.Width, v.Height)
	return fmt.Sprintf("%d:%d", v.Width/g, v.Height/g)
}

// IsPortrait returns true if the video is taller than it is wide.
func (v *VideoInfo) IsPortrait() bool {
	return v.Height > v.Width
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
