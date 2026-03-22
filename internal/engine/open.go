package engine

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenFolder opens the file manager showing the given path.
// On Windows it uses explorer /select, on macOS open -R, on Linux xdg-open.
func OpenFolder(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", fmt.Sprintf("/select,%s", absPath))
	case "darwin":
		cmd = exec.Command("open", "-R", absPath)
	default: // linux and other unix
		dir := filepath.Dir(absPath)
		cmd = exec.Command("xdg-open", dir)
	}

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

// OpenFolderCommand returns the command that would be used to open the folder,
// without executing it. Useful for testing.
func OpenFolderCommand(path string) (name string, args []string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	switch runtime.GOOS {
	case "windows":
		return "explorer", []string{fmt.Sprintf("/select,%s", absPath)}
	case "darwin":
		return "open", []string{"-R", absPath}
	default:
		dir := filepath.Dir(absPath)
		return "xdg-open", []string{dir}
	}
}
