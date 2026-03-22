package screens

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// BatchCompleteModel is the model for the batch completion screen (Screen 9c).
type BatchCompleteModel struct {
	queue     *engine.JobQueue
	elapsed   time.Duration
	width     int
	height    int
	scrollTop int
	cursor    int
}

// NewBatchCompleteModel creates a new batch completion model.
func NewBatchCompleteModel() BatchCompleteModel {
	return BatchCompleteModel{}
}

// SetResults sets the queue and elapsed time for display.
func (m *BatchCompleteModel) SetResults(queue *engine.JobQueue, elapsed time.Duration) {
	m.queue = queue
	m.elapsed = elapsed
}

// Init returns nil.
func (m BatchCompleteModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the batch completion screen.
func (m BatchCompleteModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
			return m, nil

		case "down", "j":
			qLen := 0
			if m.queue != nil {
				qLen = m.queue.Len()
			}
			if m.cursor < qLen-1 {
				m.cursor++
				m.adjustScroll()
			}
			return m, nil

		case "r":
			// Retry failed jobs
			if m.queue != nil {
				failed := m.queue.ByStatus(engine.JobStatusFailed)
				if len(failed) > 0 {
					for _, job := range failed {
						m.queue.Update(job.ID, func(j *engine.Job) {
							j.Status = engine.JobStatusPending
							j.Progress = 0
							j.Error = ""
						})
					}
					return m, func() tea.Msg {
						return messages.BatchStartMsg{Queue: m.queue}
					}
				}
			}
			return m, nil

		case "n":
			// New batch
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenFilePicker}
			}

		case "o":
			// Open output folder of first completed job
			if m.queue != nil {
				completed := m.queue.ByStatus(engine.JobStatusComplete)
				if len(completed) > 0 {
					outputPath := completed[0].OutputPath
					if outputPath != "" {
						return m, func() tea.Msg {
							_ = engine.OpenFolder(outputPath)
							return nil
						}
					}
				}
			}
			return m, nil

		case "q":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		}
	}

	return m, nil
}

func (m *BatchCompleteModel) adjustScroll() {
	visibleRows := m.visibleRows()
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.cursor >= m.scrollTop+visibleRows {
		m.scrollTop = m.cursor - visibleRows + 1
	}
}

func (m BatchCompleteModel) visibleRows() int {
	rows := m.height - 18
	if rows < 3 {
		rows = 3
	}
	return rows
}

// View renders the batch completion screen.
func (m BatchCompleteModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Batch Complete!")
	b.WriteString(title)
	b.WriteString("\n")

	if m.queue == nil {
		b.WriteString(style.MutedStyle().Render("No results."))
		return b.String()
	}

	stats := m.queue.Stats()

	b.WriteString(style.SuccessStyle().Render(fmt.Sprintf(
		"Processed %d files in %s", stats.Total, formatDurationShort(m.elapsed))))
	b.WriteString("\n\n")

	// Results table
	nameW := m.width - 60
	if nameW < 15 {
		nameW = 15
	}

	header := fmt.Sprintf("  %-*s %10s %10s %8s %8s",
		nameW, "Filename", "Input", "Output", "Savings", "Status")
	b.WriteString(style.AccentStyle().Render(header))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", m.width-4))
	b.WriteString("\n")

	jobs := m.queue.All()
	visibleRows := m.visibleRows()
	endIdx := m.scrollTop + visibleRows
	if endIdx > len(jobs) {
		endIdx = len(jobs)
	}

	for i := m.scrollTop; i < endIdx; i++ {
		job := jobs[i]
		name := filepath.Base(job.InputPath)
		if len(name) > nameW {
			name = name[:nameW-3] + "..."
		}

		inputStr := formatBytes(job.InputSize)
		outputStr := ""
		savingsStr := ""

		var statusStr string
		switch job.Status {
		case engine.JobStatusComplete:
			statusStr = style.SuccessStyle().Render("done")
			outputStr = formatBytes(job.OutputSize)
			if job.InputSize > 0 && job.OutputSize > 0 {
				savings := float64(job.InputSize-job.OutputSize) / float64(job.InputSize) * 100
				savingsStr = fmt.Sprintf("%.1f%%", savings)
			}
		case engine.JobStatusFailed:
			statusStr = style.ErrorStyle().Render("FAIL")
		case engine.JobStatusSkipped:
			statusStr = style.WarningStyle().Render("skip")
		default:
			statusStr = string(job.Status)
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		row := fmt.Sprintf("%s%-*s %10s %10s %8s %8s",
			cursor, nameW, name, inputStr, outputStr, savingsStr, statusStr)
		b.WriteString(row)
		b.WriteString("\n")

		// Show error for failed jobs when selected
		if i == m.cursor && job.Status == engine.JobStatusFailed && job.Error != "" {
			errLine := fmt.Sprintf("    Error: %s", job.Error)
			if len(errLine) > m.width-4 {
				errLine = errLine[:m.width-7] + "..."
			}
			b.WriteString(style.ErrorStyle().Render(errLine))
			b.WriteString("\n")
		}
	}

	// Totals
	b.WriteString("\n")
	totalSavings := ""
	if stats.TotalInputSize > 0 && stats.TotalOutputSize > 0 {
		savings := float64(stats.TotalInputSize-stats.TotalOutputSize) / float64(stats.TotalInputSize) * 100
		savedBytes := stats.TotalInputSize - stats.TotalOutputSize
		totalSavings = fmt.Sprintf(" | Saved: %s (%.1f%%)", formatBytes(savedBytes), savings)
	}

	b.WriteString(style.ValueStyle().Render(fmt.Sprintf(
		"Total Input: %s | Output: %s%s",
		formatBytes(stats.TotalInputSize),
		formatBytes(stats.TotalOutputSize),
		totalSavings,
	)))
	b.WriteString("\n")

	b.WriteString(style.MutedStyle().Render(fmt.Sprintf(
		"Complete: %d | Failed: %d | Skipped: %d | Time: %s",
		stats.Complete, stats.Failed, stats.Skipped, formatDurationShort(m.elapsed),
	)))
	b.WriteString("\n\n")

	// Hints
	hints := "[n] New batch  [q] Quit"
	if stats.Failed > 0 {
		hints = "[r] Retry failed  " + hints
	}
	hints = "[o] Open folder  " + hints
	b.WriteString(style.KeyHintStyle().Render(hints))

	return b.String()
}
