package screens

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

// PresetsModel is the model for the preset selection screen.
type PresetsModel struct {
	allPresets      []presets.Preset
	filtered        []presets.Preset
	cursor          int
	activeCategory  int // 0=Quality, 1=Purpose, 2=Platform
	videoInfo       *engine.VideoInfo
	recommendations []presets.Recommendation
	width           int
	height          int
}

var categories = []presets.Category{
	presets.CategoryQuality,
	presets.CategoryPurpose,
	presets.CategoryPlatform,
}

var categoryLabels = []string{"Quality", "Purpose", "Platform"}

// NewPresetsModel creates a new presets selection model.
func NewPresetsModel() PresetsModel {
	m := PresetsModel{
		activeCategory: 0,
	}
	m.refreshPresets()
	return m
}

// SetVideoInfo sets the source video info for size estimation.
func (m *PresetsModel) SetVideoInfo(info *engine.VideoInfo) {
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
	}
}

func (m *PresetsModel) refreshPresets() {
	m.allPresets = presets.All()
	cat := categories[m.activeCategory]
	m.filtered = presets.ByCategory(cat)

	// Sort recommended presets first
	if len(m.recommendations) > 0 {
		recKeys := make(map[string]int)
		for _, r := range m.recommendations {
			recKeys[r.Preset.Key] = r.Score
		}

		// Stable sort: recommended first, then by original order
		recommended := make([]presets.Preset, 0)
		normal := make([]presets.Preset, 0)
		for _, p := range m.filtered {
			if _, ok := recKeys[p.Key]; ok {
				recommended = append(recommended, p)
			} else {
				normal = append(normal, p)
			}
		}
		m.filtered = append(recommended, normal...)
	}
}

// Init returns nil.
func (m PresetsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the presets screen.
func (m PresetsModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		numPresets := len(m.filtered)
		cols := m.columns()

		switch msg.String() {
		case "up":
			if m.cursor >= cols {
				m.cursor -= cols
			}
		case "down":
			if m.cursor+cols < numPresets {
				m.cursor += cols
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < numPresets-1 {
				m.cursor++
			}
		case "tab":
			m.activeCategory = (m.activeCategory + 1) % len(categories)
			m.cursor = 0
			m.refreshPresets()
		case "shift+tab":
			m.activeCategory = (m.activeCategory + len(categories) - 1) % len(categories)
			m.cursor = 0
			m.refreshPresets()
		case "enter":
			if m.cursor < numPresets {
				selected := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return messages.PresetSelectedMsg{Preset: selected}
				}
			}
		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		}
	}

	return m, nil
}

func (m PresetsModel) columns() int {
	if m.width >= 80 {
		return 2
	}
	return 1
}

// View renders the preset selection screen.
func (m PresetsModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Choose a Preset")
	b.WriteString(title)
	b.WriteString("\n")

	// Category tabs
	var tabs []string
	for i, label := range categoryLabels {
		count := len(presets.ByCategory(categories[i]))
		tabText := fmt.Sprintf(" %s (%d) ", label, count)
		if i == m.activeCategory {
			tabs = append(tabs, style.BadgeStyle().Render(tabText))
		} else {
			tabs = append(tabs, style.MutedStyle().Render(tabText))
		}
	}
	b.WriteString(strings.Join(tabs, " "))
	b.WriteString("\n\n")

	cols := m.columns()
	isWide := m.width >= 80
	cardWidth := m.cardWidth()

	// Render presets in grid
	for i := 0; i < len(m.filtered); i += cols {
		var row []string
		for j := 0; j < cols && i+j < len(m.filtered); j++ {
			idx := i + j
			p := m.filtered[idx]
			card := m.renderPresetCard(p, idx == m.cursor, cardWidth)
			row = append(row, card)
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
		b.WriteString("\n")
	}

	// Detail panel for selected preset (wide mode)
	if isWide && m.cursor < len(m.filtered) {
		b.WriteString("\n")
		detail := m.renderPresetDetail(m.filtered[m.cursor])
		b.WriteString(detail)
	}

	// Navigation hints
	b.WriteString("\n")
	b.WriteString(style.KeyHintStyle().Render("[Arrows] Navigate  [Tab] Category  [Enter] Select  [Esc] Back"))

	return b.String()
}

func (m PresetsModel) cardWidth() int {
	cols := m.columns()
	available := m.width - 4
	if available < 30 {
		available = 30
	}
	w := available / cols
	if w > 38 {
		w = 38
	}
	return w
}

// isRecommended checks if a preset is in the recommendations list.
func (m PresetsModel) isRecommended(key string) bool {
	for _, r := range m.recommendations {
		if r.Preset.Key == key {
			return true
		}
	}
	return false
}

func (m PresetsModel) renderPresetCard(p presets.Preset, active bool, width int) string {
	var cardSt lipgloss.Style
	if active {
		cardSt = style.CardActiveStyle().Width(width)
	} else {
		cardSt = style.CardStyle().Width(width)
	}

	isNarrow := m.width < 60

	var content strings.Builder

	// Recommended badge
	if m.isRecommended(p.Key) {
		content.WriteString(style.WarningStyle().Render("\u2605 "))
	}

	content.WriteString(p.Icon + " ")
	content.WriteString(style.ValueStyle().Render(p.Name))
	content.WriteString("\n")

	if !isNarrow {
		content.WriteString(style.MutedStyle().Render(p.Description))
		content.WriteString("\n")
	}

	// Estimated size using real estimation engine
	if m.videoInfo != nil && m.videoInfo.Size > 0 {
		est := engine.EstimateSize(m.videoInfo, p)
		if est > 0 {
			content.WriteString(style.AccentStyle().Render(fmt.Sprintf("~%s", formatBytes(est))))

			// Compression ratio
			ratio := float64(m.videoInfo.Size) / float64(est)
			if ratio > 1 {
				content.WriteString(style.MutedStyle().Render(fmt.Sprintf(" (%.1fx)", ratio)))
			}
		}
	}

	// Feasibility warning for target-size presets
	if p.TargetSizeMB > 0 && m.videoInfo != nil && m.videoInfo.Duration > 0 {
		dur := time.Duration(m.videoInfo.Duration * float64(time.Second))
		result := engine.FeasibilityCheck(
			int64(p.TargetSizeMB*1024*1024),
			dur,
			m.videoInfo.Width,
			m.videoInfo.Height,
		)
		switch result.Status {
		case engine.FeasibilityImpossible:
			content.WriteString("\n")
			content.WriteString(style.ErrorStyle().Render("\u26a0 Impossible at this resolution"))
		case engine.FeasibilityWarning:
			content.WriteString("\n")
			content.WriteString(style.WarningStyle().Render("\u26a0 Quality loss likely"))
		}
	}

	// Two-pass indicator
	if p.TwoPass || p.TargetSizeMB > 0 {
		content.WriteString("\n")
		content.WriteString(style.MutedStyle().Render("[two-pass]"))
	}

	return cardSt.Render(content.String())
}

func (m PresetsModel) renderPresetDetail(p presets.Preset) string {
	var detail strings.Builder

	detail.WriteString(style.SubtitleStyle().Render("Preset Details"))
	detail.WriteString("\n\n")
	detail.WriteString(style.LabelStyle().Render("Codec:     "))
	detail.WriteString(style.ValueStyle().Render(strings.ToUpper(p.Codec)))
	detail.WriteString("\n")
	detail.WriteString(style.LabelStyle().Render("Container: "))
	detail.WriteString(style.ValueStyle().Render(p.Container))
	detail.WriteString("\n")

	if p.TargetSizeMB > 0 {
		detail.WriteString(style.LabelStyle().Render("Target:    "))
		detail.WriteString(style.ValueStyle().Render(fmt.Sprintf("%.0f MB", p.TargetSizeMB)))
		detail.WriteString("\n")
	}

	detail.WriteString(style.LabelStyle().Render("CRF:       "))
	detail.WriteString(style.ValueStyle().Render(fmt.Sprintf("%d", p.CRF)))
	detail.WriteString("\n")
	detail.WriteString(style.LabelStyle().Render("Speed:     "))
	detail.WriteString(style.ValueStyle().Render(p.SpeedPreset))
	detail.WriteString("\n")
	if p.Resolution != "" {
		detail.WriteString(style.LabelStyle().Render("Max Res:   "))
		detail.WriteString(style.ValueStyle().Render(p.Resolution))
		detail.WriteString("\n")
	}
	if p.MaxFPS > 0 {
		detail.WriteString(style.LabelStyle().Render("Max FPS:   "))
		detail.WriteString(style.ValueStyle().Render(fmt.Sprintf("%d", p.MaxFPS)))
		detail.WriteString("\n")
	}
	detail.WriteString(style.LabelStyle().Render("Audio:     "))
	if p.AudioCodec == "copy" {
		detail.WriteString(style.ValueStyle().Render("copy (passthrough)"))
	} else if p.AudioBitrate != "" {
		detail.WriteString(style.ValueStyle().Render(fmt.Sprintf("%s @ %s", p.AudioCodec, p.AudioBitrate)))
	} else {
		detail.WriteString(style.ValueStyle().Render(p.AudioCodec))
	}
	detail.WriteString("\n")

	if p.TwoPass {
		detail.WriteString(style.LabelStyle().Render("Encoding:  "))
		detail.WriteString(style.ValueStyle().Render("Two-pass"))
		detail.WriteString("\n")
	}

	// Estimated compression ratio
	if m.videoInfo != nil && m.videoInfo.Size > 0 {
		est := engine.EstimateSize(m.videoInfo, p)
		if est > 0 {
			ratio := float64(m.videoInfo.Size) / float64(est)
			savings := (1 - float64(est)/float64(m.videoInfo.Size)) * 100
			detail.WriteString("\n")
			detail.WriteString(style.LabelStyle().Render("Est. Size: "))
			detail.WriteString(style.AccentStyle().Render(formatBytes(est)))
			detail.WriteString("\n")
			detail.WriteString(style.LabelStyle().Render("Savings:   "))
			if savings > 0 {
				detail.WriteString(style.SuccessStyle().Render(fmt.Sprintf("~%.0f%% (%.1fx)", savings, ratio)))
			} else {
				detail.WriteString(style.WarningStyle().Render("File may get larger"))
			}
			detail.WriteString("\n")
		}
	}

	width := m.width - 4
	if width > 44 {
		width = 44
	}
	return style.CardStyle().Width(width).Render(detail.String())
}
