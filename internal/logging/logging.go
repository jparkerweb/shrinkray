package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jparkerweb/shrinkray/internal/config"
)

// Setup initializes the logging system.
// mode is either "tui" or "headless".
// level is one of: debug, info, warn, error.
// If the SHRINKRAY_LOG_LEVEL environment variable is set, it overrides the level parameter.
func Setup(mode string, level string) error {
	// Env var override
	if envLevel := os.Getenv("SHRINKRAY_LOG_LEVEL"); envLevel != "" {
		level = envLevel
	}

	parsedLevel := parseLevel(level)

	var w io.Writer

	if mode == "tui" {
		logDir, err := config.LogDir()
		if err != nil {
			return fmt.Errorf("failed to determine log directory: %w", err)
		}
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		logPath := filepath.Join(logDir, "shrinkray.log")
		// Truncate on startup for simplicity (rotating can be added later)
		f, err := os.Create(logPath)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		w = f
	} else {
		w = os.Stderr
	}

	logger := log.NewWithOptions(w, log.Options{
		ReportTimestamp: true,
		Level:           parsedLevel,
	})

	// Set as default slog logger for stdlib compatibility
	slog.SetDefault(slog.New(logger))

	return nil
}

func parseLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn", "warning":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}
