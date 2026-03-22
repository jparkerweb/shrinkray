package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds all color values for the application's visual theme.
type Theme struct {
	Primary       color.Color
	Secondary     color.Color
	Accent        color.Color
	Success       color.Color
	Warning       color.Color
	Error         color.Color
	Muted         color.Color
	Background    color.Color
	Surface       color.Color
	Text          color.Color
	TextDim       color.Color
	Border        color.Color
	Highlight     color.Color
	GradientStart color.Color
	GradientEnd   color.Color
}

// ThemeName identifies a theme by name string.
type ThemeName string

const (
	ThemeNeonDusk       ThemeName = "neon-dusk"
	ThemeElectricSunset ThemeName = "electric-sunset"
)

// NeonDusk returns the default "Neon Dusk" theme with deep purples and
// neon cyan/magenta accents.
func NeonDusk() Theme {
	return Theme{
		Primary:       lipgloss.Color("#7B2FF7"), // electric violet
		Secondary:     lipgloss.Color("#FF2D95"), // hot pink
		Accent:        lipgloss.Color("#00F0FF"), // cyan
		Success:       lipgloss.Color("#39FF14"), // neon green
		Warning:       lipgloss.Color("#FFAB00"), // amber
		Error:         lipgloss.Color("#FF4444"), // red
		Muted:         lipgloss.Color("#6C6C8A"), // soft gray
		Background:    lipgloss.Color("#1A1A2E"), // deep navy
		Surface:       lipgloss.Color("#25253E"), // slightly lighter navy
		Text:          lipgloss.Color("#E8E8F0"), // near-white
		TextDim:       lipgloss.Color("#9999B3"), // dimmed text
		Border:        lipgloss.Color("#6C6C8A"), // soft gray
		Highlight:     lipgloss.Color("#7B2FF7"), // electric violet
		GradientStart: lipgloss.Color("#7B2FF7"), // electric violet
		GradientEnd:   lipgloss.Color("#FF2D95"), // hot pink
	}
}

// ElectricSunset returns the "Electric Sunset" theme with warm oranges,
// coral reds, and golden yellows.
func ElectricSunset() Theme {
	return Theme{
		Primary:       lipgloss.Color("#FF6B6B"), // coral red
		Secondary:     lipgloss.Color("#FFD93D"), // golden yellow
		Accent:        lipgloss.Color("#FF8E53"), // warm orange
		Success:       lipgloss.Color("#39FF14"), // neon green
		Warning:       lipgloss.Color("#FFAB00"), // amber
		Error:         lipgloss.Color("#FF4444"), // red
		Muted:         lipgloss.Color("#6C6C6C"), // gray
		Background:    lipgloss.Color("#1A1A1A"), // dark charcoal
		Surface:       lipgloss.Color("#2A2A2A"), // slightly lighter charcoal
		Text:          lipgloss.Color("#F0E8E0"), // warm off-white
		TextDim:       lipgloss.Color("#A09888"), // warm dimmed text
		Border:        lipgloss.Color("#6C6C6C"), // gray
		Highlight:     lipgloss.Color("#FF6B6B"), // coral red
		GradientStart: lipgloss.Color("#FF6B6B"), // coral red
		GradientEnd:   lipgloss.Color("#FFD93D"), // golden yellow
	}
}

// ActiveTheme is the currently active theme. It is initialized to Neon Dusk.
var ActiveTheme = NeonDusk()

// ActiveThemeName tracks which theme is currently active.
var ActiveThemeName ThemeName = ThemeNeonDusk

// ToggleTheme switches the active theme between Neon Dusk and Electric Sunset.
// It returns the new theme name.
func ToggleTheme() ThemeName {
	if ActiveThemeName == ThemeNeonDusk {
		ActiveTheme = ElectricSunset()
		ActiveThemeName = ThemeElectricSunset
	} else {
		ActiveTheme = NeonDusk()
		ActiveThemeName = ThemeNeonDusk
	}
	return ActiveThemeName
}

// SetTheme sets the active theme by name. Returns false if the name is unknown.
func SetTheme(name ThemeName) bool {
	switch name {
	case ThemeNeonDusk:
		ActiveTheme = NeonDusk()
		ActiveThemeName = ThemeNeonDusk
		return true
	case ThemeElectricSunset:
		ActiveTheme = ElectricSunset()
		ActiveThemeName = ThemeElectricSunset
		return true
	default:
		return false
	}
}
