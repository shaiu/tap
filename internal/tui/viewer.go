package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiungar/tap/internal/core"
)

// ViewerModel displays script source code with scrolling.
type ViewerModel struct {
	script *core.Script
	lines  []string // source lines
	offset int      // first visible line index
	width  int
	height int
	err    error
}

// NewViewerModel creates a viewer for the given script.
func NewViewerModel(script core.Script, width, height int) ViewerModel {
	m := ViewerModel{
		script: &script,
		width:  width,
		height: height,
	}

	data, err := os.ReadFile(script.Path)
	if err != nil {
		m.err = err
		return m
	}

	m.lines = strings.Split(string(data), "\n")
	// Remove trailing empty line from split
	if len(m.lines) > 0 && m.lines[len(m.lines)-1] == "" {
		m.lines = m.lines[:len(m.lines)-1]
	}

	return m
}

// visibleLines returns how many code lines fit in the view area.
func (m ViewerModel) visibleLines() int {
	// 2 for top border+padding, 2 for bottom border+padding, 2 for title+blank, 1 for footer hint
	available := m.height - 7
	if available < 1 {
		return 1
	}
	return available
}

// Update handles key input for the viewer.
func (m ViewerModel) Update(msg tea.KeyMsg) (ViewerModel, tea.Cmd) {
	visible := m.visibleLines()
	maxOffset := len(m.lines) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}

	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.offset > 0 {
			m.offset--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.offset < maxOffset {
			m.offset++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
		m.offset -= visible / 2
		if m.offset < 0 {
			m.offset = 0
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
		m.offset += visible / 2
		if m.offset > maxOffset {
			m.offset = maxOffset
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
		m.offset = 0
	case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
		m.offset = maxOffset
	}

	return m, nil
}

// View renders the code viewer.
func (m ViewerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error reading file: %s", m.err)
	}

	// Calculate inner dimensions
	contentWidth := m.width - 6 // borders + padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	visible := m.visibleLines()

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Primary)

	path := m.script.Path
	title := titleStyle.Render(fmt.Sprintf("%s %s", IconForShell(m.script.Shell), m.script.Name))

	pathStyle := lipgloss.NewStyle().Foreground(Theme.Subtle)
	posStyle := lipgloss.NewStyle().Foreground(Theme.Subtle)

	// Line position indicator
	endLine := m.offset + visible
	if endLine > len(m.lines) {
		endLine = len(m.lines)
	}
	position := posStyle.Render(fmt.Sprintf("[%d-%d / %d]", m.offset+1, endLine, len(m.lines)))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("  ")
	b.WriteString(position)
	b.WriteString("\n")
	b.WriteString(pathStyle.Render(TruncateLineSimple(path, contentWidth)))
	b.WriteString("\n")

	// Separator
	separator := lipgloss.NewStyle().Foreground(Theme.Border).Render(strings.Repeat("─", contentWidth))
	b.WriteString(separator)
	b.WriteString("\n")

	// Line number gutter width
	gutterWidth := len(fmt.Sprintf("%d", len(m.lines)))
	if gutterWidth < 3 {
		gutterWidth = 3
	}

	lineNumStyle := lipgloss.NewStyle().Foreground(Theme.Subtle).Width(gutterWidth).Align(lipgloss.Right)
	codeStyle := lipgloss.NewStyle().Foreground(Theme.Foreground)

	// Render visible lines
	for i := 0; i < visible; i++ {
		lineIdx := m.offset + i
		if lineIdx >= len(m.lines) {
			b.WriteString("\n")
			continue
		}

		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", lineIdx+1))
		divider := lipgloss.NewStyle().Foreground(Theme.Border).Render(" │ ")

		// Truncate code line to fit
		maxCodeWidth := contentWidth - gutterWidth - 3 // gutter + divider
		codeLine := m.lines[lineIdx]
		// Replace tabs with spaces for display
		codeLine = strings.ReplaceAll(codeLine, "\t", "    ")
		if len(codeLine) > maxCodeWidth {
			codeLine = codeLine[:maxCodeWidth]
		}

		b.WriteString(lineNum)
		b.WriteString(divider)
		b.WriteString(codeStyle.Render(codeLine))
		if i < visible-1 {
			b.WriteString("\n")
		}
	}

	// Wrap in a panel
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BorderActive).
		Padding(0, 1).
		Width(m.width - BorderWidth)

	return boxStyle.Render(b.String())
}
