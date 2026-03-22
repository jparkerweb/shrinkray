package engine

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// FFmpegInfo holds information about the detected FFmpeg binary.
type FFmpegInfo struct {
	Path    string
	Version string
}

// FFprobeInfo holds information about the detected FFprobe binary.
type FFprobeInfo struct {
	Path    string
	Version string
}

// DetectFFmpeg locates the FFmpeg binary and returns its path and version.
// It checks the FFMPEG_PATH environment variable first, then exec.LookPath.
func DetectFFmpeg() (*FFmpegInfo, error) {
	path, err := findBinary("ffmpeg", "FFMPEG_PATH")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w\n\n%s", err, installGuidance("ffmpeg"))
	}

	version, err := extractVersion(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get ffmpeg version: %w", err)
	}

	return &FFmpegInfo{Path: path, Version: version}, nil
}

// DetectFFprobe locates the FFprobe binary and returns its path and version.
// It checks the FFPROBE_PATH environment variable first, then exec.LookPath.
func DetectFFprobe() (*FFprobeInfo, error) {
	path, err := findBinary("ffprobe", "FFPROBE_PATH")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found: %w\n\n%s", err, installGuidance("ffprobe"))
	}

	version, err := extractVersion(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get ffprobe version: %w", err)
	}

	return &FFprobeInfo{Path: path, Version: version}, nil
}

func findBinary(name, envVar string) (string, error) {
	// Check environment variable first
	if envPath := os.Getenv(envVar); envPath != "" {
		// If it's a direct path, check it exists
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		// Also try LookPath in case it's just a name
		if p, err := exec.LookPath(envPath); err == nil {
			return p, nil
		}
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH", name)
	}
	return path, nil
}

var versionRegexp = regexp.MustCompile(`version\s+([\w.\-]+)`)

func extractVersion(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "-version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	matches := versionRegexp.FindStringSubmatch(string(out))
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// Fallback: return first line
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "unknown", nil
}

func installGuidance(tool string) string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("Install %s with Homebrew:\n  brew install ffmpeg", tool)
	case "windows":
		return fmt.Sprintf("Install %s with winget or chocolatey:\n  winget install ffmpeg\n  choco install ffmpeg", tool)
	default: // linux and others
		return fmt.Sprintf("Install %s with your package manager:\n  sudo apt install ffmpeg     # Debian/Ubuntu\n  sudo dnf install ffmpeg     # Fedora\n  sudo pacman -S ffmpeg       # Arch", tool)
	}
}
