package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FeedbackType represents the type of feedback message.
type FeedbackType int

const (
	FeedbackNone FeedbackType = iota
	FeedbackSuccess
	FeedbackError
	FeedbackRunning
)

// FeedbackMsg is sent to display a temporary feedback message.
type FeedbackMsg struct {
	Type    FeedbackType
	Message string
}

// ClearFeedbackMsg is sent to clear the feedback message.
type ClearFeedbackMsg struct{}

// FeedbackDuration is how long feedback messages are displayed.
const FeedbackDuration = 2 * time.Second

// FooterContext represents the current context for displaying footer hints.
type FooterContext struct {
	State       ViewState
	ActivePanel Panel
	LayoutMode  LayoutMode
	HasParams   bool // Whether the selected script has parameters
}

// KeyHint represents a single key-action pair for the footer.
type KeyHint struct {
	Key    string
	Action string
}

// FooterModel manages the footer bar display.
type FooterModel struct {
	context      FooterContext
	width        int
	feedbackType FeedbackType
	feedbackMsg  string
}

// NewFooterModel creates a new footer model.
func NewFooterModel() FooterModel {
	return FooterModel{
		context: FooterContext{
			State:       StateBrowsing,
			ActivePanel: PanelSidebar,
			LayoutMode:  LayoutThreePanel,
			HasParams:   false,
		},
		width: 80,
	}
}

// SetContext updates the footer context.
func (m *FooterModel) SetContext(ctx FooterContext) {
	m.context = ctx
}

// SetWidth sets the footer width.
func (m *FooterModel) SetWidth(width int) {
	m.width = width
}

// SetFeedback sets a temporary feedback message.
func (m *FooterModel) SetFeedback(feedbackType FeedbackType, message string) {
	m.feedbackType = feedbackType
	m.feedbackMsg = message
}

// ClearFeedback removes the feedback message.
func (m *FooterModel) ClearFeedback() {
	m.feedbackType = FeedbackNone
	m.feedbackMsg = ""
}

// HasFeedback returns true if there's an active feedback message.
func (m FooterModel) HasFeedback() bool {
	return m.feedbackType != FeedbackNone
}

// ClearFeedbackAfter returns a command that clears feedback after the duration.
func ClearFeedbackAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return ClearFeedbackMsg{}
	})
}

// hintsForContext returns the appropriate hints for the current context.
func (m FooterModel) hintsForContext() []KeyHint {
	switch m.context.State {
	case StateFilter:
		return m.filterHints()
	case StateHelp:
		return m.helpHints()
	case StateForm:
		return m.formHints()
	case StateBrowsing:
		return m.browsingHints()
	case StateCategoryList, StateScriptList:
		return m.legacyMenuHints()
	default:
		return m.defaultHints()
	}
}

// browsingHints returns hints for the 3-panel browsing mode.
func (m FooterModel) browsingHints() []KeyHint {
	hints := []KeyHint{
		{Key: "↑↓", Action: "navigate"},
	}

	// Context-aware action based on active panel
	switch m.context.ActivePanel {
	case PanelSidebar:
		hints = append(hints, KeyHint{Key: "enter", Action: "select"})
	case PanelScripts:
		hints = append(hints, KeyHint{Key: "enter", Action: "run"})
	case PanelDetails:
		hints = append(hints, KeyHint{Key: "enter", Action: "run"})
	}

	// Show tab hint if we have multiple panels
	if m.context.LayoutMode != LayoutOnePanel {
		hints = append(hints, KeyHint{Key: "tab", Action: "panel"})
	}

	hints = append(hints, KeyHint{Key: "/", Action: "filter"})

	// Show view/edit hints when a script is selectable
	if m.context.ActivePanel == PanelScripts || m.context.ActivePanel == PanelDetails {
		hints = append(hints, KeyHint{Key: "v", Action: "view"})
		hints = append(hints, KeyHint{Key: "e", Action: "edit"})
	}

	// Show params hint if selected script has parameters
	if m.context.HasParams && m.context.ActivePanel == PanelScripts {
		hints = append(hints, KeyHint{Key: "p", Action: "params"})
	}

	hints = append(hints,
		KeyHint{Key: "?", Action: "help"},
		KeyHint{Key: "q", Action: "quit"},
	)

	return hints
}

// filterHints returns hints for the filter state.
func (m FooterModel) filterHints() []KeyHint {
	return []KeyHint{
		{Key: "enter", Action: "select"},
		{Key: "esc", Action: "cancel"},
	}
}

// helpHints returns hints for the help state.
func (m FooterModel) helpHints() []KeyHint {
	return []KeyHint{
		{Key: "any key", Action: "close"},
	}
}

// formHints returns hints for the form state.
func (m FooterModel) formHints() []KeyHint {
	return []KeyHint{
		{Key: "↑↓", Action: "navigate"},
		{Key: "enter", Action: "submit"},
		{Key: "esc", Action: "cancel"},
	}
}

// legacyMenuHints returns hints for the legacy category/script list views.
func (m FooterModel) legacyMenuHints() []KeyHint {
	hints := []KeyHint{
		{Key: "↑↓", Action: "navigate"},
		{Key: "enter", Action: "select"},
	}

	if m.context.State == StateScriptList {
		hints = append(hints, KeyHint{Key: "esc", Action: "back"})
	}

	hints = append(hints,
		KeyHint{Key: "/", Action: "filter"},
		KeyHint{Key: "?", Action: "help"},
		KeyHint{Key: "q", Action: "quit"},
	)

	return hints
}

// defaultHints returns default hints for unknown states.
func (m FooterModel) defaultHints() []KeyHint {
	return []KeyHint{
		{Key: "↑↓", Action: "navigate"},
		{Key: "enter", Action: "select"},
		{Key: "?", Action: "help"},
		{Key: "q", Action: "quit"},
	}
}

// View renders the footer bar.
func (m FooterModel) View() string {
	// If there's feedback, show it instead of/alongside hints
	if m.feedbackType != FeedbackNone {
		return m.renderFeedback()
	}

	hints := m.hintsForContext()
	return m.renderHints(hints)
}

// renderFeedback renders the feedback message with appropriate styling.
func (m FooterModel) renderFeedback() string {
	var icon string
	var style lipgloss.Style

	switch m.feedbackType {
	case FeedbackSuccess:
		icon = Icons.Success
		style = Styles.FeedbackSuccess
	case FeedbackError:
		icon = Icons.Error
		style = Styles.FeedbackError
	case FeedbackRunning:
		icon = Icons.Running
		style = Styles.FeedbackRunning
	default:
		return m.renderHints(m.hintsForContext())
	}

	content := style.Render(icon + " " + m.feedbackMsg)

	return Styles.Footer.Width(m.width).Render(content)
}

// renderHints formats the hints as a styled footer line.
func (m FooterModel) renderHints(hints []KeyHint) string {
	var parts []string
	for _, hint := range hints {
		parts = append(parts, formatKeyHint(hint.Key, hint.Action))
	}

	content := strings.Join(parts, "  ")

	// Apply footer style with width
	return Styles.Footer.Width(m.width).Render(content)
}

// formatKeyHint formats a single key-action pair with styling.
func formatKeyHint(key, action string) string {
	return Styles.Key.Render(key) + " " + Styles.Action.Render(action)
}

// RenderSimple renders a simple footer without the full context.
// Useful for overlays that need quick footer rendering.
func RenderSimple(hints []KeyHint, width int) string {
	var parts []string
	for _, hint := range hints {
		parts = append(parts, formatKeyHint(hint.Key, hint.Action))
	}

	content := strings.Join(parts, "  ")
	return Styles.Footer.Width(width).Render(content)
}

// RenderCompact renders hints without the footer border/padding.
// Useful for inline footer display within panels.
func RenderCompact(hints []KeyHint) string {
	var parts []string
	for _, hint := range hints {
		parts = append(parts, formatKeyHint(hint.Key, hint.Action))
	}
	return strings.Join(parts, "  ")
}

// HintsForState returns the appropriate hints for a given state.
// This is a convenience function for use without creating a FooterModel.
func HintsForState(state ViewState, panel Panel, layout LayoutMode, hasParams bool) []KeyHint {
	m := FooterModel{
		context: FooterContext{
			State:       state,
			ActivePanel: panel,
			LayoutMode:  layout,
			HasParams:   hasParams,
		},
	}
	return m.hintsForContext()
}

// FilterFooter returns the standard filter mode footer.
func FilterFooter(width int) string {
	return RenderSimple([]KeyHint{
		{Key: "enter", Action: "select"},
		{Key: "esc", Action: "cancel"},
	}, width)
}

// HelpFooter returns the standard help mode footer.
func HelpFooter(width int) string {
	return RenderSimple([]KeyHint{
		{Key: "any key", Action: "close"},
	}, width)
}

// FormFooter returns the standard form mode footer.
func FormFooter(width int) string {
	return RenderSimple([]KeyHint{
		{Key: "↑↓", Action: "navigate"},
		{Key: "enter", Action: "submit"},
		{Key: "esc", Action: "cancel"},
	}, width)
}

// FooterHeight returns the number of lines the footer occupies.
// The footer uses a top border plus one line of content.
func FooterHeight() int {
	return 2
}

// GetFooterStyle returns the footer style for external use.
// This allows consistent styling when rendering footers manually.
func GetFooterStyle() lipgloss.Style {
	return Styles.Footer
}
