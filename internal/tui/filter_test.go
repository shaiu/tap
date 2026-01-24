package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFilterModel(t *testing.T) {
	m := NewFilterModel()

	assert.Equal(t, "", m.Value(), "Initial value should be empty")
	assert.Equal(t, 0, m.MatchCount(), "Initial match count should be 0")
	assert.Equal(t, 0, m.TotalCount(), "Initial total count should be 0")
	assert.Equal(t, 80, m.width, "Default width should be 80")
	assert.Equal(t, 24, m.height, "Default height should be 24")
}

func TestFilterModel_SetSize(t *testing.T) {
	m := NewFilterModel()
	m.SetSize(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Greater(t, m.input.Width, 20, "Input width should be adjusted")
}

func TestFilterModel_SetCounts(t *testing.T) {
	m := NewFilterModel()
	m.SetCounts(5, 12)

	assert.Equal(t, 5, m.MatchCount())
	assert.Equal(t, 12, m.TotalCount())
}

func TestFilterModel_Value(t *testing.T) {
	m := NewFilterModel()
	m.SetValue("deploy")

	assert.Equal(t, "deploy", m.Value())
}

func TestFilterModel_Reset(t *testing.T) {
	m := NewFilterModel()
	m.SetValue("deploy")
	m.SetCounts(5, 12)

	m.Reset()

	assert.Equal(t, "", m.Value())
	assert.Equal(t, 0, m.MatchCount())
	assert.Equal(t, 0, m.TotalCount())
}

func TestFilterModel_View(t *testing.T) {
	m := NewFilterModel()
	m.SetSize(100, 30)
	m.SetCounts(3, 12)

	view := m.View()

	// Should contain the filter icon
	assert.Contains(t, view, Icons.Search, "View should contain search icon")

	// Should contain "Filter"
	assert.Contains(t, view, "Filter", "View should contain 'Filter' text")

	// Should have a box border (rounded)
	assert.Contains(t, view, "╭", "View should have rounded top-left corner")
	assert.Contains(t, view, "╮", "View should have rounded top-right corner")
	assert.Contains(t, view, "╰", "View should have rounded bottom-left corner")
	assert.Contains(t, view, "╯", "View should have rounded bottom-right corner")
}

func TestFilterModel_ViewWithCounts(t *testing.T) {
	m := NewFilterModel()
	m.SetSize(100, 30)
	m.SetCounts(3, 12)

	view := m.View()

	// Should show match count in [x/y] format
	assert.Contains(t, view, "[3/12]", "View should display match count")
}

func TestFilterModel_ViewWithoutCounts(t *testing.T) {
	m := NewFilterModel()
	m.SetSize(100, 30)
	// Don't set counts (totalCount is 0)

	view := m.View()

	// Should NOT show match count when totalCount is 0
	assert.NotContains(t, view, "[0/0]", "View should not display zero counts")
}

func TestFilterModel_Update(t *testing.T) {
	m := NewFilterModel()

	// Simulate typing a character
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updated, _ := m.Update(keyMsg)

	// Input should process the key
	// Note: The exact behavior depends on the textinput model
	require.NotNil(t, updated)
}

func TestFilterModel_OverlayWidth(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		minWidth  int
		maxWidth  int
	}{
		{
			name:     "narrow terminal",
			width:    50,
			minWidth: 40,
			maxWidth: 40, // Clamped to min
		},
		{
			name:     "normal terminal",
			width:    100,
			minWidth: 40,
			maxWidth: 60,
		},
		{
			name:     "wide terminal",
			width:    200,
			minWidth: 60,
			maxWidth: 60, // Clamped to max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewFilterModel()
			m.SetSize(tt.width, 30)

			overlayW := m.overlayWidth()

			assert.GreaterOrEqual(t, overlayW, tt.minWidth, "Overlay should be at least min width")
			assert.LessOrEqual(t, overlayW, tt.maxWidth, "Overlay should be at most max width")
		})
	}
}

func TestFilterModel_RenderOverlayAt(t *testing.T) {
	m := NewFilterModel()
	m.SetSize(100, 30)
	m.SetCounts(3, 12)

	overlay := m.RenderOverlayAt(100, 30)

	// Overlay should be positioned (has leading newlines for vertical padding)
	lines := strings.Split(overlay, "\n")
	assert.Greater(t, len(lines), 3, "Overlay should have multiple lines including padding")

	// First few lines should be empty (vertical padding)
	assert.Equal(t, "", strings.TrimSpace(lines[0]), "First line should be padding")
}

func TestFilterModel_FocusBlur(t *testing.T) {
	m := NewFilterModel()

	// Focus should return a command
	cmd := m.Focus()
	assert.NotNil(t, cmd, "Focus should return a blink command")

	// Blur should work without panic
	m.Blur()
	// No assertion needed, just checking it doesn't panic
}
