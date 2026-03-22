package screens

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/progress"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
)

// EncodingModel is the model for the encoding progress screen.
type EncodingModel struct {
	progress   progress.Model
	opts       *engine.EncodeOptions
	update     engine.ProgressUpdate
	logLines   []string
	startTime  time.Time
	cancel     context.CancelFunc
	progressCh <-chan engine.ProgressUpdate
	tempPath   string
	outputPath string
	done       bool
	errMsg     string
	isTwoPass  bool
	width      int
	height     int
}

// NewEncodingModel creates a new encoding model.
func NewEncodingModel() EncodingModel {
	p := progress.New(
		progress.WithDefaultBlend(),
		progress.WithWidth(50),
	)

	return EncodingModel{
		progress: p,
	}
}

// StartEncode begins the encoding process with the given options.
func (m *EncodingModel) StartEncode(opts engine.EncodeOptions) tea.Cmd {
	m.opts = &opts
	m.done = false
	m.errMsg = ""
	m.logLines = nil
	m.startTime = time.Now()
	m.update = engine.ProgressUpdate{}
	m.isTwoPass = engine.ShouldUseTwoPass(opts)

	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())

		// Use a temp path for safety
		tempPath := engine.TempPath(opts.Output)
		encOpts := opts
		encOpts.Output = tempPath

		var ch <-chan engine.ProgressUpdate
		var err error

		if engine.ShouldUseTwoPass(opts) {
			// Calculate bitrate for target-size presets
			if opts.Preset.TargetSizeMB > 0 {
				// We need to probe to get duration and audio bitrate
				probeInfo, probeErr := engine.Probe(ctx, opts.Input)
				if probeErr == nil && probeInfo.Duration > 0 {
					dur := time.Duration(probeInfo.Duration * float64(time.Second))
					audioBitrate := int64(128000) // default
					if probeInfo.AudioBitrate > 0 {
						audioBitrate = probeInfo.AudioBitrate
					}
					targetBytes := int64(opts.Preset.TargetSizeMB * 1024 * 1024)
					encOpts.VideoBitrate = engine.CalculateBitrate(targetBytes, dur, audioBitrate)

					// Check if we need adaptive resolution
					adaptW, adaptH := engine.AdaptiveResolution(
						targetBytes, dur,
						probeInfo.Width, probeInfo.Height,
						probeInfo.Framerate,
					)
					if adaptW < probeInfo.Width || adaptH < probeInfo.Height {
						encOpts.ResolutionOverride = fmt.Sprintf("%dx%d", adaptW, adaptH)
					}
				}
			}
			ch, err = engine.EncodeTwoPass(ctx, encOpts)
		} else {
			ch, err = engine.Encode(ctx, encOpts)
		}

		if err != nil {
			cancel()
			return messages.EncodeErrorMsg{Err: err}
		}

		return encodingStartedMsg{
			cancel:     cancel,
			progressCh: ch,
			tempPath:   tempPath,
			outputPath: opts.Output,
		}
	}
}

type encodingStartedMsg struct {
	cancel     context.CancelFunc
	progressCh <-chan engine.ProgressUpdate
	tempPath   string
	outputPath string
}

// Init returns nil.
func (m EncodingModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the encoding screen.
func (m EncodingModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		pw := m.width - 10
		if pw < 20 {
			pw = 20
		}
		if pw > 60 {
			pw = 60
		}
		m.progress.SetWidth(pw)
		return m, nil

	case encodingStartedMsg:
		m.cancel = msg.cancel
		m.progressCh = msg.progressCh
		m.tempPath = msg.tempPath
		m.outputPath = msg.outputPath
		return m, readOneProgress(m.progressCh, m.tempPath, m.outputPath)

	case messages.EncodeProgressMsg:
		m.update = msg.Update

		if msg.Update.Error != nil {
			m.done = true
			m.errMsg = msg.Update.Error.Error()
			return m, nil
		}

		if msg.Update.Done {
			m.done = true
			return m, nil
		}

		// Add log line
		passLabel := ""
		if m.isTwoPass && msg.Update.Pass > 0 {
			passLabel = fmt.Sprintf("Pass %d/2 | ", msg.Update.Pass)
		}
		logLine := fmt.Sprintf("%s%.1f%% | Speed: %.1fx | FPS: %.0f | %s",
			passLabel, msg.Update.Percent, msg.Update.Speed, msg.Update.FPS, msg.Update.Bitrate)
		m.logLines = append(m.logLines, logLine)
		if len(m.logLines) > 5 {
			m.logLines = m.logLines[len(m.logLines)-5:]
		}

		// Chain: read the next progress update
		return m, readOneProgress(m.progressCh, m.tempPath, m.outputPath)

	case messages.EncodeCompleteMsg:
		m.done = true
		return m, nil

	case messages.EncodeErrorMsg:
		m.done = true
		m.errMsg = msg.Err.Error()
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "c":
			if !m.done && m.cancel != nil {
				m.cancel()
				m.done = true
				return m, func() tea.Msg { return messages.EncodeCancelMsg{} }
			}
		}

	default:
		// Update the progress bar model
		prog, cmd := m.progress.Update(msg)
		m.progress = prog
		return m, cmd
	}

	return m, nil
}

// readOneProgress reads a single update from the channel and returns a message.
func readOneProgress(ch <-chan engine.ProgressUpdate, tempPath, outputPath string) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return messages.EncodeErrorMsg{Err: fmt.Errorf("no progress channel")}
		}

		update, ok := <-ch
		if !ok {
			return messages.EncodeErrorMsg{Err: fmt.Errorf("progress channel closed unexpectedly")}
		}

		if update.Error != nil {
			return messages.EncodeErrorMsg{Err: update.Error}
		}

		if update.Done {
			// Move temp file to final output
			if err := os.Rename(tempPath, outputPath); err != nil {
				return messages.EncodeErrorMsg{Err: fmt.Errorf("failed to move output: %w", err)}
			}

			var outputSize int64
			if stat, err := os.Stat(outputPath); err == nil {
				outputSize = stat.Size()
			}

			return messages.EncodeCompleteMsg{
				OutputPath: outputPath,
				OutputSize: outputSize,
			}
		}

		return messages.EncodeProgressMsg{Update: update}
	}
}

// View renders the encoding progress screen.
func (m EncodingModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Encoding")
	b.WriteString(title)
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(style.ErrorStyle().Render("Error: " + m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(style.KeyHintStyle().Render("[Esc] Back"))
		return b.String()
	}

	// Two-pass indicator
	if m.isTwoPass && m.update.Pass > 0 {
		passLabel := fmt.Sprintf("Pass %d/2", m.update.Pass)
		b.WriteString(style.BadgeStyle().Render(passLabel))
		b.WriteString("\n\n")
	}

	// Progress bar
	pct := m.update.Percent / 100.0
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}
	b.WriteString(m.progress.ViewAs(pct))
	b.WriteString("\n\n")

	// Stat boxes
	elapsed := time.Since(m.startTime)
	isNarrow := m.width < 60

	if isNarrow {
		fmt.Fprintf(&b, "  Percent: %.1f%%\n", m.update.Percent)
		if m.isTwoPass && m.update.Pass > 0 {
			b.WriteString(fmt.Sprintf("  Pass:    %d/2\n", m.update.Pass))
		}
		b.WriteString(fmt.Sprintf("  Speed:   %.1fx\n", m.update.Speed))
		b.WriteString(fmt.Sprintf("  FPS:     %.0f\n", m.update.FPS))
		b.WriteString(fmt.Sprintf("  Bitrate: %s\n", m.update.Bitrate))
		b.WriteString(fmt.Sprintf("  Size:    %s\n", formatBytes(m.update.Size)))
		b.WriteString(fmt.Sprintf("  ETA:     %s\n", formatDurationShort(m.update.ETA)))
	} else {
		boxes := []string{
			style.StatBoxStyle().Render(fmt.Sprintf("Percent\n%.1f%%", m.update.Percent)),
			style.StatBoxStyle().Render(fmt.Sprintf("Speed\n%.1fx", m.update.Speed)),
			style.StatBoxStyle().Render(fmt.Sprintf("FPS\n%.0f", m.update.FPS)),
			style.StatBoxStyle().Render(fmt.Sprintf("Bitrate\n%s", m.update.Bitrate)),
			style.StatBoxStyle().Render(fmt.Sprintf("Size\n%s", formatBytes(m.update.Size))),
		}

		if m.isTwoPass && m.update.Pass > 0 {
			boxes = append(boxes, style.StatBoxStyle().Render(fmt.Sprintf("Pass\n%d/2", m.update.Pass)))
		}

		boxes = append(boxes, style.StatBoxStyle().Render(fmt.Sprintf("ETA\n%s", formatDurationShort(m.update.ETA))))

		if m.width >= 80 {
			b.WriteString(strings.Join(boxes, " "))
		} else {
			mid := len(boxes) / 2
			b.WriteString(strings.Join(boxes[:mid], " "))
			b.WriteString("\n")
			b.WriteString(strings.Join(boxes[mid:], " "))
		}
	}

	b.WriteString("\n\n")

	b.WriteString(style.MutedStyle().Render(fmt.Sprintf("Elapsed: %s", formatDurationShort(elapsed))))
	b.WriteString("\n\n")

	// Log panel
	if len(m.logLines) > 0 {
		b.WriteString(style.MutedStyle().Render("Log:"))
		b.WriteString("\n")
		for _, line := range m.logLines {
			b.WriteString("  " + style.MutedStyle().Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.done {
		b.WriteString(style.KeyHintStyle().Render("[Esc] Back"))
	} else {
		b.WriteString(style.KeyHintStyle().Render("[Esc/c] Cancel"))
	}

	return b.String()
}
