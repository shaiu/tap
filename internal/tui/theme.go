// Package tui provides the terminal user interface for tap.
package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds color definitions for the UI.
// Default palette is Catppuccin Mocha, inspired by superfile.
var Theme = struct {
	// Base colors
	Background lipgloss.Color // Dark background
	Foreground lipgloss.Color // Light text
	Subtle     lipgloss.Color // Muted text

	// Accent colors
	Primary   lipgloss.Color // Blue (active/selected)
	Secondary lipgloss.Color // Mauve (secondary accent)
	Success   lipgloss.Color // Green (success states)
	Warning   lipgloss.Color // Yellow (warnings)
	Error     lipgloss.Color // Red (errors)

	// UI Elements
	Border       lipgloss.Color // Inactive borders
	BorderActive lipgloss.Color // Active panel border
	Selection    lipgloss.Color // Selection background

	// Gradient (for title bars - future use)
	GradientStart lipgloss.Color
	GradientEnd   lipgloss.Color
}{
	// Base - Catppuccin Mocha
	Background: lipgloss.Color("#1e1e2e"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Subtle:     lipgloss.Color("#6c7086"),

	// Accents
	Primary:   lipgloss.Color("#89b4fa"), // Blue
	Secondary: lipgloss.Color("#cba6f7"), // Mauve
	Success:   lipgloss.Color("#a6e3a1"), // Green
	Warning:   lipgloss.Color("#f9e2af"), // Yellow
	Error:     lipgloss.Color("#f38ba8"), // Red

	// UI Elements
	Border:       lipgloss.Color("#6c7086"), // Same as Subtle
	BorderActive: lipgloss.Color("#b4befe"), // Lavender
	Selection:    lipgloss.Color("#313244"), // Surface0

	// Gradient
	GradientStart: lipgloss.Color("#89b4fa"),
	GradientEnd:   lipgloss.Color("#cba6f7"),
}
