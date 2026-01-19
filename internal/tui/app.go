package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

	// Menu
	menu MenuModel

	// Filter
	filterInput textinput.Model

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

	// Default dimensions (will be updated on WindowSizeMsg)
	width, height := 80, 24

	// Create filter input
	fi := textinput.New()
	fi.Placeholder = "Filter..."
	fi.CharLimit = 50
	fi.Width = 30

	return AppModel{
		state:          state,
		categories:     categories,
		menu:           NewMenuModel(categories, width, height),
		filterInput:    fi,
		selectedCatIdx: -1,
		width:          width,
		height:         height,
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
		m.menu.SetSize(msg.Width, msg.Height)
		return m, nil

	case ScriptsLoadedMsg:
		m.categories = msg.Categories
		m.menu.SetCategories(msg.Categories)
		if len(m.categories) > 0 {
			m.state = StateCategoryList
		}
		return m, nil

	case ScriptSelectedMsg:
		m.selectedScript = &msg.Script
		return m, tea.Quit

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
	case key.Matches(msg, m.keys.Filter):
		m.prevState = m.state
		m.state = StateFilter
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Refresh):
		return m, func() tea.Msg { return RefreshMsg{} }
	}

	// Delegate to menu model for navigation
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	// Check if we've drilled into scripts
	if m.menu.ShowingScripts() {
		m.state = StateScriptList
		m.selectedCatIdx = m.menu.SelectedCategoryIndex()
	}

	return m, cmd
}

// updateScriptList handles input in the script list state.
func (m AppModel) updateScriptList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Filter):
		m.prevState = m.state
		m.state = StateFilter
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Refresh):
		return m, func() tea.Msg { return RefreshMsg{} }
	}

	// Delegate to menu model for navigation
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)

	// Check if we've navigated back to categories
	if !m.menu.ShowingScripts() {
		m.state = StateCategoryList
		m.selectedCatIdx = -1
	}

	return m, cmd
}

// updateFilter handles input in the filter state.
func (m AppModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Confirm filter - keep filtered results and exit filter mode
		m.filterInput.Blur()
		m.state = m.prevState
		return m, nil
	case tea.KeyEscape:
		// Cancel filter - restore full list and exit filter mode
		m.filterInput.Blur()
		m.filterInput.SetValue("")
		m.menu.ClearFilter()
		m.state = m.prevState
		return m, nil
	}

	// Update the text input
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	// Apply filter to menu
	m.menu.ApplyFilter(m.filterInput.Value())

	return m, cmd
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
	case StateFilter:
		return m.renderFilterView()
	default:
		// CategoryList, ScriptList views are rendered by MenuModel
		return m.menu.View()
	}
}

// renderFilterView renders the view with the filter input overlay.
func (m AppModel) renderFilterView() string {
	var s strings.Builder

	// Header
	if m.menu.ShowingScripts() && m.selectedCatIdx >= 0 && m.selectedCatIdx < len(m.categories) {
		s.WriteString(Styles.Header.Render("tap - " + m.categories[m.selectedCatIdx].Name))
	} else {
		s.WriteString(Styles.Header.Render("tap - Script Runner"))
	}
	s.WriteString("\n\n")

	// Filter input
	s.WriteString(Styles.FilterInput.Render("Filter: " + m.filterInput.View()))
	s.WriteString("\n\n")

	// Show filtered list from menu
	if m.menu.ShowingScripts() {
		s.WriteString(m.menu.scriptList.View())
	} else {
		s.WriteString(m.menu.categoryList.View())
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(Styles.Footer.Render("enter select  esc cancel"))

	return s.String()
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
	m.menu.SetCategories(categories)
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
