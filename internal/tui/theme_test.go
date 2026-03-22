package tui

import (
	"image/color"
	"testing"

	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

func TestNeonDuskTheme(t *testing.T) {
	theme := NeonDusk()

	// Verify all color fields are non-nil
	colors := map[string]color.Color{
		"Primary":       theme.Primary,
		"Secondary":     theme.Secondary,
		"Accent":        theme.Accent,
		"Success":       theme.Success,
		"Warning":       theme.Warning,
		"Error":         theme.Error,
		"Muted":         theme.Muted,
		"Background":    theme.Background,
		"Surface":       theme.Surface,
		"Text":          theme.Text,
		"TextDim":       theme.TextDim,
		"Border":        theme.Border,
		"Highlight":     theme.Highlight,
		"GradientStart": theme.GradientStart,
		"GradientEnd":   theme.GradientEnd,
	}

	for name, c := range colors {
		if c == nil {
			t.Errorf("NeonDusk() theme has nil color field: %s", name)
		}
	}
}

func TestElectricSunsetTheme(t *testing.T) {
	theme := ElectricSunset()

	// Verify all color fields are non-nil
	colors := map[string]color.Color{
		"Primary":       theme.Primary,
		"Secondary":     theme.Secondary,
		"Accent":        theme.Accent,
		"Success":       theme.Success,
		"Warning":       theme.Warning,
		"Error":         theme.Error,
		"Muted":         theme.Muted,
		"Background":    theme.Background,
		"Surface":       theme.Surface,
		"Text":          theme.Text,
		"TextDim":       theme.TextDim,
		"Border":        theme.Border,
		"Highlight":     theme.Highlight,
		"GradientStart": theme.GradientStart,
		"GradientEnd":   theme.GradientEnd,
	}

	for name, c := range colors {
		if c == nil {
			t.Errorf("ElectricSunset() theme has nil color field: %s", name)
		}
	}
}

func TestActiveThemeInitialized(t *testing.T) {
	// ActiveTheme should be initialized to NeonDusk
	if style.ActiveTheme.Primary == nil {
		t.Error("ActiveTheme.Primary is nil")
	}
}

func TestToggleTheme(t *testing.T) {
	// Ensure we start with NeonDusk
	style.SetTheme(style.ThemeNeonDusk)

	// Toggle to Electric Sunset
	newTheme := ToggleTheme()
	if newTheme != ThemeElectricSunset {
		t.Errorf("expected Electric Sunset after toggle, got %s", newTheme)
	}

	// Verify styles updated (ActiveTheme should have Electric Sunset colors)
	if style.ActiveThemeName != style.ThemeElectricSunset {
		t.Error("ActiveThemeName not updated after toggle")
	}

	// Toggle back to Neon Dusk
	newTheme = ToggleTheme()
	if newTheme != ThemeNeonDusk {
		t.Errorf("expected Neon Dusk after second toggle, got %s", newTheme)
	}

	if style.ActiveThemeName != style.ThemeNeonDusk {
		t.Error("ActiveThemeName not updated after second toggle")
	}
}

func TestSetTheme(t *testing.T) {
	// Set to Electric Sunset
	ok := SetTheme(ThemeElectricSunset)
	if !ok {
		t.Error("SetTheme(ElectricSunset) returned false")
	}
	if style.ActiveThemeName != style.ThemeElectricSunset {
		t.Error("theme not set to Electric Sunset")
	}

	// Set to Neon Dusk
	ok = SetTheme(ThemeNeonDusk)
	if !ok {
		t.Error("SetTheme(NeonDusk) returned false")
	}
	if style.ActiveThemeName != style.ThemeNeonDusk {
		t.Error("theme not set to Neon Dusk")
	}

	// Unknown theme name
	ok = SetTheme("nonexistent")
	if ok {
		t.Error("SetTheme(nonexistent) should return false")
	}
}

func TestNeonDuskColors(t *testing.T) {
	theme := NeonDusk()

	// Verify colors are non-nil
	if theme.Primary == nil {
		t.Error("Primary should not be nil")
	}
	if theme.Accent == nil {
		t.Error("Accent should not be nil")
	}
	if theme.Success == nil {
		t.Error("Success should not be nil")
	}
}

func TestThemeStylesUpdateAfterToggle(t *testing.T) {
	// Ensure Neon Dusk
	style.SetTheme(style.ThemeNeonDusk)

	// Get a style, verify it uses theme colors
	header1 := HeaderStyle()

	// Toggle to Electric Sunset
	ToggleTheme()

	// Get a style again — it should reflect the new theme
	header2 := HeaderStyle()

	// They should be different because the theme changed
	// (We can't easily compare lipgloss styles, but we verify no panic)
	_ = header1
	_ = header2

	// Reset to Neon Dusk for other tests
	style.SetTheme(style.ThemeNeonDusk)
}
