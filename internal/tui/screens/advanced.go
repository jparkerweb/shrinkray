package screens

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

// Form field identifiers
const (
	fieldCodec        = 0
	fieldCRF          = 1
	fieldResolution   = 2
	fieldMaxFPS       = 3
	fieldAudioCodec   = 4
	fieldAudioBitrate = 5
	fieldAudioChan    = 6
	fieldEncoder      = 7
	fieldSuffix       = 8
	fieldConflict     = 9
	fieldCount        = 10
)

// Available options for each dropdown
var (
	codecOptions      = []string{"h264", "h265", "av1", "vp9"}
	resolutionOptions = []string{"source", "2160p (3840x2160)", "1440p (2560x1440)", "1080p (1920x1080)", "720p (1280x720)", "480p (854x480)"}
	audioCodecOptions = []string{"aac", "opus", "copy"}
	audioBitrateOpts  = []string{"64k", "96k", "128k", "192k", "256k", "320k"}
	channelOptions    = []string{"source", "stereo", "mono"}
	conflictOptions   = []string{"autorename", "skip", "overwrite"}
)

// AdvancedModel is the model for the advanced options screen.
type AdvancedModel struct {
	width  int
	height int

	// Current field index
	focusIndex int

	// Video group
	codecIdx      int
	crf           int
	resolutionIdx int
	maxFPS        int

	// Audio group
	audioCodecIdx   int
	audioBitrateIdx int
	audioChannelIdx int

	// Encoder group
	encoderIdx   int
	encoderNames []string // populated from HW detection

	// Output group
	suffix      string
	conflictIdx int

	// Reference data
	preset    *presets.Preset
	videoInfo *engine.VideoInfo
}

// NewAdvancedModel creates a new advanced options model.
func NewAdvancedModel() AdvancedModel {
	return AdvancedModel{
		crf:             23,
		maxFPS:          0,
		suffix:          "_shrunk",
		codecIdx:        1, // h265 default
		resolutionIdx:   0, // source
		audioCodecIdx:   0, // aac
		audioBitrateIdx: 2, // 128k
		audioChannelIdx: 0, // source
		encoderNames:    []string{"Software"},
		encoderIdx:      0,
		conflictIdx:     0, // autorename
	}
}

// SetPreset populates form with preset values.
func (m *AdvancedModel) SetPreset(p *presets.Preset) {
	m.preset = p
	if p == nil {
		return
	}

	// Codec
	for i, c := range codecOptions {
		if c == p.Codec {
			m.codecIdx = i
			break
		}
	}

	// CRF
	m.crf = p.CRF

	// Resolution
	m.resolutionIdx = 0 // source
	if p.Resolution != "" {
		for i, r := range resolutionOptions {
			if strings.Contains(r, p.Resolution) {
				m.resolutionIdx = i
				break
			}
		}
	}

	// MaxFPS
	m.maxFPS = p.MaxFPS

	// Audio codec
	for i, a := range audioCodecOptions {
		if a == p.AudioCodec {
			m.audioCodecIdx = i
			break
		}
	}

	// Audio bitrate
	for i, b := range audioBitrateOpts {
		if b == p.AudioBitrate {
			m.audioBitrateIdx = i
			break
		}
	}

	// Audio channels
	if p.AudioChannels == 2 {
		m.audioChannelIdx = 1
	} else if p.AudioChannels == 1 {
		m.audioChannelIdx = 2
	} else {
		m.audioChannelIdx = 0
	}

	// Suffix
	if m.suffix == "" {
		m.suffix = "_shrunk"
	}
}

// SetVideoInfo sets the source video info.
func (m *AdvancedModel) SetVideoInfo(info *engine.VideoInfo) {
	m.videoInfo = info
}

// SetHWEncoders populates the encoder dropdown with detected hardware encoders.
func (m *AdvancedModel) SetHWEncoders(encoders []engine.HWEncoder) {
	m.encoderNames = []string{"Software"}
	for _, enc := range encoders {
		if enc.Available {
			label := enc.DisplayName
			if enc.GPU != "" {
				label += " (" + enc.GPU + ")"
			}
			m.encoderNames = append(m.encoderNames, label+"|"+enc.Name)
		}
	}
}

// Init returns nil.
func (m AdvancedModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the advanced options screen.
func (m AdvancedModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }

		case "enter":
			return m, func() tea.Msg {
				return messages.AdvancedOptionsMsg{
					Opts: m.buildOptions(),
				}
			}

		case "tab", "down", "j":
			m.focusIndex = (m.focusIndex + 1) % fieldCount
			return m, nil

		case "shift+tab", "up", "k":
			m.focusIndex = (m.focusIndex - 1 + fieldCount) % fieldCount
			return m, nil

		case "left", "h":
			m.adjustField(-1)
			return m, nil

		case "right", "l":
			m.adjustField(1)
			return m, nil
		}
	}

	return m, nil
}

// adjustField changes the currently focused field's value.
func (m *AdvancedModel) adjustField(delta int) {
	switch m.focusIndex {
	case fieldCodec:
		m.codecIdx = clampIdx(m.codecIdx+delta, len(codecOptions))
		// VP9 forces opus audio
		if codecOptions[m.codecIdx] == "vp9" {
			for i, a := range audioCodecOptions {
				if a == "opus" {
					m.audioCodecIdx = i
					break
				}
			}
		}
	case fieldCRF:
		maxCRF := 51
		if codecOptions[m.codecIdx] == "av1" {
			maxCRF = 63
		}
		m.crf = clamp(m.crf+delta, 0, maxCRF)
	case fieldResolution:
		m.resolutionIdx = clampIdx(m.resolutionIdx+delta, len(resolutionOptions))
	case fieldMaxFPS:
		m.maxFPS = clamp(m.maxFPS+delta*5, 0, 240)
	case fieldAudioCodec:
		// VP9 forces opus — don't allow changing
		if codecOptions[m.codecIdx] == "vp9" {
			return
		}
		m.audioCodecIdx = clampIdx(m.audioCodecIdx+delta, len(audioCodecOptions))
	case fieldAudioBitrate:
		m.audioBitrateIdx = clampIdx(m.audioBitrateIdx+delta, len(audioBitrateOpts))
	case fieldAudioChan:
		m.audioChannelIdx = clampIdx(m.audioChannelIdx+delta, len(channelOptions))
	case fieldEncoder:
		m.encoderIdx = clampIdx(m.encoderIdx+delta, len(m.encoderNames))
	case fieldSuffix:
		// No left/right for text — could add char editing later
	case fieldConflict:
		m.conflictIdx = clampIdx(m.conflictIdx+delta, len(conflictOptions))
	}
}

// buildOptions creates modified EncodeOptions from the form state.
func (m AdvancedModel) buildOptions() messages.AdvancedOptions {
	opts := messages.AdvancedOptions{
		Codec:        codecOptions[m.codecIdx],
		CRF:          m.crf,
		MaxFPS:       m.maxFPS,
		AudioCodec:   audioCodecOptions[m.audioCodecIdx],
		AudioBitrate: audioBitrateOpts[m.audioBitrateIdx],
		Suffix:       m.suffix,
		ConflictMode: conflictOptions[m.conflictIdx],
	}

	// Resolution
	if m.resolutionIdx > 0 {
		res := resolutionOptions[m.resolutionIdx]
		// Extract "WxH" from "1080p (1920x1080)"
		if paren := strings.Index(res, "("); paren >= 0 {
			inner := res[paren+1 : len(res)-1]
			opts.Resolution = inner
		}
	}

	// Audio channels
	switch channelOptions[m.audioChannelIdx] {
	case "stereo":
		opts.AudioChannels = 2
	case "mono":
		opts.AudioChannels = 1
	default:
		opts.AudioChannels = 0
	}

	// HW encoder
	if m.encoderIdx > 0 && m.encoderIdx < len(m.encoderNames) {
		name := m.encoderNames[m.encoderIdx]
		if parts := strings.SplitN(name, "|", 2); len(parts) == 2 {
			opts.HWEncoderName = parts[1]
		}
	}

	return opts
}

// View renders the advanced options form.
func (m AdvancedModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Advanced Options")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Video group
	b.WriteString(style.AccentStyle().Render("--- Video ---"))
	b.WriteString("\n")
	m.renderField(&b, fieldCodec, "Codec", codecOptions[m.codecIdx])
	crfRange := "0-51"
	if codecOptions[m.codecIdx] == "av1" {
		crfRange = "0-63"
	}
	m.renderField(&b, fieldCRF, "CRF", fmt.Sprintf("%d (%s)", m.crf, crfRange))
	m.renderField(&b, fieldResolution, "Resolution", resolutionOptions[m.resolutionIdx])
	fpsStr := "source"
	if m.maxFPS > 0 {
		fpsStr = strconv.Itoa(m.maxFPS)
	}
	m.renderField(&b, fieldMaxFPS, "Max FPS", fpsStr)
	b.WriteString("\n")

	// Audio group
	b.WriteString(style.AccentStyle().Render("--- Audio ---"))
	b.WriteString("\n")
	audioLabel := audioCodecOptions[m.audioCodecIdx]
	if codecOptions[m.codecIdx] == "vp9" {
		audioLabel += " (required for VP9)"
	}
	m.renderField(&b, fieldAudioCodec, "Audio Codec", audioLabel)
	m.renderField(&b, fieldAudioBitrate, "Bitrate", audioBitrateOpts[m.audioBitrateIdx])
	m.renderField(&b, fieldAudioChan, "Channels", channelOptions[m.audioChannelIdx])
	b.WriteString("\n")

	// Encoder group
	b.WriteString(style.AccentStyle().Render("--- Encoder ---"))
	b.WriteString("\n")
	encoderDisplay := "Software"
	if m.encoderIdx > 0 && m.encoderIdx < len(m.encoderNames) {
		parts := strings.SplitN(m.encoderNames[m.encoderIdx], "|", 2)
		encoderDisplay = parts[0]
	}
	m.renderField(&b, fieldEncoder, "HW Encoder", encoderDisplay)
	b.WriteString("\n")

	// Output group
	b.WriteString(style.AccentStyle().Render("--- Output ---"))
	b.WriteString("\n")
	m.renderField(&b, fieldSuffix, "Suffix", m.suffix)
	m.renderField(&b, fieldConflict, "Conflict", conflictOptions[m.conflictIdx])
	b.WriteString("\n")

	// Navigation hints
	b.WriteString(style.KeyHintStyle().Render("[Arrows] Change values  [Enter] Apply  [Esc] Cancel"))

	return b.String()
}

func (m AdvancedModel) renderField(b *strings.Builder, idx int, label, value string) {
	cursor := "  "
	if m.focusIndex == idx {
		cursor = style.AccentStyle().Render("> ")
	}

	labelStyle := style.LabelStyle()
	valueStyle := style.ValueStyle()

	if m.focusIndex == idx {
		valueStyle = style.SuccessStyle()
	}

	b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, labelStyle.Render(label+":"), valueStyle.Render(value)))
}

func clampIdx(v, max int) int {
	if v < 0 {
		return max - 1
	}
	if v >= max {
		return 0
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

