// Package tui provides the terminal user interface for tap.
package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the lipgloss styles used in the TUI.
// All styles use the Theme colors for consistent Catppuccin Mocha styling.
var Styles = struct {
	// Existing styles (updated to use Theme)
	Header      lipgloss.Style
	Footer      lipgloss.Style
	Help        lipgloss.Style
	Selected    lipgloss.Style
	Dimmed      lipgloss.Style
	Title       lipgloss.Style
	Error       lipgloss.Style
	FilterInput lipgloss.Style
	Required    lipgloss.Style

	// Panel styles (new)
	Panel       lipgloss.Style
	PanelActive lipgloss.Style

	// List item styles (new)
	Item         lipgloss.Style
	ItemSelected lipgloss.Style
	ItemDesc     lipgloss.Style

	// Footer hint styles (new)
	Key    lipgloss.Style
	Action lipgloss.Style

	// Filter styles (for inline filter bar)
	FilterQuery lipgloss.Style
	FilterCount lipgloss.Style
}{
	Header: lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Primary).
		MarginBottom(1),

	Footer: lipgloss.NewStyle().
		Foreground(Theme.Subtle).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(Theme.Border).
		Padding(0, 1),

	Help: lipgloss.NewStyle().
		Foreground(Theme.Subtle).
		Padding(1, 2),

	Selected: lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true),

	Dimmed: lipgloss.NewStyle().
		Foreground(Theme.Subtle),

	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Primary).
		Padding(0, 1),

	Error: lipgloss.NewStyle().
		Foreground(Theme.Error).
		Bold(true),

	FilterInput: lipgloss.NewStyle().
		Foreground(Theme.Primary).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Border).
		Padding(0, 1),

	Required: lipgloss.NewStyle().
		Foreground(Theme.Error),

	// Panel with rounded border (inactive)
	Panel: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Border).
		Padding(0, 1),

	// Panel with active (focused) border
	PanelActive: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BorderActive).
		Padding(0, 1),

	// Normal list item
	Item: lipgloss.NewStyle().
		Foreground(Theme.Foreground),

	// Selected list item (cursor on it)
	ItemSelected: lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Background(Theme.Selection).
		Bold(true),

	// Item description (secondary text)
	ItemDesc: lipgloss.NewStyle().
		Foreground(Theme.Subtle),

	// Key hint in footer (highlighted)
	Key: lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true),

	// Action text in footer
	Action: lipgloss.NewStyle().
		Foreground(Theme.Foreground),

	// Filter query text (highlighted)
	FilterQuery: lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true),

	// Filter match count (dimmed, like [3/12])
	FilterCount: lipgloss.NewStyle().
		Foreground(Theme.Subtle),
}
