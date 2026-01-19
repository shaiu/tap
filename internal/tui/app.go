package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

// ViewState represents the current state of the TUI.
type ViewState int

const (
	StateLoading ViewState = iota
	StateCategoryList
	StateScriptList
	StateFilter
	StateHelp
)

// String returns a human-readable name for the view state.
func (s ViewState) String() string {
	switch s {
	case StateLoading:
		return "loading"
	case StateCategoryList:
		return "category-list"
	case StateScriptList:
		return "script-list"
	case StateFilter:
		return "filter"
	case StateHelp:
		return "help"
	default:
		return "unknown"
	}
}

// Messages for the TUI.

// ScriptSelectedMsg is sent when a script is chosen for execution.
type ScriptSelectedMsg struct {
	Script core.Script
}

// ScriptsLoadedMsg is sent when scripts have been loaded.
type ScriptsLoadedMsg struct {
	Categories []core.Category
}

// RefreshMsg triggers a rescan of script directories.
type RefreshMsg struct{}

// ErrorMsg carries error information.
type ErrorMsg struct {
	Err error
}

// AppModel is the root model for the tap TUI.
type AppModel struct {
	// State
	state      ViewState
	prevState  ViewState // State before filter/help
	categories []core.Category

	// Selection
	selectedCatIdx int
	selectedScript *core.Script

	// Dimensions
	width  int
	height int

	// Key bindings
	keys KeyMap

	// Error handling
	err error
}

// NewAppModel creates a new AppModel with the given categories.
func NewAppModel(categories []core.Category) AppModel {
	state := StateCategoryList
	if len(categories) == 0 {
		state = StateLoading
	}

	return AppModel{
		state:          state,
		categories:     categories,
		selectedCatIdx: -1,
		keys:           DefaultKeyMap(),
	}
}

// Init implements tea.Model.
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ScriptsLoadedMsg:
		m.categories = msg.Categories
		if len(m.categories) > 0 {
			m.state = StateCategoryList
		}
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

// handleKeyMsg processes keyboard input.
func (m AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys that work regardless of state
	switch {
	case key.Matches(msg, m.keys.Quit):
		if m.state != StateFilter {
			return m, tea.Quit
		}
	case key.Matches(msg, m.keys.Help):
		if m.state != StateFilter && m.state != StateHelp {
			m.prevState = m.state
			m.state = StateHelp
			return m, nil
		}
	}

	// State-specific handling
	switch m.state {
	case StateCategoryList:
		return m.updateCategoryList(msg)
	case StateScriptList:
		return m.updateScriptList(msg)
	case StateFilter:
		return m.updateFilter(msg)
	case StateHelp:
		return m.updateHelp(msg)
	}

	return m, nil
}

// updateCategoryList handles input in the category list state.
func (m AppModel) updateCategoryList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		// Will be handled by MenuModel in next task
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.prevState = m.state
		m.state = StateFilter
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		return m, func() tea.Msg { return RefreshMsg{} }
	}
	return m, nil
}

// updateScriptList handles input in the script list state.
func (m AppModel) updateScriptList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		// Will emit ScriptSelectedMsg when MenuModel is implemented
		return m, nil
	case key.Matches(msg, m.keys.Back):
		m.state = StateCategoryList
		m.selectedCatIdx = -1
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.prevState = m.state
		m.state = StateFilter
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		return m, func() tea.Msg { return RefreshMsg{} }
	}
	return m, nil
}

// updateFilter handles input in the filter state.
func (m AppModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select):
		// Confirm filter, stay in previous state's list
		m.state = m.prevState
		return m, nil
	case key.Matches(msg, m.keys.Back):
		// Cancel filter, restore state
		m.state = m.prevState
		return m, nil
	}
	return m, nil
}

// updateHelp handles input in the help state.
func (m AppModel) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key exits help
	m.state = m.prevState
	return m, nil
}

// View implements tea.Model.
func (m AppModel) View() string {
	if m.err != nil {
		return Styles.Error.Render("Error: " + m.err.Error())
	}

	switch m.state {
	case StateLoading:
		return "Loading scripts..."
	case StateHelp:
		return m.renderHelp()
	default:
		// CategoryList, ScriptList, Filter views will be rendered by MenuModel
		return m.renderPlaceholder()
	}
}

// renderPlaceholder renders a placeholder view until MenuModel is implemented.
func (m AppModel) renderPlaceholder() string {
	header := Styles.Header.Render("tap - Script Runner")
	footer := Styles.Footer.Render("↑/↓ navigate  enter select  / filter  q quit")

	content := "Categories:\n"
	for i, cat := range m.categories {
		prefix := "  "
		if i == 0 {
			prefix = "> "
		}
		content += Styles.Dimmed.Render(prefix+cat.Name) + "\n"
	}

	return header + "\n\n" + content + "\n" + footer
}

// renderHelp renders the help overlay.
func (m AppModel) renderHelp() string {
	help := `
Keyboard Shortcuts

Navigation
  ↑/k       Move up
  ↓/j       Move down
  enter     Select / Run
  esc       Go back

Filtering
  /         Start filtering
  esc       Cancel filter
  enter     Select match

Other
  ?         Show this help
  r         Refresh scripts
  q         Quit

Press any key to close...
`
	return Styles.Help.Render(help)
}

// State returns the current view state.
func (m AppModel) State() ViewState {
	return m.state
}

// Categories returns the loaded categories.
func (m AppModel) Categories() []core.Category {
	return m.categories
}

// SelectedScript returns the selected script, if any.
func (m AppModel) SelectedScript() *core.Script {
	return m.selectedScript
}

// SetCategories updates the categories (for testing).
func (m *AppModel) SetCategories(categories []core.Category) {
	m.categories = categories
	if len(categories) > 0 && m.state == StateLoading {
		m.state = StateCategoryList
	}
}

// Width returns the terminal width.
func (m AppModel) Width() int {
	return m.width
}

// Height returns the terminal height.
func (m AppModel) Height() int {
	return m.height
}
