// Package tui provides the terminal user interface for tap.
package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the lipgloss styles used in the TUI.
var Styles = struct {
	Header   lipgloss.Style
	Footer   lipgloss.Style
	Help     lipgloss.Style
	Selected lipgloss.Style
	Dimmed   lipgloss.Style
	Title    lipgloss.Style
	Error    lipgloss.Style
}{
	Header: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1),

	Footer: lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1),

	Help: lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Padding(1, 2),

	Selected: lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true),

	Dimmed: lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")),

	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")),

	Error: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true),
}
