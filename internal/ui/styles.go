package ui

import "github.com/charmbracelet/lipgloss"

// Styles using ANSI color numbers (0-7) to respect terminal themes.
var (
	Green  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(2))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(1))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3))
	Cyan   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(6))
	Bold   = lipgloss.NewStyle().Bold(true)
	Faint  = lipgloss.NewStyle().Faint(true)

	Success = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(2)).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(6))
	Warn    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3))
)
