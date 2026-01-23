// Package tui provides the terminal user interface for tap.
package tui

// IconSet holds icon strings for UI elements.
// Supports both Nerd Font icons and ASCII fallbacks.
type IconSet struct {
	Script   string // Script file icon
	Category string // Category/folder icon
	Folder   string // Folder icon (alias for Category)
	Pin      string // Pinned item icon
	Search   string // Search/filter icon
	Running  string // Running/executing icon
	Success  string // Success/checkmark icon
	Error    string // Error/X icon
	Arrow    string // Arrow/chevron icon
	Bash     string // Bash shell icon
	Python   string // Python icon
	Zsh      string // Zsh shell icon
}

// NerdFontIcons uses Nerd Font symbols for a polished look.
// Requires a Nerd Font to be installed and active in the terminal.
var NerdFontIcons = IconSet{
	Script:   "",  // nf-fa-file_code
	Category: "󰉋", // nf-md-folder
	Folder:   "󰉋", // nf-md-folder
	Pin:      "󰐕", // nf-md-pin
	Search:   "",  // nf-fa-search
	Running:  "",  // nf-fa-play
	Success:  "",  // nf-fa-check
	Error:    "",  // nf-fa-times
	Arrow:    "",  // nf-fa-chevron_right
	Bash:     "",  // nf-dev-terminal
	Python:   "",  // nf-dev-python
	Zsh:      "",  // nf-dev-terminal
}

// ASCIIIcons uses ASCII characters for compatibility with any terminal.
var ASCIIIcons = IconSet{
	Script:   "*",
	Category: "#",
	Folder:   "#",
	Pin:      "^",
	Search:   "/",
	Running:  ">",
	Success:  "+",
	Error:    "x",
	Arrow:    ">",
	Bash:     "$",
	Python:   "P",
	Zsh:      "%",
}

// Icons is the currently active icon set.
// Defaults to Nerd Font icons; call UseASCIIIcons() to switch to fallback.
var Icons = NerdFontIcons

// UseNerdFontIcons switches to Nerd Font icons (the default).
func UseNerdFontIcons() {
	Icons = NerdFontIcons
}

// UseASCIIIcons switches to ASCII fallback icons.
// Use this when Nerd Fonts are not available or for maximum compatibility.
func UseASCIIIcons() {
	Icons = ASCIIIcons
}

// IsUsingNerdFonts returns true if currently using Nerd Font icons.
func IsUsingNerdFonts() bool {
	return Icons.Script == NerdFontIcons.Script
}

// IconForShell returns the appropriate icon for a shell type.
func IconForShell(shell string) string {
	switch shell {
	case "bash", "sh":
		return Icons.Bash
	case "python", "python3":
		return Icons.Python
	case "zsh":
		return Icons.Zsh
	default:
		return Icons.Script
	}
}
