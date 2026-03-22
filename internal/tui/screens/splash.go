package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

const splashAutoAdvance = 2 * time.Second

// asciiLogo is the ASCII art for the shrinkray logo.
var asciiLogo = []string{
	`         __         _       __                    `,
	`   _____/ /_  _____(_)___  / /____________  __  __`,
	`  / ___/ __ \/ ___/ / __ \/ //_/ ___/ __  \/ / / /`,
	` (__  ) / / / /  / / / / / ,< / /  / /_/ / /_/ / `,
	`/____/_/ /_/_/  /_/_/ /_/_/|_/_/   \__,_/\__, /  `,
	`                                        /____/   `,
}

type splashTickMsg struct{}

type ffmpegDetectedMsg struct {
	FFmpegVer  string
	FFprobeVer string
}

// SplashModel is the model for the splash screen.
type SplashModel struct {
	width       int
	height      int
	ffmpegVer   string
	ffprobeVer  string
	hwEncoders  []engine.HWEncoder
	hwDetected  bool
	hwDetecting bool
	timerFired  bool
}

// NewSplashModel creates a new splash screen model.
func NewSplashModel() SplashModel {
	return SplashModel{
		hwDetecting: true,
	}
}

// Init starts the auto-advance timer and detects FFmpeg versions and HW encoders.
func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(splashAutoAdvance, func(time.Time) tea.Msg {
			return splashTickMsg{}
		}),
		detectFFmpeg,
		detectHWEncoders,
	)
}

func detectFFmpeg() tea.Msg {
	ffmpegVer := "not found"
	ffprobeVer := "not found"

	if info, err := engine.DetectFFmpeg(); err == nil {
		ffmpegVer = info.Version
	}
	if info, err := engine.DetectFFprobe(); err == nil {
		ffprobeVer = info.Version
	}

	return ffmpegDetectedMsg{FFmpegVer: ffmpegVer, FFprobeVer: ffprobeVer}
}

func detectHWEncoders() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	encoders, _ := engine.DetectHardware(ctx)
	return messages.HWDetectedMsg{Encoders: encoders}
}

// Update handles messages for the splash screen.
func (m SplashModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case splashTickMsg:
		m.timerFired = true
		// Only auto-advance if HW detection is done
		if m.hwDetected {
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenFilePicker}
			}
		}
		// If HW detection takes >2s, reset timer for another 2s
		return m, tea.Tick(splashAutoAdvance, func(time.Time) tea.Msg {
			return splashTickMsg{}
		})

	case ffmpegDetectedMsg:
		m.ffmpegVer = msg.FFmpegVer
		m.ffprobeVer = msg.FFprobeVer
		return m, nil

	case messages.HWDetectedMsg:
		m.hwEncoders = msg.Encoders
		m.hwDetected = true
		m.hwDetecting = false
		// If timer already fired, auto-advance now
		if m.timerFired {
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenFilePicker}
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		// Any key press advances past splash
		return m, func() tea.Msg {
			return messages.NavigateMsg{Screen: messages.ScreenFilePicker}
		}
	}

	return m, nil
}

// View renders the splash screen.
func (m SplashModel) View() string {
	var b strings.Builder

	// Center vertically
	logoHeight := len(asciiLogo) + 12
	topPad := (m.height - logoHeight) / 3
	if topPad < 1 {
		topPad = 1
	}
	for i := 0; i < topPad; i++ {
		b.WriteString("\n")
	}

	// Render logo with gradient colors
	primaryStyle := style.TitleStyle().MarginBottom(0)
	accentStyle := style.AccentStyle()

	for i, line := range asciiLogo {
		if i%2 == 0 {
			b.WriteString(centerText(primaryStyle.Render(line), m.width))
		} else {
			b.WriteString(centerText(accentStyle.Render(line), m.width))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Tagline
	tagline := style.SubtitleStyle().Render("Less bytes, same vibes.")
	b.WriteString(centerText(tagline, m.width))
	b.WriteString("\n\n")

	// System info
	mutedStyle := style.MutedStyle()
	successStyle := style.SuccessStyle()

	ffmpegLine := fmt.Sprintf("FFmpeg:  %s", m.ffmpegVer)
	if m.ffmpegVer == "" {
		ffmpegLine = "FFmpeg:  detecting..."
	}
	b.WriteString(centerText(mutedStyle.Render(ffmpegLine), m.width))
	b.WriteString("\n")

	ffprobeLine := fmt.Sprintf("FFprobe: %s", m.ffprobeVer)
	if m.ffprobeVer == "" {
		ffprobeLine = "FFprobe: detecting..."
	}
	b.WriteString(centerText(mutedStyle.Render(ffprobeLine), m.width))
	b.WriteString("\n")

	// Hardware encoder status
	if m.hwDetecting {
		hwLine := "HW Accel: detecting..."
		b.WriteString(centerText(mutedStyle.Render(hwLine), m.width))
		b.WriteString("\n")
	} else if m.hwDetected {
		hasAvailable := false
		for _, enc := range m.hwEncoders {
			if enc.Available {
				hasAvailable = true
				codecStr := strings.Join(enc.Codecs, ", ")
				label := enc.DisplayName
				if enc.GPU != "" {
					label += " (" + enc.GPU + ")"
				}
				line := fmt.Sprintf("+ %s -- %s", label, codecStr)
				b.WriteString(centerText(successStyle.Render(line), m.width))
				b.WriteString("\n")
			}
		}
		// Show unavailable encoders dimmed
		for _, enc := range m.hwEncoders {
			if !enc.Available {
				line := fmt.Sprintf("- %s -- not available", enc.DisplayName)
				b.WriteString(centerText(mutedStyle.Render(line), m.width))
				b.WriteString("\n")
			}
		}
		if !hasAvailable {
			b.WriteString(centerText(mutedStyle.Render("HW Accel: none detected (software encoding)"), m.width))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Hint
	hint := style.KeyHintStyle().Render("Press any key to continue...")
	b.WriteString(centerText(hint, m.width))

	return b.String()
}

// centerText centers a string within the given width.
func centerText(s string, width int) string {
	if width <= 0 {
		return s
	}
	pad := (width - len(stripANSI(s))) / 2
	if pad <= 0 {
		return s
	}
	return strings.Repeat(" ", pad) + s
}

// stripANSI removes ANSI escape codes for length calculation.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
