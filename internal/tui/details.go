package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiungar/tap/internal/core"
)

// DetailsModel handles the script details panel.
type DetailsModel struct {
	script  *core.Script
	focused bool
	width   int
	height  int
}

// NewDetailsModel creates a new DetailsModel.
func NewDetailsModel() DetailsModel {
	return DetailsModel{
		script:  nil,
		focused: false,
	}
}

// Init implements tea.Model.
func (m DetailsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m DetailsModel) Update(msg tea.Msg) (DetailsModel, tea.Cmd) {
	// Details panel is read-only, no key handling needed
	return m, nil
}

// View renders the details panel.
func (m DetailsModel) View() string {
	// Determine panel style based on focus
	panelStyle := Styles.Panel
	if m.focused {
		panelStyle = Styles.PanelActive
	}

	// Calculate inner dimensions (content area)
	innerWidth, innerHeight := InnerDimensions(m.width, m.height)
	_ = innerWidth // Used in renderScriptDetails

	// Build lines array
	var lines []string

	// Title (takes 2 lines: title + blank)
	title := Styles.Title.Render(fmt.Sprintf("%s Details", Icons.Script))
	lines = append(lines, title)
	lines = append(lines, "")

	if m.script == nil {
		lines = append(lines, Styles.ItemDesc.Render("  Select a script to view details"))
	} else {
		// renderScriptDetails returns a multi-line string
		details := m.renderScriptDetails()
		detailLines := strings.Split(details, "\n")
		lines = append(lines, detailLines...)
	}

	// Pad content to exact inner height
	content := BuildPanelContent(lines, innerHeight)

	// Apply panel style - only set Width, NOT Height
	return panelStyle.
		Width(m.width - BorderWidth).
		Render(content)
}

// renderScriptDetails renders the full script information.
func (m DetailsModel) renderScriptDetails() string {
	if m.script == nil {
		return ""
	}

	var s strings.Builder
	script := m.script
	innerWidth := m.width - 6 // Account for borders and padding
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Script name with icon
	icon := IconForShell(script.Shell)
	nameStyle := lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true)
	s.WriteString(nameStyle.Render(fmt.Sprintf("%s %s", icon, script.Name)))
	s.WriteString("\n\n")

	// Description (word-wrapped)
	if script.Description != "" {
		desc := m.wrapText(script.Description, innerWidth)
		s.WriteString(Styles.Item.Render(desc))
		s.WriteString("\n\n")
	}

	// Separator
	separator := strings.Repeat("─", innerWidth)
	s.WriteString(Styles.ItemDesc.Render(separator))
	s.WriteString("\n\n")

	// Metadata section
	labelStyle := lipgloss.NewStyle().
		Foreground(Theme.Subtle).
		Width(12)
	valueStyle := lipgloss.NewStyle().
		Foreground(Theme.Foreground)

	// Category
	s.WriteString(labelStyle.Render("Category"))
	s.WriteString(valueStyle.Render(script.Category))
	s.WriteString("\n")

	// Shell
	s.WriteString(labelStyle.Render("Shell"))
	s.WriteString(valueStyle.Render(script.Shell))
	s.WriteString("\n")

	// Mode (interactive)
	if script.Interactive {
		modeStyle := lipgloss.NewStyle().Foreground(Theme.Secondary)
		s.WriteString(labelStyle.Render("Mode"))
		s.WriteString(modeStyle.Render("Interactive"))
		s.WriteString("\n")
	}

	// Path (shortened)
	path := script.Path
	homeDir := "~"
	if strings.HasPrefix(path, homeDir) || len(path) > innerWidth-12 {
		// Show just the directory
		path = filepath.Dir(path)
		if len(path) > innerWidth-15 {
			path = "..." + path[len(path)-innerWidth+18:]
		}
	}
	s.WriteString(labelStyle.Render("Path"))
	s.WriteString(valueStyle.Render(path))
	s.WriteString("\n")

	// Parameters section (if any)
	if len(script.Parameters) > 0 {
		s.WriteString("\n")
		s.WriteString(Styles.ItemDesc.Render(separator))
		s.WriteString("\n\n")

		paramHeader := lipgloss.NewStyle().
			Foreground(Theme.Secondary).
			Bold(true)
		s.WriteString(paramHeader.Render("Parameters"))
		s.WriteString("\n\n")

		for _, param := range script.Parameters {
			s.WriteString(m.renderParameter(param, innerWidth))
			s.WriteString("\n")
		}
	}

	// Tags (if any)
	if len(script.Tags) > 0 {
		s.WriteString("\n")
		tagStyle := lipgloss.NewStyle().
			Foreground(Theme.Subtle)
		s.WriteString(tagStyle.Render("Tags: " + strings.Join(script.Tags, ", ")))
	}

	return s.String()
}

// renderParameter renders a single parameter entry.
func (m DetailsModel) renderParameter(param core.Parameter, maxWidth int) string {
	var s strings.Builder

	// Parameter name with required indicator
	nameStyle := lipgloss.NewStyle().Foreground(Theme.Primary)
	name := param.Name
	if param.Required {
		name += " *"
	}
	s.WriteString("  • ")
	s.WriteString(nameStyle.Render(name))

	// Type and default
	typeInfo := fmt.Sprintf(" (%s", param.Type)
	if param.Default != "" && !param.Required {
		typeInfo += fmt.Sprintf(", default: %s", param.Default)
	}
	typeInfo += ")"
	s.WriteString(Styles.ItemDesc.Render(typeInfo))
	s.WriteString("\n")

	// Description (if any)
	if param.Description != "" {
		desc := m.wrapText(param.Description, maxWidth-6)
		lines := strings.Split(desc, "\n")
		for _, line := range lines {
			s.WriteString("    ")
			s.WriteString(Styles.ItemDesc.Render(line))
			s.WriteString("\n")
		}
	}

	return s.String()
}

// wrapText wraps text to the specified width.
func (m DetailsModel) wrapText(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// SetScript sets the script to display.
func (m *DetailsModel) SetScript(script *core.Script) {
	m.script = script
}

// Script returns the current script.
func (m DetailsModel) Script() *core.Script {
	return m.script
}

// SetFocused sets whether the panel is focused.
func (m *DetailsModel) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the panel is focused.
func (m DetailsModel) IsFocused() bool {
	return m.focused
}

// SetSize updates the panel dimensions.
func (m *DetailsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// PanelTitle returns the title for this panel.
func (m DetailsModel) PanelTitle() string {
	return "Details"
}
