package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilterModel manages the filter overlay display.
type FilterModel struct {
	input      textinput.Model
	width      int
	height     int
	matchCount int
	totalCount int
}

// NewFilterModel creates a new filter model.
func NewFilterModel() FilterModel {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30
	ti.Prompt = ""

	return FilterModel{
		input:      ti,
		width:      80,
		height:     24,
		matchCount: 0,
		totalCount: 0,
	}
}

// Init implements tea.Model.
func (m FilterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// SetSize sets the overlay dimensions.
func (m *FilterModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Adjust input width based on overlay size
	overlayWidth := m.overlayWidth()
	m.input.Width = overlayWidth - 10 // Account for prompt, cursor, padding
	if m.input.Width < 20 {
		m.input.Width = 20
	}
}

// overlayWidth calculates the overlay box width.
func (m FilterModel) overlayWidth() int {
	// Overlay takes about 60% of screen width, min 40, max 60
	w := m.width * 60 / 100
	if w < 40 {
		w = 40
	}
	if w > 60 {
		w = 60
	}
	return w
}

// SetCounts updates the match/total counts.
func (m *FilterModel) SetCounts(matched, total int) {
	m.matchCount = matched
	m.totalCount = total
}

// Focus focuses the text input.
func (m *FilterModel) Focus() tea.Cmd {
	return m.input.Focus()
}

// Blur removes focus from the text input.
func (m *FilterModel) Blur() {
	m.input.Blur()
}

// Value returns the current filter query.
func (m FilterModel) Value() string {
	return m.input.Value()
}

// SetValue sets the filter query.
func (m *FilterModel) SetValue(s string) {
	m.input.SetValue(s)
}

// Reset clears the filter input.
func (m *FilterModel) Reset() {
	m.input.SetValue("")
	m.matchCount = 0
	m.totalCount = 0
}

// View renders the filter overlay.
func (m FilterModel) View() string {
	return m.renderOverlay()
}

// renderOverlay renders the styled overlay box.
func (m FilterModel) renderOverlay() string {
	overlayWidth := m.overlayWidth()

	// Build content
	var b strings.Builder

	// Title line with icon and match count
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Primary)

	countStyle := lipgloss.NewStyle().
		Foreground(Theme.Subtle)

	title := titleStyle.Render(fmt.Sprintf("%s Filter", Icons.Search))
	count := ""
	if m.totalCount > 0 {
		count = countStyle.Render(fmt.Sprintf(" [%d/%d]", m.matchCount, m.totalCount))
	}

	// Calculate padding to right-align the count
	titleLen := lipgloss.Width(title)
	countLen := lipgloss.Width(count)
	availableWidth := overlayWidth - 6 // Account for box padding
	spacingLen := availableWidth - titleLen - countLen
	if spacingLen < 1 {
		spacingLen = 1
	}
	spacing := strings.Repeat(" ", spacingLen)

	b.WriteString(title)
	b.WriteString(spacing)
	b.WriteString(count)
	b.WriteString("\n\n")

	// Input field
	inputStyle := lipgloss.NewStyle().
		Foreground(Theme.Foreground)

	b.WriteString(inputStyle.Render(m.input.View()))

	content := b.String()

	// Create the overlay box with rounded border (superfile-style)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BorderActive).
		Padding(1, 2).
		Width(overlayWidth)

	return boxStyle.Render(content)
}

// RenderOverlayAt renders the filter overlay positioned for use in a composite view.
// It returns the overlay box that can be placed on top of other content.
func (m FilterModel) RenderOverlayAt(contentWidth, contentHeight int) string {
	overlay := m.renderOverlay()

	// Calculate centering within the content area
	overlayLines := strings.Split(overlay, "\n")
	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := len(overlayLines)

	// Horizontal centering
	horizPadding := (contentWidth - overlayWidth) / 2
	if horizPadding < 0 {
		horizPadding = 0
	}

	// Vertical position - place near top, not dead center
	vertPadding := contentHeight / 6
	if vertPadding < 2 {
		vertPadding = 2
	}

	var result strings.Builder

	// Add top padding
	for i := 0; i < vertPadding; i++ {
		result.WriteString("\n")
	}

	// Add centered overlay lines
	indent := strings.Repeat(" ", horizPadding)
	for i, line := range overlayLines {
		result.WriteString(indent)
		result.WriteString(line)
		if i < overlayHeight-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// MatchCount returns the current match count.
func (m FilterModel) MatchCount() int {
	return m.matchCount
}

// TotalCount returns the total count.
func (m FilterModel) TotalCount() int {
	return m.totalCount
}
