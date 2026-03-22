package screens

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/filepicker"

	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// videoExtensions lists the allowed video file extensions.
var videoExtensions = []string{
	".mp4", ".mkv", ".webm", ".avi", ".mov", ".wmv", ".flv", ".ts", ".m4v",
}

// FilePickerModel is the model for the file picker screen.
type FilePickerModel struct {
	picker       filepicker.Model
	textInput    string
	useText      bool // true when text input is focused
	selected     string
	selectedMany map[string]bool // multi-select: path -> selected
	errMsg       string
	width        int
	height       int
	recursive    bool // whether to recurse into directories
}

// NewFilePickerModel creates a new file picker model.
func NewFilePickerModel() FilePickerModel {
	fp := filepicker.New()
	fp.AllowedTypes = videoExtensions
	fp.SetHeight(15)

	// Start in current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	fp.CurrentDirectory = cwd

	return FilePickerModel{
		picker:       fp,
		selectedMany: make(map[string]bool),
	}
}

// SetRecursive enables or disables recursive directory walking.
func (m *FilePickerModel) SetRecursive(r bool) {
	m.recursive = r
}

// Init initializes the file picker.
func (m FilePickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

// Update handles messages for the file picker screen.
func (m FilePickerModel) Update(msg tea.Msg) (style.ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := m.height - 10
		if h < 5 {
			h = 5
		}
		m.picker.SetHeight(h)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			m.useText = !m.useText
			m.errMsg = ""
			return m, nil

		case "esc":
			return m, func() tea.Msg { return messages.BackMsg{} }
		}

		if m.useText {
			return m.handleTextInput(msg)
		}

		// "b" key: proceed to batch queue with selected files
		if msg.String() == "b" && len(m.selectedMany) > 0 {
			paths := m.selectedPaths()
			if len(paths) == 1 {
				path := paths[0]
				m.selected = path
				return m, tea.Batch(
					func() tea.Msg { return messages.FileSelectedMsg{Path: path} },
					probeFile(path),
				)
			}
			return m, func() tea.Msg {
				return messages.FilesSelectedMsg{Paths: paths}
			}
		}

		// "x" key: clear multi-selection
		if msg.String() == "x" {
			m.selectedMany = make(map[string]bool)
			m.errMsg = ""
			return m, nil
		}
	}

	if !m.useText {
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)

		// Check if file was selected via Enter
		if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
			// If user is building a multi-select (has files already),
			// add to batch and stay on picker
			if len(m.selectedMany) > 0 {
				m.selectedMany[path] = true
				m.errMsg = ""
				return m, nil
			}
			// Single file selection — proceed to info screen
			m.selected = path
			m.errMsg = ""
			return m, tea.Batch(
				func() tea.Msg { return messages.FileSelectedMsg{Path: path} },
				probeFile(path),
			)
		}

		// Check if file was disabled (wrong type)
		if didSelect, path := m.picker.DidSelectDisabledFile(msg); didSelect {
			m.errMsg = fmt.Sprintf("Cannot select %s - not a supported video file", filepath.Base(path))
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

func (m FilePickerModel) selectedPaths() []string {
	paths := make([]string, 0, len(m.selectedMany))
	for p := range m.selectedMany {
		paths = append(paths, p)
	}
	return paths
}

func (m FilePickerModel) handleTextInput(msg tea.KeyPressMsg) (style.ScreenModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		path := strings.TrimSpace(m.textInput)
		if path == "" {
			m.errMsg = "Please enter a file path"
			return m, nil
		}

		// Support comma-separated paths
		rawPaths := strings.Split(path, ",")
		var validPaths []string

		for _, p := range rawPaths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			stat, err := os.Stat(p)
			if err != nil {
				m.errMsg = fmt.Sprintf("File not found: %s", p)
				return m, nil
			}

			if stat.IsDir() {
				// Walk directory for video files
				found := discoverVideoFiles(p, m.recursive)
				if len(found) == 0 {
					m.errMsg = fmt.Sprintf("No video files found in: %s", p)
					return m, nil
				}
				validPaths = append(validPaths, found...)
			} else {
				// Check extension
				ext := strings.ToLower(filepath.Ext(p))
				valid := false
				for _, ve := range videoExtensions {
					if ext == ve {
						valid = true
						break
					}
				}
				if !valid {
					m.errMsg = fmt.Sprintf("Unsupported format: %s", ext)
					return m, nil
				}
				validPaths = append(validPaths, p)
			}
		}

		if len(validPaths) == 0 {
			m.errMsg = "No valid video files found"
			return m, nil
		}

		if len(validPaths) == 1 {
			m.selected = validPaths[0]
			m.errMsg = ""
			return m, tea.Batch(
				func() tea.Msg { return messages.FileSelectedMsg{Path: validPaths[0]} },
				probeFile(validPaths[0]),
			)
		}

		// Multiple files — go to batch
		m.errMsg = ""
		return m, func() tea.Msg {
			return messages.FilesSelectedMsg{Paths: validPaths}
		}

	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
		return m, nil

	default:
		s := msg.String()
		if len(s) == 1 && s[0] >= 32 {
			m.textInput += s
		}
		return m, nil
	}
}

// discoverVideoFiles walks a directory and returns all video file paths.
func discoverVideoFiles(dir string, recursive bool) []string {
	var files []string

	if recursive {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if isVideoFile(path) {
				files = append(files, path)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name())
			if isVideoFile(path) {
				files = append(files, path)
			}
		}
	}

	return files
}

// isVideoFile checks if a file path has a supported video extension.
func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, ve := range videoExtensions {
		if ext == ve {
			return true
		}
	}
	return false
}

// DiscoverVideoFiles is the exported version for use by CLI code.
func DiscoverVideoFiles(dir string, recursive bool) []string {
	return discoverVideoFiles(dir, recursive)
}

func probeFile(path string) tea.Cmd {
	return func() tea.Msg {
		info, err := engine.Probe(context.Background(), path)
		return messages.VideoProbeCompleteMsg{Info: info, Err: err}
	}
}

// View renders the file picker screen.
func (m FilePickerModel) View() string {
	var b strings.Builder

	title := style.TitleStyle().Render("Select Video File(s)")
	b.WriteString(title)
	b.WriteString("\n")

	if m.useText {
		b.WriteString(style.SubtitleStyle().Render("Enter file path (comma-separated for batch, or directory):"))
		b.WriteString("\n\n")

		w := m.width - 6
		if w < 20 {
			w = 20
		}
		inputStyle := style.CardStyle().Width(w)
		prompt := "> " + m.textInput + "_"
		b.WriteString(inputStyle.Render(prompt))
		b.WriteString("\n")
	} else {
		b.WriteString(m.picker.View())
		b.WriteString("\n")
	}

	// Error message
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(style.ErrorStyle().Render(m.errMsg))
		b.WriteString("\n")
	}

	// Multi-select count
	if len(m.selectedMany) > 0 {
		b.WriteString("\n")
		b.WriteString(style.AccentStyle().Render(fmt.Sprintf("%d file(s) selected for batch", len(m.selectedMany))))
		b.WriteString("\n")
	}

	// Selected file info (single-select)
	if m.selected != "" && len(m.selectedMany) == 0 {
		b.WriteString("\n")
		stat, err := os.Stat(m.selected)
		if err == nil {
			info := fmt.Sprintf("Selected: %s (%s)",
				filepath.Base(m.selected),
				formatBytes(stat.Size()),
			)
			b.WriteString(style.SuccessStyle().Render(info))
		}
		b.WriteString("\n")
	}

	// Mode toggle hint
	b.WriteString("\n")
	if m.useText {
		b.WriteString(style.KeyHintStyle().Render("[Tab] Switch to browser  [Enter] Confirm  [Esc] Back"))
	} else {
		hints := "[Tab] Path input  [Enter] Select  [Space] Multi-select  [Esc] Back"
		b.WriteString(style.KeyHintStyle().Render(hints))
	}

	return b.String()
}
