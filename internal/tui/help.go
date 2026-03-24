package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpSection represents a categorized group of keybindings.
type HelpSection struct {
	Title    string
	Bindings []HelpBinding
}

// HelpBinding represents a single keybinding with its description.
type HelpBinding struct {
	Key  string
	Desc string
}

// HelpModel manages the help overlay display.
type HelpModel struct {
	width  int
	height int
}

// NewHelpModel creates a new help model.
func NewHelpModel() HelpModel {
	return HelpModel{
		width:  80,
		height: 24,
	}
}

// SetSize sets the help overlay dimensions.
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Sections returns the categorized help sections.
func (m HelpModel) Sections() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Bindings: []HelpBinding{
				{Key: "↑ / k", Desc: "Move up"},
				{Key: "↓ / j", Desc: "Move down"},
				{Key: "enter", Desc: "Select / Run script"},
				{Key: "esc", Desc: "Go back / Cancel"},
			},
		},
		{
			Title: "Panels",
			Bindings: []HelpBinding{
				{Key: "tab", Desc: "Next panel"},
				{Key: "shift+tab", Desc: "Previous panel"},
			},
		},
		{
			Title: "Filtering",
			Bindings: []HelpBinding{
				{Key: "/", Desc: "Start filtering"},
				{Key: "esc", Desc: "Cancel filter"},
				{Key: "enter", Desc: "Confirm filter"},
			},
		},
		{
			Title: "Script",
			Bindings: []HelpBinding{
				{Key: "v", Desc: "View script source code"},
				{Key: "e", Desc: "Edit script in $EDITOR"},
			},
		},
		{
			Title: "Actions",
			Bindings: []HelpBinding{
				{Key: "r", Desc: "Refresh scripts"},
				{Key: "?", Desc: "Toggle this help"},
				{Key: "q", Desc: "Quit"},
			},
		},
	}
}

// View renders the help overlay.
func (m HelpModel) View() string {
	// Define styles for help overlay
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Primary).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Secondary).
		MarginTop(1).
		MarginBottom(0)

	keyStyle := lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true).
		Width(14)

	descStyle := lipgloss.NewStyle().
		Foreground(Theme.Foreground)

	footerStyle := lipgloss.NewStyle().
		Foreground(Theme.Subtle).
		MarginTop(1).
		Italic(true)

	// Build the content
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n")

	// Sections
	for _, section := range m.Sections() {
		b.WriteString("\n")
		b.WriteString(sectionTitleStyle.Render(section.Title))
		b.WriteString("\n")

		for _, binding := range section.Bindings {
			b.WriteString("  ")
			b.WriteString(keyStyle.Render(binding.Key))
			b.WriteString(descStyle.Render(binding.Desc))
			b.WriteString("\n")
		}
	}

	// Footer hint
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Press any key to close"))

	content := b.String()

	// Create the overlay box with rounded border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BorderActive).
		Padding(1, 3).
		Width(50)

	box := boxStyle.Render(content)

	// Center the box
	return m.centerBox(box)
}

// centerBox centers the box in the available space.
func (m HelpModel) centerBox(box string) string {
	boxLines := strings.Split(box, "\n")
	boxHeight := len(boxLines)
	boxWidth := lipgloss.Width(box)

	// Calculate vertical padding
	vertPadding := max((m.height-boxHeight)/2, 0)

	// Calculate horizontal padding
	horizPadding := max((m.width-boxWidth)/2, 0)

	var result strings.Builder

	// Add top padding
	for range vertPadding {
		result.WriteString("\n")
	}

	// Add centered box lines
	indent := strings.Repeat(" ", horizPadding)
	for _, line := range boxLines {
		result.WriteString(indent)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// RenderHelp renders the help overlay with the given dimensions.
// This is a convenience function for direct rendering without creating a model.
func RenderHelp(width, height int) string {
	m := HelpModel{
		width:  width,
		height: height,
	}
	return m.View()
}
