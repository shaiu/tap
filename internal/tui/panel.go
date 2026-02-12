package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Panel dimension constants for bordered panels with padding.
// These account for lipgloss RoundedBorder and Padding(0, 1).
const (
	// BorderWidth is the total horizontal space taken by borders (1 left + 1 right)
	BorderWidth = 2

	// PaddingWidth is the total horizontal space taken by padding (1 left + 1 right)
	PaddingWidth = 2

	// BorderHeight is the total vertical space taken by borders (1 top + 1 bottom)
	BorderHeight = 2

	// MinPanelWidth is the minimum usable width for a panel
	MinPanelWidth = 20

	// MinPanelHeight is the minimum usable height for a panel
	MinPanelHeight = 5
)

// InnerDimensions calculates the content area dimensions given outer panel dimensions.
// It accounts for borders and padding.
func InnerDimensions(outerWidth, outerHeight int) (innerWidth, innerHeight int) {
	innerWidth = outerWidth - BorderWidth - PaddingWidth
	innerHeight = outerHeight - BorderHeight

	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	return innerWidth, innerHeight
}

// PadLines pads or truncates a slice of lines to exactly targetHeight.
// If lines is shorter, empty lines are appended.
// If lines is longer, excess lines are truncated.
func PadLines(lines []string, targetHeight int) []string {
	if targetHeight <= 0 {
		return []string{}
	}

	result := make([]string, targetHeight)

	for i := 0; i < targetHeight; i++ {
		if i < len(lines) {
			result[i] = lines[i]
		} else {
			result[i] = ""
		}
	}

	return result
}

// TruncateLine truncates a line to maxWidth, adding ellipsis if truncated.
// It accounts for ANSI escape sequences when measuring width.
func TruncateLine(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// Use lipgloss width which handles ANSI codes
	width := lipgloss.Width(line)
	if width <= maxWidth {
		return line
	}

	// For lines with ANSI codes, we need to be careful
	// Simple approach: truncate the visible content
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	// Strip ANSI and truncate, then note that we lose styling
	// For styled content, the caller should handle truncation before styling
	runes := []rune(lipgloss.NewStyle().Render(line))
	if len(runes) <= maxWidth {
		return line
	}

	return string(runes[:maxWidth-3]) + "..."
}

// TruncateLineSimple truncates a plain string (no ANSI) to maxWidth with ellipsis.
func TruncateLineSimple(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(line)
	if len(runes) <= maxWidth {
		return line
	}

	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}

	return string(runes[:maxWidth-3]) + "..."
}

// BuildPanelContent joins lines and ensures the result fits the target height.
// This is the main function panels should use to prepare their content.
func BuildPanelContent(lines []string, targetHeight int) string {
	padded := PadLines(lines, targetHeight)
	return strings.Join(padded, "\n")
}
