package screens

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// BatchQueueModel is the model for the batch queue overview screen (Screen 9a).
type BatchQueueModel struct {
	queue       *engine.JobQueue
	cursor      int
	preset      *presets.Preset
	width       int
	height      int
	scrollTop   int
	errMsg      string
}

// NewBatchQueueModel creates a new batch queue model.
func NewBatchQueueModel() BatchQueueModel {
	return BatchQueueModel{}
}

// SetQueue sets the job queue for display.
func (m *BatchQueueModel) SetQueue(queue *engine.JobQueue) {
	m.queue = queue
}

// SetPreset sets the active preset.
func (m *BatchQueueModel) SetPreset(p *presets.Preset) {
	m.preset = p
}

// Init returns nil.
func (m BatchQueueModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the batch queue screen.
func (m BatchQueueModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case messages.BatchQueueReadyMsg:
		m.queue = msg.Queue
		return m, nil

	case tea.KeyPressMsg:
		if m.queue == nil {
			break
		}
		qLen := m.queue.Len()

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
			return m, nil

		case "down", "j":
			if m.cursor < qLen-1 {
				m.cursor++
				m.adjustScroll()
			}
			return m, nil

		case "shift+up", "K":
			// Move current job up
			if m.cursor > 0 {
				m.queue.Swap(m.cursor, m.cursor-1)
				m.cursor--
				m.adjustScroll()
			}
			return m, nil

		case "shift+down", "J":
			// Move current job down
			if m.cursor < qLen-1 {
				m.queue.Swap(m.cursor, m.cursor+1)
				m.cursor++
				m.adjustScroll()
			}
			return m, nil

		case "d", "delete":
			// Remove current job
			if qLen > 0 {
				if job, ok := m.queue.Get(m.cursor); ok {
					m.queue.Remove(job.ID)
					if m.cursor >= m.queue.Len() && m.cursor > 0 {
						m.cursor--
					}
				}
			}
			return m, nil

		case "p":
			// Change preset — for now just cycle through presets
			// A full modal picker would be ideal but is complex
			return m, nil

		case "enter":
			// Start batch encoding
			if m.queue != nil && m.queue.Len() > 0 {
				return m, func() tea.Msg {
					return messages.BatchStartMsg{Queue: m.queue}
				}
			}
			return m, nil

		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		}
	}

	return m, nil
}

func (m *BatchQueueModel) adjustScroll() {
	visibleRows := m.visibleRows()
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.cursor >= m.scrollTop+visibleRows {
		m.scrollTop = m.cursor - visibleRows + 1
	}
}

func (m BatchQueueModel) visibleRows() int {
	rows := m.height - 14 // header, totals, hints
	if rows < 3 {
		rows = 3
	}
	return rows
}

// View renders the batch queue screen.
func (m BatchQueueModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Batch Queue")
	b.WriteString(title)
	b.WriteString("\n")

	if m.queue == nil || m.queue.Len() == 0 {
		b.WriteString(style.MutedStyle().Render("No files in queue."))
		b.WriteString("\n\n")
		b.WriteString(style.KeyHintStyle().Render("[Esc] Back"))
		return b.String()
	}

	presetName := "balanced"
	if m.preset != nil {
		presetName = m.preset.Name
	}
	b.WriteString(style.SubtitleStyle().Render(fmt.Sprintf("Preset: %s", presetName)))
	b.WriteString("\n\n")

	// Table header
	nameW := m.width - 50
	if nameW < 20 {
		nameW = 20
	}
	header := fmt.Sprintf("  %-*s %10s %12s %8s",
		nameW, "Filename", "Size", "Est. Output", "Status")
	b.WriteString(style.AccentStyle().Render(header))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", m.width-4))
	b.WriteString("\n")

	// Table rows
	jobs := m.queue.All()
	visibleRows := m.visibleRows()
	endIdx := m.scrollTop + visibleRows
	if endIdx > len(jobs) {
		endIdx = len(jobs)
	}

	for i := m.scrollTop; i < endIdx; i++ {
		job := jobs[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		name := filepath.Base(job.InputPath)
		if len(name) > nameW {
			name = name[:nameW-3] + "..."
		}

		estOutput := "..."
		if m.preset != nil {
			// Quick estimate
			estSize := estimateJobSize(job, *m.preset)
			if estSize > 0 {
				estOutput = formatBytes(estSize)
			}
		}

		statusStr := string(job.Status)

		row := fmt.Sprintf("%s%-*s %10s %12s %8s",
			cursor, nameW, name, formatBytes(job.InputSize), estOutput, statusStr)

		if i == m.cursor {
			b.WriteString(style.ValueStyle().Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	if len(jobs) > visibleRows {
		b.WriteString(style.MutedStyle().Render(
			fmt.Sprintf("  ... showing %d-%d of %d files", m.scrollTop+1, endIdx, len(jobs))))
		b.WriteString("\n")
	}

	// Totals
	b.WriteString("\n")
	stats := m.queue.Stats()
	b.WriteString(style.ValueStyle().Render(fmt.Sprintf(
		"Total: %d files | Input: %s | Pending: %d",
		stats.Total, formatBytes(stats.TotalInputSize), stats.Pending,
	)))
	b.WriteString("\n\n")

	// Hints
	b.WriteString(style.KeyHintStyle().Render(
		"[Enter] Start  [d] Remove  [Shift+Up/Down] Reorder  [Esc] Back"))

	return b.String()
}

// estimateJobSize provides a rough estimate for a single job.
func estimateJobSize(job engine.Job, preset presets.Preset) int64 {
	if job.InputSize <= 0 {
		return 0
	}
	// Quick probe for estimate (only if we have duration)
	if job.Duration > 0 {
		info := &engine.VideoInfo{
			Size:     job.InputSize,
			Duration: job.Duration,
			Width:    1920, // default assumption
			Height:   1080,
		}
		return engine.EstimateSize(info, preset)
	}
	return 0
}

// ProbeAndAddFiles probes each file and adds it to the queue.
func ProbeAndAddFiles(paths []string, presetKey string) tea.Cmd {
	return func() tea.Msg {
		queue := engine.NewJobQueue()
		for _, path := range paths {
			stat, err := os.Stat(path)
			if err != nil {
				continue
			}

			var duration float64
			info, err := engine.Probe(context.Background(), path)
			if err == nil {
				duration = info.Duration
			}

			queue.Add(engine.Job{
				InputPath: path,
				PresetKey: presetKey,
				Status:    engine.JobStatusPending,
				InputSize: stat.Size(),
				Duration:  duration,
			})
		}
		return messages.BatchQueueReadyMsg{Queue: queue}
	}
}
