package display

import "github.com/charmbracelet/lipgloss"

// Color palette. Kept small and semantic so presence states read consistently
// across `list`, `status`, and prompts.
var (
	colGreen  = lipgloss.Color("42")  // online
	colYellow = lipgloss.Color("214") // coding / warning
	colGray   = lipgloss.Color("244") // idle / offline / muted
	colRed    = lipgloss.Color("196") // error / MITM
	colBlue   = lipgloss.Color("39")  // headings, accents
)

// Theme bundles the styles the CLI uses. Build one with NewTheme so --no-color
// can disable ANSI styling globally.
type Theme struct {
	Online  lipgloss.Style
	Coding  lipgloss.Style
	Idle    lipgloss.Style
	Offline lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Heading lipgloss.Style
	Muted   lipgloss.Style
	Bold    lipgloss.Style
}

// NewTheme returns a theme. When color is false all styles are plain, so output
// is clean in pipes, logs, and color-averse terminals.
func NewTheme(color bool) Theme {
	if !color {
		plain := lipgloss.NewStyle()
		return Theme{
			Online: plain, Coding: plain, Idle: plain, Offline: plain,
			Warning: plain, Error: plain, Heading: plain, Muted: plain, Bold: plain,
		}
	}
	return Theme{
		Online:  lipgloss.NewStyle().Foreground(colGreen),
		Coding:  lipgloss.NewStyle().Foreground(colYellow),
		Idle:    lipgloss.NewStyle().Foreground(colGray),
		Offline: lipgloss.NewStyle().Foreground(colGray),
		Warning: lipgloss.NewStyle().Foreground(colYellow),
		Error:   lipgloss.NewStyle().Foreground(colRed).Bold(true),
		Heading: lipgloss.NewStyle().Foreground(colBlue).Bold(true),
		Muted:   lipgloss.NewStyle().Foreground(colGray),
		Bold:    lipgloss.NewStyle().Bold(true),
	}
}

// Active is the process-wide theme, set once during startup.
var Active = NewTheme(true)
