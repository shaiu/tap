package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNerdFontIcons(t *testing.T) {
	// Verify Nerd Font icons have expected symbols
	assert.Equal(t, "", NerdFontIcons.Script)
	assert.Equal(t, "󰉋", NerdFontIcons.Category)
	assert.Equal(t, "󰉋", NerdFontIcons.Folder)
	assert.Equal(t, "󰐕", NerdFontIcons.Pin)
	assert.Equal(t, "", NerdFontIcons.Search)
	assert.Equal(t, "", NerdFontIcons.Running)
	assert.Equal(t, "", NerdFontIcons.Success)
	assert.Equal(t, "", NerdFontIcons.Error)
	assert.Equal(t, "", NerdFontIcons.Arrow)
	assert.Equal(t, "", NerdFontIcons.Bash)
	assert.Equal(t, "", NerdFontIcons.Python)
	assert.Equal(t, "", NerdFontIcons.Zsh)
}

func TestASCIIIcons(t *testing.T) {
	// Verify ASCII fallback icons
	assert.Equal(t, "*", ASCIIIcons.Script)
	assert.Equal(t, "#", ASCIIIcons.Category)
	assert.Equal(t, "#", ASCIIIcons.Folder)
	assert.Equal(t, "^", ASCIIIcons.Pin)
	assert.Equal(t, "/", ASCIIIcons.Search)
	assert.Equal(t, ">", ASCIIIcons.Running)
	assert.Equal(t, "+", ASCIIIcons.Success)
	assert.Equal(t, "x", ASCIIIcons.Error)
	assert.Equal(t, ">", ASCIIIcons.Arrow)
	assert.Equal(t, "$", ASCIIIcons.Bash)
	assert.Equal(t, "P", ASCIIIcons.Python)
	assert.Equal(t, "%", ASCIIIcons.Zsh)
}

func TestIconSwitching(t *testing.T) {
	// Start with Nerd Fonts (default)
	UseNerdFontIcons()
	assert.True(t, IsUsingNerdFonts())
	assert.Equal(t, "", Icons.Script)

	// Switch to ASCII
	UseASCIIIcons()
	assert.False(t, IsUsingNerdFonts())
	assert.Equal(t, "*", Icons.Script)

	// Switch back to Nerd Fonts
	UseNerdFontIcons()
	assert.True(t, IsUsingNerdFonts())
	assert.Equal(t, "", Icons.Script)
}

func TestIconForShell(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		useNerd  bool
		expected string
	}{
		{"bash with nerd", "bash", true, ""},
		{"sh with nerd", "sh", true, ""},
		{"python with nerd", "python", true, ""},
		{"python3 with nerd", "python3", true, ""},
		{"zsh with nerd", "zsh", true, ""},
		{"unknown with nerd", "fish", true, ""},
		{"bash with ascii", "bash", false, "$"},
		{"python with ascii", "python", false, "P"},
		{"zsh with ascii", "zsh", false, "%"},
		{"unknown with ascii", "fish", false, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useNerd {
				UseNerdFontIcons()
			} else {
				UseASCIIIcons()
			}
			assert.Equal(t, tt.expected, IconForShell(tt.shell))
		})
	}

	// Restore default
	UseNerdFontIcons()
}

func TestDefaultIconsAreNerdFonts(t *testing.T) {
	// Reset to default
	Icons = NerdFontIcons
	assert.True(t, IsUsingNerdFonts())
}
