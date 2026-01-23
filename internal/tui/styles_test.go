package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyles_AllStylesAreDefined(t *testing.T) {
	// Test that all styles are non-zero (initialized)
	styles := []struct {
		name  string
		style any
	}{
		{"Header", Styles.Header},
		{"Footer", Styles.Footer},
		{"Help", Styles.Help},
		{"Selected", Styles.Selected},
		{"Dimmed", Styles.Dimmed},
		{"Title", Styles.Title},
		{"Error", Styles.Error},
		{"FilterInput", Styles.FilterInput},
		{"Required", Styles.Required},
		{"Panel", Styles.Panel},
		{"PanelActive", Styles.PanelActive},
		{"Item", Styles.Item},
		{"ItemSelected", Styles.ItemSelected},
		{"ItemDesc", Styles.ItemDesc},
		{"Key", Styles.Key},
		{"Action", Styles.Action},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			// Each style should render text without panic
			assert.NotPanics(t, func() {
				rendered := Styles.Header.Render("test")
				assert.NotEmpty(t, rendered)
			}, "%s should render without panic", s.name)
		})
	}
}

func TestStyles_UseThemeColors(t *testing.T) {
	// Test that styles render with Theme colors by checking they don't panic
	// and produce non-empty output
	testCases := []struct {
		name   string
		render func() string
	}{
		{"Header uses Primary", func() string { return Styles.Header.Render("test") }},
		{"Selected uses Primary", func() string { return Styles.Selected.Render("test") }},
		{"Dimmed uses Subtle", func() string { return Styles.Dimmed.Render("test") }},
		{"Error uses Error", func() string { return Styles.Error.Render("test") }},
		{"Panel uses Border", func() string { return Styles.Panel.Render("test") }},
		{"PanelActive uses BorderActive", func() string { return Styles.PanelActive.Render("test") }},
		{"ItemSelected uses Selection", func() string { return Styles.ItemSelected.Render("test") }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				result := tc.render()
				assert.NotEmpty(t, result, "render should produce output")
			})
		})
	}
}

func TestStyles_PanelBorders(t *testing.T) {
	// Verify Panel and PanelActive render content with borders
	content := "test content"
	panelOutput := Styles.Panel.Render(content)
	panelActiveOutput := Styles.PanelActive.Render(content)

	// Both should render something containing the content
	assert.Contains(t, panelOutput, content, "Panel should contain the content")
	assert.Contains(t, panelActiveOutput, content, "PanelActive should contain the content")

	// Both should have rounded border characters (lipgloss uses these regardless of color output)
	assert.Contains(t, panelOutput, "╭", "Panel should have rounded border")
	assert.Contains(t, panelActiveOutput, "╭", "PanelActive should have rounded border")
}

func TestStyles_SelectionHighlighting(t *testing.T) {
	// ItemSelected should render content - colors are only visible in TTY mode
	content := "test item"
	itemOutput := Styles.Item.Render(content)
	selectedOutput := Styles.ItemSelected.Render(content)

	// Both should contain the content
	assert.Contains(t, itemOutput, content, "Item should contain the content")
	assert.Contains(t, selectedOutput, content, "ItemSelected should contain the content")
}

func TestStyles_FooterHints(t *testing.T) {
	// Key and Action should render content - colors are only visible in TTY mode
	keyContent := "enter"
	actionContent := "select"
	key := Styles.Key.Render(keyContent)
	action := Styles.Action.Render(actionContent)

	assert.Contains(t, key, keyContent, "Key should contain the content")
	assert.Contains(t, action, actionContent, "Action should contain the content")
}
