package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// CompleteModel is the model for the completion screen.
type CompleteModel struct {
	inputPath   string
	outputPath  string
	inputSize   int64
	outputSize  int64
	elapsedTime time.Duration
	replaced    bool   // true after "delete original" was performed
	errMsg      string // error from replace operation
	width       int
	height      int
}

// NewCompleteModel creates a new completion model.
func NewCompleteModel() CompleteModel {
	return CompleteModel{}
}

// SetResults sets the encoding results for display.
func (m *CompleteModel) SetResults(inputPath, outputPath string, inputSize, outputSize int64, elapsed time.Duration) {
	m.inputPath = inputPath
	m.outputPath = outputPath
	m.inputSize = inputSize
	m.outputSize = outputSize
	m.elapsedTime = elapsed
	m.replaced = false
}

// Init returns nil.
func (m CompleteModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the completion screen.
func (m CompleteModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case replaceResultMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.replaced = true
			m.outputPath = msg.newPath
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "o":
			if m.outputPath != "" {
				return m, func() tea.Msg {
					_ = engine.OpenFolder(m.outputPath)
					return nil
				}
			}
			return m, nil
		case "d":
			if m.inputPath != "" && m.outputPath != "" && !m.replaced {
				inputPath := m.inputPath
				outputPath := m.outputPath
				return m, func() tea.Msg {
					return replaceOriginal(inputPath, outputPath)
				}
			}
			return m, nil
		case "r":
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenPresets}
			}
		case "n":
			return m, func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenFilePicker}
			}
		case "q":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		}
	}

	return m, nil
}

// replaceResultMsg is the result of deleting the original and renaming.
type replaceResultMsg struct {
	newPath string
	err     error
}

// replaceOriginal deletes the original file and renames the output to the
// original's filename (preserving the output's extension if it differs).
func replaceOriginal(inputPath, outputPath string) replaceResultMsg {
	// Build the target path: original's directory + original's base name + output's extension
	inputDir := filepath.Dir(inputPath)
	inputBase := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputExt := filepath.Ext(outputPath)
	targetPath := filepath.Join(inputDir, inputBase+outputExt)

	// Delete the original
	if err := os.Remove(inputPath); err != nil {
		return replaceResultMsg{err: fmt.Errorf("failed to delete original: %w", err)}
	}

	// Rename output to the target path
	if err := os.Rename(outputPath, targetPath); err != nil {
		return replaceResultMsg{err: fmt.Errorf("deleted original but failed to rename output: %w", err)}
	}

	return replaceResultMsg{newPath: targetPath}
}

// View renders the completion screen.
func (m CompleteModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Encoding Complete!")
	b.WriteString(title)
	b.WriteString("\n")

	b.WriteString(style.SuccessStyle().Render("Done!"))
	b.WriteString("\n\n")

	// Bar chart comparing sizes
	maxBarWidth := m.width - 20
	if maxBarWidth > 50 {
		maxBarWidth = 50
	}
	if maxBarWidth < 10 {
		maxBarWidth = 10
	}

	inputLabel := fmt.Sprintf("Input:  %s", formatBytes(m.inputSize))
	outputLabel := fmt.Sprintf("Output: %s", formatBytes(m.outputSize))

	var inputBarLen, outputBarLen int
	if m.inputSize > 0 {
		inputBarLen = maxBarWidth
		outputBarLen = int(float64(maxBarWidth) * float64(m.outputSize) / float64(m.inputSize))
		if outputBarLen < 1 && m.outputSize > 0 {
			outputBarLen = 1
		}
	}

	inputBar := strings.Repeat("=", inputBarLen)
	outputBar := strings.Repeat("=", outputBarLen)

	b.WriteString(style.LabelStyle().Width(0).Render(inputLabel))
	b.WriteString("\n")
	b.WriteString(style.WarningStyle().Render(inputBar))
	b.WriteString("\n\n")

	b.WriteString(style.LabelStyle().Width(0).Render(outputLabel))
	b.WriteString("\n")
	b.WriteString(style.SuccessStyle().Render(outputBar))
	b.WriteString("\n\n")

	// Savings
	if m.inputSize > 0 && m.outputSize > 0 {
		savings := float64(m.inputSize-m.outputSize) / float64(m.inputSize) * 100
		ratio := float64(m.inputSize) / float64(m.outputSize)
		savedBytes := m.inputSize - m.outputSize

		b.WriteString(style.ValueStyle().Render(fmt.Sprintf("Saved: %s (%.1f%%)", formatBytes(savedBytes), savings)))
		b.WriteString("\n")
		b.WriteString(style.MutedStyle().Render(fmt.Sprintf("Compression ratio: %.1fx", ratio)))
		b.WriteString("\n")
	}

	// Output path
	b.WriteString("\n")
	b.WriteString(style.MutedStyle().Render("Output: "))
	b.WriteString(m.outputPath)
	b.WriteString("\n")

	// Elapsed time
	b.WriteString(style.MutedStyle().Render(fmt.Sprintf("Time: %s", formatDurationShort(m.elapsedTime))))
	b.WriteString("\n")

	// Replace status
	if m.replaced {
		b.WriteString("\n")
		b.WriteString(style.SuccessStyle().Render("Original deleted and output renamed."))
		b.WriteString("\n")
	}
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(style.ErrorStyle().Render("Error: " + m.errMsg))
		b.WriteString("\n")
	}

	return b.String()
}
