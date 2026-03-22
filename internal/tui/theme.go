package tui

import "github.com/jparkerweb/shrinkray/internal/tui/style"

// Theme is re-exported from tui/style for backward compatibility.
type Theme = style.Theme

// ThemeName is re-exported from tui/style for backward compatibility.
type ThemeName = style.ThemeName

// Theme name constants re-exported.
const (
	ThemeNeonDusk       = style.ThemeNeonDusk
	ThemeElectricSunset = style.ThemeElectricSunset
)

// NeonDusk returns the default theme.
func NeonDusk() Theme { return style.NeonDusk() }

// ElectricSunset returns the Electric Sunset theme.
func ElectricSunset() Theme { return style.ElectricSunset() }

// ToggleTheme switches the active theme.
func ToggleTheme() ThemeName { return style.ToggleTheme() }

// SetTheme sets the active theme by name.
func SetTheme(name ThemeName) bool { return style.SetTheme(name) }

// ActiveTheme is re-exported for backward compatibility.
var ActiveTheme = style.ActiveTheme
