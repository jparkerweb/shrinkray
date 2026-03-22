package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Build info — set via SetBuildInfo() from main.go (injected via ldflags).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetBuildInfo sets the build information from main.go ldflags.
func SetBuildInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// Global flags
var (
	flagConfig   string
	flagLogLevel string
	flagNoColor  bool
)

// Execute is the main entry point for the CLI. Called from cmd/shrinkray/main.go.
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "shrinkray",
		Short: "Less bytes, same vibes.",
		Long:  "shrinkray is a cross-platform video compression tool powered by FFmpeg.",
		// Default command is "run" — the encode workflow
		RunE: runCmd,
	}

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable color output")

	// Register encoding flags on the root command (since run is the default)
	registerRunFlags(rootCmd)

	// Register custom completions for flags
	registerCompletions(rootCmd)

	// Add subcommands
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(presetsCmd())
	rootCmd.AddCommand(probeCmd())
	rootCmd.AddCommand(completionCmd())

	return fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version),
		fang.WithCommit(commit),
		fang.WithNotifySignal(os.Interrupt),
	)
}
