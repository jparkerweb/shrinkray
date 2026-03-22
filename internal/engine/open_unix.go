//go:build !windows

package engine

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

func openFolderCmd(absPath string) *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", "-R", absPath)
	}
	// linux and other unix — open the directory
	dir := filepath.Dir(absPath)
	return exec.Command("xdg-open", dir)
}

// OpenFolderCommand returns the command that would be used to open the folder,
// without executing it. Useful for testing.
func OpenFolderCommand(path string) (name string, args []string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	if runtime.GOOS == "darwin" {
		return "open", []string{"-R", absPath}
	}
	dir := filepath.Dir(absPath)
	return "xdg-open", []string{dir}
}
