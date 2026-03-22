package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/jparkerweb/shrinkray/internal/engine"
)

func probeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "probe <file>",
		Short: "Display detailed video metadata",
		Long:  "Probe a video file with FFprobe and display all metadata in a formatted output.",
		Args:  cobra.ExactArgs(1),
		RunE:  runProbe,
	}
}

func runProbe(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Validate file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Probe the file
	info, err := engine.Probe(context.Background(), filePath)
	if err != nil {
		return fmt.Errorf("failed to probe file: %w", err)
	}

	// Render output
	if flagNoColor {
		renderProbePlain(info)
	} else {
		renderProbeStyled(info)
	}

	return nil
}

func renderProbeStyled(info *engine.VideoInfo) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7B2FF7"))
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C6C8A")).
		Width(16)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8E8F0")).
		Bold(true)
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00F0FF")).
		Bold(true)
	hdrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAB00")).
		Bold(true)

	fmt.Println(titleStyle.Render("shrinkray probe"))
	fmt.Println()

	// File info
	fmt.Println(sectionStyle.Render("File"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Path:"), valueStyle.Render(info.Path))
	fmt.Printf("  %s %s\n", labelStyle.Render("Format:"), valueStyle.Render(info.Format))
	fmt.Printf("  %s %s\n", labelStyle.Render("Duration:"), valueStyle.Render(formatDuration(info.Duration)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Size:"), valueStyle.Render(formatBytes(info.Size)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Bitrate:"), valueStyle.Render(formatBitrateStr(info.Bitrate)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Streams:"), valueStyle.Render(fmt.Sprintf("%d", info.StreamCount)))
	fmt.Println()

	// Video info
	fmt.Println(sectionStyle.Render("Video"))
	fmt.Printf("  %s %s\n", labelStyle.Render("Codec:"), valueStyle.Render(fmt.Sprintf("%s (%s)", strings.ToUpper(info.Codec), info.CodecLong)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Resolution:"), valueStyle.Render(info.Resolution()))
	fmt.Printf("  %s %s\n", labelStyle.Render("Aspect Ratio:"), valueStyle.Render(info.AspectRatio()))
	fmt.Printf("  %s %s\n", labelStyle.Render("Framerate:"), valueStyle.Render(fmt.Sprintf("%.3f fps", info.Framerate)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Pixel Format:"), valueStyle.Render(info.PixelFormat))
	if info.ColorSpace != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Color Space:"), valueStyle.Render(info.ColorSpace))
	}
	if info.IsPortrait() {
		fmt.Printf("  %s %s\n", labelStyle.Render("Orientation:"), valueStyle.Render("Portrait"))
	} else {
		fmt.Printf("  %s %s\n", labelStyle.Render("Orientation:"), valueStyle.Render("Landscape"))
	}
	fmt.Println()

	// HDR
	if info.IsHDR {
		fmt.Printf("  %s %s\n", labelStyle.Render("HDR:"), hdrStyle.Render(info.HDRFormat))
	} else {
		fmt.Printf("  %s %s\n", labelStyle.Render("HDR:"), valueStyle.Render("SDR"))
	}
	fmt.Println()

	// Audio info
	if info.AudioCodec != "" {
		fmt.Println(sectionStyle.Render("Audio"))
		fmt.Printf("  %s %s\n", labelStyle.Render("Codec:"), valueStyle.Render(info.AudioCodec))
		fmt.Printf("  %s %s\n", labelStyle.Render("Channels:"), valueStyle.Render(fmt.Sprintf("%d", info.AudioChannels)))
		if info.AudioSampleRate > 0 {
			fmt.Printf("  %s %s\n", labelStyle.Render("Sample Rate:"), valueStyle.Render(fmt.Sprintf("%d Hz", info.AudioSampleRate)))
		}
		if info.AudioBitrate > 0 {
			fmt.Printf("  %s %s\n", labelStyle.Render("Bitrate:"), valueStyle.Render(formatBitrateStr(info.AudioBitrate)))
		}
		fmt.Println()
	}

	// Subtitles
	if info.SubtitleCount > 0 {
		fmt.Println(sectionStyle.Render("Subtitles"))
		fmt.Printf("  %s %s\n", labelStyle.Render("Tracks:"), valueStyle.Render(fmt.Sprintf("%d", info.SubtitleCount)))
		fmt.Println()
	}
}

func renderProbePlain(info *engine.VideoInfo) {
	fmt.Println("shrinkray probe")
	fmt.Println()

	fmt.Println("File")
	fmt.Printf("  %-16s %s\n", "Path:", info.Path)
	fmt.Printf("  %-16s %s\n", "Format:", info.Format)
	fmt.Printf("  %-16s %s\n", "Duration:", formatDuration(info.Duration))
	fmt.Printf("  %-16s %s\n", "Size:", formatBytes(info.Size))
	fmt.Printf("  %-16s %s\n", "Bitrate:", formatBitrateStr(info.Bitrate))
	fmt.Printf("  %-16s %d\n", "Streams:", info.StreamCount)
	fmt.Println()

	fmt.Println("Video")
	fmt.Printf("  %-16s %s (%s)\n", "Codec:", strings.ToUpper(info.Codec), info.CodecLong)
	fmt.Printf("  %-16s %s\n", "Resolution:", info.Resolution())
	fmt.Printf("  %-16s %s\n", "Aspect Ratio:", info.AspectRatio())
	fmt.Printf("  %-16s %.3f fps\n", "Framerate:", info.Framerate)
	fmt.Printf("  %-16s %s\n", "Pixel Format:", info.PixelFormat)
	if info.ColorSpace != "" {
		fmt.Printf("  %-16s %s\n", "Color Space:", info.ColorSpace)
	}
	if info.IsHDR {
		fmt.Printf("  %-16s %s\n", "HDR:", info.HDRFormat)
	} else {
		fmt.Printf("  %-16s SDR\n", "HDR:")
	}
	fmt.Println()

	if info.AudioCodec != "" {
		fmt.Println("Audio")
		fmt.Printf("  %-16s %s\n", "Codec:", info.AudioCodec)
		fmt.Printf("  %-16s %d\n", "Channels:", info.AudioChannels)
		if info.AudioSampleRate > 0 {
			fmt.Printf("  %-16s %d Hz\n", "Sample Rate:", info.AudioSampleRate)
		}
		if info.AudioBitrate > 0 {
			fmt.Printf("  %-16s %s\n", "Bitrate:", formatBitrateStr(info.AudioBitrate))
		}
		fmt.Println()
	}

	if info.SubtitleCount > 0 {
		fmt.Println("Subtitles")
		fmt.Printf("  %-16s %d\n", "Tracks:", info.SubtitleCount)
		fmt.Println()
	}
}

func formatBitrateStr(bps int64) string {
	if bps <= 0 {
		return "N/A"
	}
	kbps := float64(bps) / 1000
	if kbps >= 1000 {
		return fmt.Sprintf("%.1f Mbps", kbps/1000)
	}
	return fmt.Sprintf("%.0f kbps", kbps)
}
