package tui

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	bindings := map[string]key.Binding{
		"Quit":        km.Quit,
		"Back":        km.Back,
		"Enter":       km.Enter,
		"Help":        km.Help,
		"Tab":         km.Tab,
		"Up":          km.Up,
		"Down":        km.Down,
		"Left":        km.Left,
		"Right":       km.Right,
		"ThemeToggle": km.ThemeToggle,
	}

	for name, binding := range bindings {
		keys := binding.Keys()
		if len(keys) == 0 {
			t.Errorf("DefaultKeyMap().%s has no key bindings", name)
		}

		help := binding.Help()
		if help.Key == "" {
			t.Errorf("DefaultKeyMap().%s has no help key text", name)
		}
		if help.Desc == "" {
			t.Errorf("DefaultKeyMap().%s has no help description", name)
		}
	}
}

func TestDefaultKeyMapSpecificKeys(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name     string
		binding  key.Binding
		expected []string
	}{
		{"Quit", km.Quit, []string{"q", "ctrl+c"}},
		{"Back", km.Back, []string{"esc"}},
		{"Enter", km.Enter, []string{"enter"}},
		{"Help", km.Help, []string{"?"}},
		{"Tab", km.Tab, []string{"tab"}},
		{"Up", km.Up, []string{"up"}},
		{"Down", km.Down, []string{"down"}},
		{"Left", km.Left, []string{"left"}},
		{"Right", km.Right, []string{"right"}},
		{"ThemeToggle", km.ThemeToggle, []string{"ctrl+t"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keys := tc.binding.Keys()
			if len(keys) != len(tc.expected) {
				t.Errorf("%s: got %d keys, want %d", tc.name, len(keys), len(tc.expected))
				return
			}
			for i, k := range keys {
				if k != tc.expected[i] {
					t.Errorf("%s: key[%d] = %q, want %q", tc.name, i, k, tc.expected[i])
				}
			}
		})
	}
}
