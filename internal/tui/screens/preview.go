package screens

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

// PreviewModel is the model for the preview/confirmation screen.
type PreviewModel struct {
	videoInfo   *engine.VideoInfo
	preset      *presets.Preset
	outputPath  string
	hwEncoder   string // e.g. "hevc_nvenc"
	hwName      string // e.g. "nvenc"
	hwEncoders  []engine.HWEncoder
	width       int
	height      int
}

// NewPreviewModel creates a new preview model.
func NewPreviewModel() PreviewModel {
	return PreviewModel{}
}

// SetVideoInfo sets the source video info.
func (m *PreviewModel) SetVideoInfo(info *engine.VideoInfo) {
	m.videoInfo = info
	m.resolveOutputPath()
}

// SetPreset sets the selected preset.
func (m *PreviewModel) SetPreset(p *presets.Preset) {
	m.preset = p
	m.resolveOutputPath()
}

// SetHWEncoders stores detected HW encoders for display.
func (m *PreviewModel) SetHWEncoders(encoders []engine.HWEncoder) {
	m.hwEncoders = encoders
	// Auto-select first available encoder if any
	for _, enc := range encoders {
		if enc.Available && m.preset != nil {
			hwName := engine.HWEncoderName(enc.Name, m.preset.Codec)
			if hwName != "" {
				m.hwEncoder = hwName
				m.hwName = enc.Name
				break
			}
		}
	}
}

func (m *PreviewModel) resolveOutputPath() {
	if m.videoInfo == nil {
		return
	}
	dir := filepath.Dir(m.videoInfo.Path)
	ext := filepath.Ext(m.videoInfo.Path)
	base := strings.TrimSuffix(filepath.Base(m.videoInfo.Path), ext)

	container := "mp4"
	if m.preset != nil && m.preset.Container != "" {
		container = m.preset.Container
	}
	m.outputPath = filepath.Join(dir, base+"_shrunk."+container)
}

// Init returns nil.
func (m PreviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the preview screen.
func (m PreviewModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.videoInfo != nil && m.preset != nil {
				opts := engine.EncodeOptions{
					Input:      m.videoInfo.Path,
					Output:     m.outputPath,
					Preset:     *m.preset,
					HWEncoder:  m.hwEncoder,
					SourceInfo: m.videoInfo,
				}
				return m, func() tea.Msg {
					return messages.EncodeStartMsg{Opts: opts}
				}
			}
		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		case "a", "e":
			// Navigate to advanced options screen
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenAdvanced}
			}
		}
	}

	return m, nil
}

// View renders the preview/confirmation screen.
func (m PreviewModel) View() string {
	if m.videoInfo == nil || m.preset == nil {
		return style.MutedStyle().Render("No video or preset selected.")
	}

	var b strings.Builder
	v := m.videoInfo
	p := m.preset

	title := style.TitleStyle().Render("Preview & Confirm")
	b.WriteString(title)
	b.WriteString("\n")

	isWide := m.width >= 80
	cardWidth := (m.width - 6) / 2
	if cardWidth > 36 {
		cardWidth = 36
	}
	if cardWidth < 26 {
		cardWidth = 26
	}

	// Before card
	var beforeContent strings.Builder
	beforeContent.WriteString(style.ValueStyle().Render("BEFORE"))
	beforeContent.WriteString("\n\n")
	beforeContent.WriteString(style.LabelStyle().Render("File:     "))
	beforeContent.WriteString(filepath.Base(v.Path))
	beforeContent.WriteString("\n")
	beforeContent.WriteString(style.LabelStyle().Render("Size:     "))
	beforeContent.WriteString(style.ValueStyle().Render(formatBytes(v.Size)))
	beforeContent.WriteString("\n")
	beforeContent.WriteString(style.LabelStyle().Render("Codec:    "))
	beforeContent.WriteString(strings.ToUpper(v.Codec))
	beforeContent.WriteString("\n")
	beforeContent.WriteString(style.LabelStyle().Render("Res:      "))
	beforeContent.WriteString(v.Resolution())
	beforeContent.WriteString("\n")
	beforeContent.WriteString(style.LabelStyle().Render("Bitrate:  "))
	beforeContent.WriteString(formatBitrate(v.Bitrate))

	beforeCard := style.CardStyle().Width(cardWidth).Render(beforeContent.String())

	// After card
	estSize := engine.EstimateSize(v, *p)
	var afterContent strings.Builder
	afterContent.WriteString(style.ValueStyle().Render("AFTER (estimated)"))
	afterContent.WriteString("\n\n")
	afterContent.WriteString(style.LabelStyle().Render("Preset:   "))
	afterContent.WriteString(p.Icon + " " + p.Name)
	afterContent.WriteString("\n")
	afterContent.WriteString(style.LabelStyle().Render("Size:     "))
	afterContent.WriteString(style.SuccessStyle().Render(fmt.Sprintf("~%s", formatBytes(estSize))))
	afterContent.WriteString("\n")
	afterContent.WriteString(style.LabelStyle().Render("Codec:    "))
	afterContent.WriteString(strings.ToUpper(p.Codec))
	afterContent.WriteString("\n")
	afterContent.WriteString(style.LabelStyle().Render("Res:      "))
	if p.Resolution != "" {
		afterContent.WriteString(p.Resolution)
	} else {
		afterContent.WriteString(v.Resolution() + " (keep)")
	}
	afterContent.WriteString("\n")
	afterContent.WriteString(style.LabelStyle().Render("Container:"))
	afterContent.WriteString(p.Container)

	// Encoder info
	afterContent.WriteString("\n")
	afterContent.WriteString(style.LabelStyle().Render("Encoder:  "))
	if m.hwEncoder != "" {
		encoderDisplay := fmt.Sprintf("%s (%s)", strings.ToUpper(m.hwName), m.hwEncoder)
		afterContent.WriteString(style.AccentStyle().Render(encoderDisplay))
	} else {
		fmt.Fprintf(&afterContent, "Software (%s)", codecToLibDisplay(p.Codec))
	}

	afterCard := style.CardActiveStyle().Width(cardWidth).Render(afterContent.String())

	// Layout: side-by-side or stacked
	if isWide {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, beforeCard, "  ", afterCard))
	} else {
		b.WriteString(beforeCard)
		b.WriteString("\n")
		b.WriteString(afterCard)
	}
	b.WriteString("\n\n")

	// Savings summary
	savings := float64(v.Size-estSize) / float64(v.Size) * 100
	savingsLine := fmt.Sprintf("Estimated savings: %s (%.0f%%)",
		formatBytes(v.Size-estSize), savings)
	b.WriteString(style.SuccessStyle().Render(savingsLine))
	b.WriteString("\n")

	// Efficiency hints
	if m.hwEncoder != "" {
		b.WriteString(style.MutedStyle().Render("  ~ 2-3x faster encoding, slightly larger file size"))
		b.WriteString("\n")
	}
	if p.TwoPass {
		b.WriteString(style.MutedStyle().Render("  Two-pass encoding -- slower but precise file size"))
		b.WriteString("\n")
	}
	if p.Codec == "av1" && m.hwEncoder == "" {
		b.WriteString(style.WarningStyle().Render("  AV1 software encoding is significantly slower than H.264/H.265"))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Output path
	b.WriteString(style.MutedStyle().Render("Output: "))
	b.WriteString(m.outputPath)
	b.WriteString("\n\n")

	// Navigation hints
	b.WriteString(style.KeyHintStyle().Render("[Enter] Start encoding  [a] Advanced options  [Esc] Back"))

	return b.String()
}

// codecToLibDisplay returns a human-friendly software encoder name.
func codecToLibDisplay(codec string) string {
	switch strings.ToLower(codec) {
	case "h264":
		return "libx264"
	case "h265", "hevc":
		return "libx265"
	case "av1":
		return "libsvtav1"
	case "vp9":
		return "libvpx-vp9"
	default:
		return "libx265"
	}
}
