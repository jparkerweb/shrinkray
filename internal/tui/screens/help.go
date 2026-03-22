package screens

// Help overlay rendering is handled in the top-level app.go, which composites
// the help content on top of the current screen's view. This file provides
// the data structures and helpers used by the help overlay.
//
// The help overlay is triggered by the '?' key from any screen and shows:
// - Global key bindings (quit, back, help, theme toggle)
// - Context-sensitive key bindings for the current screen
//
// It is dismissed by pressing '?', Esc, or any navigation key.

// HelpBinding represents a single key binding entry for the help overlay.
type HelpBinding struct {
	Key         string
	Description string
}

// GlobalBindings returns the key bindings available on every screen.
func GlobalBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "Ctrl+C", Description: "Quit"},
		{Key: "Ctrl+T", Description: "Toggle theme"},
		{Key: "?", Description: "Show/hide help"},
		{Key: "Esc", Description: "Go back"},
	}
}
