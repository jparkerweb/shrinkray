package style

import tea "charm.land/bubbletea/v2"

// ScreenModel is the interface that all TUI screen sub-models implement.
// It is similar to tea.Model but View() returns a plain string instead of
// tea.View, since only the top-level App model manages AltScreen and mouse mode.
type ScreenModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (ScreenModel, tea.Cmd)
	View() string
}
