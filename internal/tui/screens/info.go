package screens

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

// InfoModel is the model for the video info screen.
type InfoModel struct {
	videoInfo       *engine.VideoInfo
	recommendations []presets.Recommendation
	recCursor       int
	width           int
	height          int
}

// NewInfoModel creates a new video info model.
func NewInfoModel() InfoModel {
	return InfoModel{}
}

// SetVideoInfo sets the video info for display.
func (m *InfoModel) SetVideoInfo(info *engine.VideoInfo) {
	m.videoInfo = info
	if info != nil {
		meta := &presets.VideoMetadata{
			Width:     info.Width,
			Height:    info.Height,
			Framerate: info.Framerate,
			Bitrate:   info.Bitrate,
			Duration:  info.Duration,
			Size:      info.Size,
			Codec:     info.Codec,
		}
		m.recommendations = presets.Recommend(meta)
		m.recCursor = 0
	}
}

// Init returns nil.
func (m InfoModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the info screen.
func (m InfoModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			// If a recommendation is highlighted, pre-select that preset
			if len(m.recommendations) > 0 && m.recCursor >= 0 && m.recCursor < len(m.recommendations) {
				selected := m.recommendations[m.recCursor].Preset
				return m, func() tea.Msg {
					return messages.PresetSelectedMsg{Preset: selected}
				}
			}
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenPresets}
			}
		case "tab":
			// Navigate to presets browser without selecting a recommendation
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenPresets}
			}
		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		case "up", "k":
			if m.recCursor > 0 {
				m.recCursor--
			}
		case "down", "j":
			maxRec := len(m.recommendations)
			if maxRec > 3 {
				maxRec = 3
			}
			if m.recCursor < maxRec-1 {
				m.recCursor++
			}
		}
	}

	return m, nil
}

// View renders the video info screen.
func (m InfoModel) View() string {
	if m.videoInfo == nil {
		return style.MutedStyle().Render("No video info available.")
	}

	var b strings.Builder
	v := m.videoInfo

	title := style.TitleStyle().Render("Video Information")
	b.WriteString(title)
	b.WriteString("\n")

	isWide := m.width >= 80
	isNarrow := m.width < 60

	// Codec badge
	codecBadge := style.BadgeStyle().Render(strings.ToUpper(v.Codec))
	if v.IsHDR {
		codecBadge += " " + style.BadgeStyle().
			Background(style.ActiveTheme.Warning).
			Render(v.HDRFormat)
	}
	b.WriteString(codecBadge)
	b.WriteString("\n\n")

	// Main info card
	cardWidth := m.width - 4
	if cardWidth > 72 {
		cardWidth = 72
	}
	if cardWidth < 30 {
		cardWidth = 30
	}

	var info strings.Builder

	addRow := func(label, value string) {
		if isNarrow {
			info.WriteString(style.MutedStyle().Render(label))
			info.WriteString(" ")
			info.WriteString(style.ValueStyle().Render(value))
			info.WriteString("\n")
		} else {
			info.WriteString(style.LabelStyle().Render(label))
			info.WriteString(style.ValueStyle().Render(value))
			info.WriteString("\n")
		}
	}

	addRow("Resolution:", v.Resolution())
	addRow("Framerate:", fmt.Sprintf("%.2f fps", v.Framerate))
	addRow("Bitrate:", formatBitrate(v.Bitrate))
	addRow("Duration:", formatDuration(v.Duration))
	addRow("File Size:", formatBytes(v.Size))
	addRow("Format:", v.Format)
	addRow("Pixel Fmt:", v.PixelFormat)

	if v.AudioCodec != "" {
		info.WriteString("\n")
		addRow("Audio:", v.AudioCodec)
		if v.AudioBitrate > 0 {
			addRow("Audio Rate:", formatBitrate(v.AudioBitrate))
		}
		addRow("Channels:", fmt.Sprintf("%d", v.AudioChannels))
		if v.AudioSampleRate > 0 {
			addRow("Sample Rate:", fmt.Sprintf("%d Hz", v.AudioSampleRate))
		}
	}

	if v.SubtitleCount > 0 {
		addRow("Subtitles:", fmt.Sprintf("%d track(s)", v.SubtitleCount))
	}

	card := style.CardStyle().Width(cardWidth).Render(info.String())
	b.WriteString(card)
	b.WriteString("\n")

	// Quality assessment
	assessment := assessQuality(v)
	b.WriteString("\n")
	b.WriteString(style.SubtitleStyle().Render("Quality Assessment: "))
	b.WriteString(assessment)
	b.WriteString("\n")

	// Stat boxes (if wide enough)
	if isWide {
		b.WriteString("\n")
		boxes := []string{
			style.StatBoxStyle().Render(fmt.Sprintf("Codec\n%s", strings.ToUpper(v.Codec))),
			style.StatBoxStyle().Render(fmt.Sprintf("Res\n%s", v.Resolution())),
			style.StatBoxStyle().Render(fmt.Sprintf("Size\n%s", formatBytes(v.Size))),
			style.StatBoxStyle().Render(fmt.Sprintf("Bitrate\n%s", formatBitrate(v.Bitrate))),
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
		b.WriteString("\n")
	}

	// Recommendations section
	if len(m.recommendations) > 0 {
		b.WriteString("\n")
		b.WriteString(style.SubtitleStyle().Render("Recommended for you"))
		b.WriteString("\n\n")

		maxRec := len(m.recommendations)
		if maxRec > 3 {
			maxRec = 3
		}

		for i := 0; i < maxRec; i++ {
			rec := m.recommendations[i]
			prefix := "  "
			if i == m.recCursor {
				prefix = style.AccentStyle().Render("> ")
			}

			star := style.WarningStyle().Render("\u2605")
			presetName := style.ValueStyle().Render(rec.Preset.Icon + " " + rec.Preset.Name)
			reason := style.MutedStyle().Render(" - " + rec.Reason)

			b.WriteString(prefix)
			b.WriteString(star)
			b.WriteString(" ")
			b.WriteString(presetName)
			b.WriteString(reason)
			b.WriteString("\n")
		}
	}

	// Navigation hints
	b.WriteString("\n")
	b.WriteString(style.KeyHintStyle().Render("[Enter] Select recommendation  [Tab] Browse all presets  [Esc] Back"))

	return b.String()
}

// assessQuality returns a quality heuristic string based on bitrate-per-pixel.
func assessQuality(v *engine.VideoInfo) string {
	if v.Width == 0 || v.Height == 0 || v.Bitrate == 0 || v.Framerate == 0 {
		return style.MutedStyle().Render("Unable to assess")
	}

	pixels := float64(v.Width * v.Height)
	bpp := float64(v.Bitrate) / (pixels * v.Framerate)

	switch {
	case bpp < 0.02:
		return style.WarningStyle().Render("Overcompressed - may already be low quality")
	case bpp < 0.08:
		return style.SuccessStyle().Render("Well compressed - modest savings possible")
	default:
		return style.SuccessStyle().Render("Could shrink significantly!")
	}
}
