package cli

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/jparkerweb/shrinkray/internal/engine"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Short commit (first 7 chars)
			shortCommit := commit
			if len(shortCommit) > 7 {
				shortCommit = shortCommit[:7]
			}

			fmt.Printf("shrinkray version %s (commit: %s, built: %s)\n", version, shortCommit, date)
			fmt.Printf("  go:      %s\n", runtime.Version())
			fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

			// Try to detect FFmpeg version
			ffmpegInfo, err := engine.DetectFFmpeg()
			if err != nil {
				slog.Debug("ffmpeg not found for version display", "error", err)
				fmt.Printf("  ffmpeg:  not found\n")
			} else {
				fmt.Printf("  ffmpeg:  %s\n", ffmpegInfo.Version)
			}

			// Try to detect FFprobe version
			ffprobeInfo, err := engine.DetectFFprobe()
			if err != nil {
				slog.Debug("ffprobe not found for version display", "error", err)
				fmt.Printf("  ffprobe: not found\n")
			} else {
				fmt.Printf("  ffprobe: %s\n", ffprobeInfo.Version)
			}

			return nil
		},
	}
}
