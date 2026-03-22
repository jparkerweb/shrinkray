package tui

import (
	"charm.land/lipgloss/v2"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// Re-export all style functions from tui/style for backward compatibility.

func HeaderStyle() lipgloss.Style  { return style.HeaderStyle() }
func FooterStyle() lipgloss.Style  { return style.FooterStyle() }
func TitleStyle() lipgloss.Style   { return style.TitleStyle() }
func SubtitleStyle() lipgloss.Style { return style.SubtitleStyle() }
func CardStyle() lipgloss.Style    { return style.CardStyle() }
func CardActiveStyle() lipgloss.Style { return style.CardActiveStyle() }
func ButtonStyle() lipgloss.Style  { return style.ButtonStyle() }
func ButtonActiveStyle() lipgloss.Style { return style.ButtonActiveStyle() }
func StatBoxStyle() lipgloss.Style { return style.StatBoxStyle() }
func ProgressBarStyle() lipgloss.Style { return style.ProgressBarStyle() }
func ErrorStyle() lipgloss.Style   { return style.ErrorStyle() }
func MutedStyle() lipgloss.Style   { return style.MutedStyle() }
func KeyHintStyle() lipgloss.Style { return style.KeyHintStyle() }
func AccentStyle() lipgloss.Style  { return style.AccentStyle() }
func SuccessStyle() lipgloss.Style { return style.SuccessStyle() }
func WarningStyle() lipgloss.Style { return style.WarningStyle() }
func LabelStyle() lipgloss.Style   { return style.LabelStyle() }
func ValueStyle() lipgloss.Style   { return style.ValueStyle() }
func BadgeStyle() lipgloss.Style   { return style.BadgeStyle() }
func TerminalWidth() int           { return style.TerminalWidth() }
func TerminalHeight() int          { return style.TerminalHeight() }
