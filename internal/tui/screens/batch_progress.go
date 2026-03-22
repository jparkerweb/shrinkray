package screens

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/progress"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// BatchProgressModel is the model for the batch progress screen (Screen 9b).
type BatchProgressModel struct {
	queue      *engine.JobQueue
	overallBar progress.Model
	startTime  time.Time
	cancel     context.CancelFunc
	eventCh    <-chan engine.BatchEvent
	done       bool
	cancelled  bool
	stats      engine.QueueStats
	currentJob string // ID of the currently encoding job
	lastUpdate engine.ProgressUpdate
	width      int
	height     int
	scrollTop  int

	// Options for encoding
	preset    presets.Preset
	hwEncoder string
	outOpts   engine.OutputOptions
	skipOpts  engine.SkipOptions
	jobs      int
	maxRetries int
}

// NewBatchProgressModel creates a new batch progress model.
func NewBatchProgressModel() BatchProgressModel {
	bar := progress.New(
		progress.WithDefaultBlend(),
		progress.WithWidth(50),
	)
	return BatchProgressModel{
		overallBar: bar,
		jobs:       1,
		maxRetries: 2,
	}
}

// SetOptions configures the batch encoding parameters.
func (m *BatchProgressModel) SetOptions(preset presets.Preset, hwEncoder string, outOpts engine.OutputOptions, skipOpts engine.SkipOptions, jobs int, maxRetries int) {
	m.preset = preset
	m.hwEncoder = hwEncoder
	m.outOpts = outOpts
	m.skipOpts = skipOpts
	m.jobs = jobs
	m.maxRetries = maxRetries
}

// StartBatch begins batch encoding and returns a command that feeds events.
func (m *BatchProgressModel) StartBatch(queue *engine.JobQueue) tea.Cmd {
	m.queue = queue
	m.done = false
	m.cancelled = false
	m.startTime = time.Now()
	m.currentJob = ""
	m.lastUpdate = engine.ProgressUpdate{}

	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())

		opts := engine.BatchOptions{
			Jobs:       m.jobs,
			Preset:     m.preset,
			HWEncoder:  m.hwEncoder,
			OutputOpts: m.outOpts,
			SkipOpts:   m.skipOpts,
			MaxRetries: m.maxRetries,
		}

		eventCh := engine.RunBatch(ctx, queue, opts)

		return batchStartedMsg{
			cancel:  cancel,
			eventCh: eventCh,
		}
	}
}

type batchStartedMsg struct {
	cancel  context.CancelFunc
	eventCh <-chan engine.BatchEvent
}

// Init returns nil.
func (m BatchProgressModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the batch progress screen.
func (m BatchProgressModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
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
		m.overallBar.SetWidth(pw)
		return m, nil

	case batchStartedMsg:
		m.cancel = msg.cancel
		m.eventCh = msg.eventCh
		return m, readOneBatchEvent(m.eventCh)

	case messages.BatchStartMsg:
		m.queue = msg.Queue
		return m, m.StartBatch(msg.Queue)

	case messages.BatchEventMsg:
		return m.handleBatchEvent(msg.Event)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.done {
				return m, func() tea.Msg {
					return messages.NavigateMsg{Screen: messages.ScreenBatchComplete}
				}
			}
			// Cancel prompt: for simplicity, cancel entire batch
			if m.cancel != nil && !m.cancelled {
				m.cancelled = true
				m.cancel()
			}
			return m, nil
		}

	default:
		prog, cmd := m.overallBar.Update(msg)
		m.overallBar = prog
		return m, cmd
	}

	return m, nil
}

func (m BatchProgressModel) handleBatchEvent(event engine.BatchEvent) (style.ScreenModel, tea.Cmd) {
	switch e := event.(type) {
	case engine.JobStartedEvent:
		m.currentJob = e.JobID
		m.lastUpdate = engine.ProgressUpdate{}

	case engine.JobProgressEvent:
		m.lastUpdate = e.Update
		if m.queue != nil {
			m.stats = m.queue.Stats()
		}

	case engine.JobCompleteEvent:
		if m.queue != nil {
			m.stats = m.queue.Stats()
		}

	case engine.JobFailedEvent:
		if m.queue != nil {
			m.stats = m.queue.Stats()
		}

	case engine.JobSkippedEvent:
		if m.queue != nil {
			m.stats = m.queue.Stats()
		}

	case engine.BatchCompleteEvent:
		m.done = true
		m.stats = e.Stats
		return m, func() tea.Msg {
			return messages.NavigateMsg{Screen: messages.ScreenBatchComplete}
		}
	}

	// Continue reading events
	return m, readOneBatchEvent(m.eventCh)
}

func readOneBatchEvent(ch <-chan engine.BatchEvent) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		event, ok := <-ch
		if !ok {
			return nil
		}
		return messages.BatchEventMsg{Event: event}
	}
}

// View renders the batch progress screen.
func (m BatchProgressModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Batch Encoding")
	b.WriteString(title)
	b.WriteString("\n")

	if m.queue == nil {
		b.WriteString(style.MutedStyle().Render("Preparing..."))
		return b.String()
	}

	stats := m.stats
	if stats.Total == 0 {
		stats = m.queue.Stats()
	}

	// Overall progress
	processed := stats.Complete + stats.Failed + stats.Skipped
	var overallPct float64
	if stats.Total > 0 {
		overallPct = float64(processed) / float64(stats.Total)
	}
	b.WriteString(style.SubtitleStyle().Render(fmt.Sprintf(
		"Overall: %d/%d files (%.0f%%)", processed, stats.Total, overallPct*100)))
	b.WriteString("\n")
	b.WriteString(m.overallBar.ViewAs(overallPct))
	b.WriteString("\n\n")

	// Per-file rows
	jobs := m.queue.All()
	visibleRows := m.height - 16
	if visibleRows < 3 {
		visibleRows = 3
	}
	endIdx := m.scrollTop + visibleRows
	if endIdx > len(jobs) {
		endIdx = len(jobs)
	}

	nameW := m.width - 40
	if nameW < 15 {
		nameW = 15
	}

	for i := m.scrollTop; i < endIdx; i++ {
		job := jobs[i]
		name := filepath.Base(job.InputPath)
		if len(name) > nameW {
			name = name[:nameW-3] + "..."
		}

		var statusBadge string
		switch job.Status {
		case engine.JobStatusPending:
			statusBadge = style.MutedStyle().Render("pending")
		case engine.JobStatusEncoding:
			pctStr := fmt.Sprintf("%.0f%%", job.Progress)
			statusBadge = style.AccentStyle().Render(pctStr)
		case engine.JobStatusComplete:
			statusBadge = style.SuccessStyle().Render("done")
		case engine.JobStatusFailed:
			statusBadge = style.ErrorStyle().Render("failed")
		case engine.JobStatusSkipped:
			statusBadge = style.WarningStyle().Render("skip")
		}

		outputInfo := ""
		if job.OutputSize > 0 {
			outputInfo = formatBytes(job.OutputSize)
		}

		row := fmt.Sprintf("  %-*s %8s %10s", nameW, name, statusBadge, outputInfo)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Current file detail
	b.WriteString("\n")
	if m.currentJob != "" && !m.done {
		elapsed := time.Since(m.startTime)

		if m.width >= 60 {
			boxes := []string{
				style.StatBoxStyle().Render(fmt.Sprintf("Percent\n%.1f%%", m.lastUpdate.Percent)),
				style.StatBoxStyle().Render(fmt.Sprintf("Speed\n%.1fx", m.lastUpdate.Speed)),
				style.StatBoxStyle().Render(fmt.Sprintf("FPS\n%.0f", m.lastUpdate.FPS)),
				style.StatBoxStyle().Render(fmt.Sprintf("ETA\n%s", formatDurationShort(m.lastUpdate.ETA))),
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
		} else {
			fmt.Fprintf(&b, "  Progress: %.1f%% | Speed: %.1fx | FPS: %.0f\n",
				m.lastUpdate.Percent, m.lastUpdate.Speed, m.lastUpdate.FPS)
		}
		b.WriteString("\n")
		b.WriteString(style.MutedStyle().Render(fmt.Sprintf("Elapsed: %s", formatDurationShort(elapsed))))
	}

	b.WriteString("\n\n")
	if m.done {
		b.WriteString(style.KeyHintStyle().Render("[Enter] View results"))
	} else if m.cancelled {
		b.WriteString(style.WarningStyle().Render("Cancelling..."))
	} else {
		b.WriteString(style.KeyHintStyle().Render("[Esc] Cancel batch"))
	}

	return b.String()
}
