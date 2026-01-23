package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestTheme_ColorsAreDefined(t *testing.T) {
	// Base colors
	assert.NotEmpty(t, string(Theme.Background), "Background should be defined")
	assert.NotEmpty(t, string(Theme.Foreground), "Foreground should be defined")
	assert.NotEmpty(t, string(Theme.Subtle), "Subtle should be defined")

	// Accent colors
	assert.NotEmpty(t, string(Theme.Primary), "Primary should be defined")
	assert.NotEmpty(t, string(Theme.Secondary), "Secondary should be defined")
	assert.NotEmpty(t, string(Theme.Success), "Success should be defined")
	assert.NotEmpty(t, string(Theme.Warning), "Warning should be defined")
	assert.NotEmpty(t, string(Theme.Error), "Error should be defined")

	// UI Elements
	assert.NotEmpty(t, string(Theme.Border), "Border should be defined")
	assert.NotEmpty(t, string(Theme.BorderActive), "BorderActive should be defined")
	assert.NotEmpty(t, string(Theme.Selection), "Selection should be defined")

	// Gradient
	assert.NotEmpty(t, string(Theme.GradientStart), "GradientStart should be defined")
	assert.NotEmpty(t, string(Theme.GradientEnd), "GradientEnd should be defined")
}

func TestTheme_CatppuccinMochaValues(t *testing.T) {
	// Verify the Catppuccin Mocha palette values
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"Background", Theme.Background, "#1e1e2e"},
		{"Foreground", Theme.Foreground, "#cdd6f4"},
		{"Subtle", Theme.Subtle, "#6c7086"},
		{"Primary", Theme.Primary, "#89b4fa"},
		{"Secondary", Theme.Secondary, "#cba6f7"},
		{"Success", Theme.Success, "#a6e3a1"},
		{"Warning", Theme.Warning, "#f9e2af"},
		{"Error", Theme.Error, "#f38ba8"},
		{"Border", Theme.Border, "#6c7086"},
		{"BorderActive", Theme.BorderActive, "#b4befe"},
		{"Selection", Theme.Selection, "#313244"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.color),
				"%s color should match Catppuccin Mocha palette", tt.name)
		})
	}
}

func TestTheme_ColorsAreValidHex(t *testing.T) {
	colors := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Background", Theme.Background},
		{"Foreground", Theme.Foreground},
		{"Subtle", Theme.Subtle},
		{"Primary", Theme.Primary},
		{"Secondary", Theme.Secondary},
		{"Success", Theme.Success},
		{"Warning", Theme.Warning},
		{"Error", Theme.Error},
		{"Border", Theme.Border},
		{"BorderActive", Theme.BorderActive},
		{"Selection", Theme.Selection},
		{"GradientStart", Theme.GradientStart},
		{"GradientEnd", Theme.GradientEnd},
	}

	for _, c := range colors {
		t.Run(c.name, func(t *testing.T) {
			hex := string(c.color)
			assert.Regexp(t, `^#[0-9a-fA-F]{6}$`, hex,
				"%s should be a valid 6-character hex color", c.name)
		})
	}
}

func TestTheme_CanBeUsedInStyles(t *testing.T) {
	// Verify theme colors work with lipgloss styles
	style := lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Background(Theme.Selection).
		BorderForeground(Theme.BorderActive)

	// Should not panic and should render text
	rendered := style.Render("test")
	assert.NotEmpty(t, rendered, "Style with theme colors should render text")
}
