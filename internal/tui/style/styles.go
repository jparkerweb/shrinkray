package style

import (
	"os"
	"runtime"
	"strings"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

// ColorSupport describes the terminal's color capability level.
type ColorSupport int

const (
	ColorNone    ColorSupport = iota // No color support
	ColorANSI16                      // Basic 16 colors
	ColorANSI256                     // 256 colors
	ColorTrue                        // TrueColor (24-bit)
)

// DetectedColorSupport holds the detected color support level.
var DetectedColorSupport = detectColorSupport()

// detectColorSupport determines the terminal's color capability.
func detectColorSupport() ColorSupport {
	// Check NO_COLOR env var (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return ColorNone
	}

	// Check TERM env var for restricted terminals
	termVar := os.Getenv("TERM")
	if termVar == "dumb" {
		return ColorNone
	}

	// Windows legacy terminal detection
	if runtime.GOOS == "windows" {
		// Check if running in Windows Terminal, ConEmu, or similar modern terminal
		if os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuPID") != "" {
			return ColorTrue
		}
		// Check COLORTERM for TrueColor
		colorTerm := os.Getenv("COLORTERM")
		if colorTerm == "truecolor" || colorTerm == "24bit" {
			return ColorTrue
		}
		// Default to ANSI256 for modern Windows 10+
		return ColorANSI256
	}

	// Check COLORTERM env
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return ColorTrue
	}
	if strings.HasPrefix(termVar, "xterm-256") {
		return ColorANSI256
	}

	return ColorANSI16
}

// HasDarkBackground attempts to detect if the terminal has a dark background.
// Returns true if detection fails (most terminals are dark).
func HasDarkBackground() bool {
	// Lip Gloss v2 handles this via colorprofile; most terminals are dark.
	// For light terminal support, we check env hints.
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg != "" {
		// COLORFGBG format: "fg;bg" where bg > 8 usually means light
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			// Light backgrounds typically have bg values like "15" (white)
			if bg == "15" || bg == "7" {
				return false
			}
		}
	}
	return true // assume dark background (most common)
}

// HeaderStyle returns a style for the top header bar.
func HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Text).
		Background(ActiveTheme.Primary).
		Bold(true).
		Padding(0, 1)
}

// FooterStyle returns a style for the bottom footer bar.
func FooterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.TextDim).
		Background(ActiveTheme.Surface).
		Padding(0, 1)
}

// TitleStyle returns a style for screen titles.
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Primary).
		Bold(true).
		MarginBottom(1)
}

// SubtitleStyle returns a style for subtitles or secondary headings.
func SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Accent).
		Italic(true)
}

// CardStyle returns a style for bordered card containers.
func CardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ActiveTheme.Border).
		Foreground(ActiveTheme.Text).
		Padding(1, 2)
}

// CardActiveStyle returns a style for a highlighted/active card.
func CardActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ActiveTheme.Primary).
		Foreground(ActiveTheme.Text).
		Bold(true).
		Padding(1, 2)
}

// ButtonStyle returns a style for inactive buttons.
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Text).
		Background(ActiveTheme.Surface).
		Padding(0, 2)
}

// ButtonActiveStyle returns a style for active/focused buttons.
func ButtonActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Text).
		Background(ActiveTheme.Primary).
		Bold(true).
		Padding(0, 2)
}

// StatBoxStyle returns a style for small bordered stat display boxes.
func StatBoxStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ActiveTheme.Muted).
		Foreground(ActiveTheme.Text).
		Padding(0, 1).
		Width(14)
}

// ProgressBarStyle returns a style for the progress bar container.
func ProgressBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Primary)
}

// ErrorStyle returns a style for error messages.
func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Error).
		Bold(true)
}

// MutedStyle returns a style for muted/secondary text.
func MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Muted)
}

// KeyHintStyle returns a style for keyboard shortcut hints.
func KeyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Accent)
}

// AccentStyle returns a style for accent-colored text.
func AccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Accent)
}

// SuccessStyle returns a style for success-colored text.
func SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Success).
		Bold(true)
}

// WarningStyle returns a style for warning-colored text.
func WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Warning)
}

// LabelStyle returns a style for labels in stat displays.
func LabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Muted).
		Width(12)
}

// ValueStyle returns a style for values in stat displays.
func ValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Text).
		Bold(true)
}

// BadgeStyle returns a style for small inline badges.
func BadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Text).
		Background(ActiveTheme.Primary).
		Padding(0, 1).
		Bold(true)
}

// TerminalWidth returns the current terminal width, defaulting to 80.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// TerminalHeight returns the current terminal height, defaulting to 24.
func TerminalHeight() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h <= 0 {
		return 24
	}
	return h
}
