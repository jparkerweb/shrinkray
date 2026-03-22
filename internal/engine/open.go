package engine

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

// OpenFolder opens the file manager showing the given path.
// On Windows it uses explorer /select, on macOS open -R, on Linux xdg-open.
func OpenFolder(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	cmd := openFolderCmd(absPath)

	if err := cmd.Start(); err != nil {
		slog.Warn("failed to open folder", "path", absPath, "error", err)
		return fmt.Errorf("could not open folder: %w", err)
	}

	// Don't wait for the process to finish — it's non-blocking
	go func() {
		_ = cmd.Wait()
	}()

	return nil
}
