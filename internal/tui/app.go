package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	StateForm
	StateBrowsing // 3-panel browsing mode
)

// Panel represents which panel is currently active in 3-panel mode.
type Panel int

const (
	PanelSidebar Panel = iota
	PanelScripts
	PanelDetails
)

// LayoutMode represents the responsive layout based on terminal width.
type LayoutMode int

const (
	LayoutThreePanel LayoutMode = iota // ≥120 cols
	LayoutTwoPanel                     // 80-119 cols
	LayoutOnePanel                     // <80 cols
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
	case StateForm:
		return "form"
	case StateBrowsing:
		return "browsing"
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

	// Legacy Menu (for backward compatibility)
	menu MenuModel

	// 3-Panel Layout Components
	sidebar     SidebarModel
	scriptsPane ScriptsModel
	detailsPane DetailsModel
	activePanel Panel
	layoutMode  LayoutMode

	// Footer
	footer FooterModel

	// Form (for parameter input)
	formModel FormModel

	// Filter
	filterInput   textinput.Model
	filterOverlay FilterModel

	// Loading spinner
	loadingSpinner spinner.Model

	// Selection
	selectedCatIdx   int
	selectedScript   *core.Script
	selectedParams   map[string]string // Parameters from form submission

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
	state := StateBrowsing
	if len(categories) == 0 {
		state = StateLoading
	}

	// Default dimensions (will be updated on WindowSizeMsg)
	width, height := 80, 24
	layoutMode := calculateLayoutMode(width)

	// Create filter input (legacy)
	fi := textinput.New()
	fi.Placeholder = "Filter..."
	fi.CharLimit = 50
	fi.Width = 30

	// Create filter overlay model
	filterOverlay := NewFilterModel()
	filterOverlay.SetSize(width, height)

	// Create loading spinner (dots style for modern look)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(Theme.Primary)

	// Create 3-panel components
	sidebar := NewSidebarModel(categories)
	scriptsPane := NewScriptsModel()
	detailsPane := NewDetailsModel()

	// Determine initial active panel based on layout mode
	// In 1-panel mode, only scripts are visible
	initialPanel := PanelSidebar
	if layoutMode == LayoutOnePanel {
		initialPanel = PanelScripts
	}

	// Create footer
	footer := NewFooterModel()
	footer.SetWidth(width)
	footer.SetContext(FooterContext{
		State:       state,
		ActivePanel: initialPanel,
		LayoutMode:  layoutMode,
		HasParams:   false,
	})

	// Initialize scripts based on sidebar selection
	if len(categories) > 0 {
		// Start with "All Scripts" selected
		scriptsPane.SetAllScripts(categories)
	}

	// Set initial focus based on layout
	sidebar.SetFocused(initialPanel == PanelSidebar)
	scriptsPane.SetFocused(initialPanel == PanelScripts)
	detailsPane.SetFocused(false)

	return AppModel{
		state:          state,
		categories:     categories,
		menu:           NewMenuModel(categories, width, height),
		sidebar:        sidebar,
		scriptsPane:    scriptsPane,
		detailsPane:    detailsPane,
		activePanel:    initialPanel,
		layoutMode:     layoutMode,
		footer:         footer,
		filterInput:    fi,
		filterOverlay:  filterOverlay,
		loadingSpinner: sp,
		selectedCatIdx: -1,
		width:          width,
		height:         height,
		keys:           DefaultKeyMap(),
	}
}

// calculateLayoutMode determines the layout mode based on terminal width.
func calculateLayoutMode(width int) LayoutMode {
	if width >= 120 {
		return LayoutThreePanel
	}
	if width >= 80 {
		return LayoutTwoPanel
	}
	return LayoutOnePanel
}

// updatePanelSizes calculates and sets panel sizes based on layout mode.
func (m *AppModel) updatePanelSizes() {
	// Reserve space for footer (2 lines)
	contentHeight := m.height - 3
	if contentHeight < 5 {
		contentHeight = 5
	}

	switch m.layoutMode {
	case LayoutThreePanel:
		// 3-panel: sidebar (22%) | scripts (43%) | details (35%)
		sidebarWidth := m.width * 22 / 100
		detailsWidth := m.width * 35 / 100
		scriptsWidth := m.width - sidebarWidth - detailsWidth - 2 // -2 for gaps

		if sidebarWidth < 20 {
			sidebarWidth = 20
		}
		if detailsWidth < 25 {
			detailsWidth = 25
		}
		if scriptsWidth < 25 {
			scriptsWidth = m.width - sidebarWidth - detailsWidth - 2
		}

		m.sidebar.SetSize(sidebarWidth, contentHeight)
		m.scriptsPane.SetSize(scriptsWidth, contentHeight)
		m.detailsPane.SetSize(detailsWidth, contentHeight)

	case LayoutTwoPanel:
		// 2-panel: sidebar (30%) | scripts (70%)
		sidebarWidth := m.width * 30 / 100
		scriptsWidth := m.width - sidebarWidth - 1

		if sidebarWidth < 20 {
			sidebarWidth = 20
		}
		if scriptsWidth < 30 {
			scriptsWidth = m.width - sidebarWidth - 1
		}

		m.sidebar.SetSize(sidebarWidth, contentHeight)
		m.scriptsPane.SetSize(scriptsWidth, contentHeight)
		m.detailsPane.SetSize(0, 0) // Hidden

	case LayoutOnePanel:
		// 1-panel: scripts only
		m.sidebar.SetSize(0, 0)      // Hidden
		m.scriptsPane.SetSize(m.width, contentHeight)
		m.detailsPane.SetSize(0, 0) // Hidden
	}
}

// Init implements tea.Model.
func (m AppModel) Init() tea.Cmd {
	if m.state == StateLoading {
		return m.loadingSpinner.Tick
	}
	return nil
}

// Update implements tea.Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutMode = calculateLayoutMode(msg.Width)
		m.menu.SetSize(msg.Width, msg.Height)

		// Update panel sizes based on layout mode
		m.updatePanelSizes()

		// Ensure active panel is valid for new layout mode
		m.ensureValidActivePanel()

		// Update filter overlay size
		m.filterOverlay.SetSize(msg.Width, msg.Height)

		// Update footer context (layout may have changed)
		m.updateFooterContext()

		if m.state == StateForm {
			model, cmd := m.formModel.Update(msg)
			if fm, ok := model.(FormModel); ok {
				m.formModel = fm
			}
			return m, cmd
		}
		return m, nil

	case ScriptsLoadedMsg:
		m.categories = msg.Categories
		m.menu.SetCategories(msg.Categories)
		m.sidebar.SetCategories(msg.Categories)
		// Initialize scripts based on sidebar selection
		if len(m.categories) > 0 {
			m.scriptsPane.SetAllScripts(m.categories)
			if script := m.scriptsPane.SelectedScript(); script != nil {
				m.detailsPane.SetScript(script)
			}
			m.state = StateBrowsing
		}
		return m, nil

	case ScriptSelectedMsg:
		if len(msg.Script.Parameters) > 0 {
			// Script has params - show form
			m.state = StateForm
			m.formModel = NewFormModel(msg.Script, m.width, m.height)
			return m, m.formModel.Init()
		}
		// No params - execute directly
		m.selectedScript = &msg.Script
		return m, tea.Quit

	case FormSubmittedMsg:
		// Form completed - execute with parameters
		m.selectedScript = &msg.Script
		m.selectedParams = msg.Parameters
		return m, tea.Quit

	case FormCancelledMsg:
		// Return to browsing
		m.state = StateBrowsing
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, nil

	case spinner.TickMsg:
		// Update spinner animation during loading
		if m.state == StateLoading {
			var cmd tea.Cmd
			m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
			return m, cmd
		}
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
			m.updateFooterContext()
			return m, nil
		}
	}

	// State-specific handling
	switch m.state {
	case StateBrowsing:
		return m.updateBrowsing(msg)
	case StateCategoryList:
		return m.updateCategoryList(msg)
	case StateScriptList:
		return m.updateScriptList(msg)
	case StateFilter:
		return m.updateFilter(msg)
	case StateHelp:
		return m.updateHelp(msg)
	case StateForm:
		return m.updateForm(msg)
	}

	return m, nil
}

// updateBrowsing handles input in the 3-panel browsing state.
func (m AppModel) updateBrowsing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle panel switching
	switch {
	case key.Matches(msg, m.keys.NextPanel):
		m.switchToNextPanel()
		return m, nil
	case key.Matches(msg, m.keys.PrevPanel):
		m.switchToPrevPanel()
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.prevState = m.state
		m.state = StateFilter
		m.updateFooterContext()
		// Enable filter mode on scripts pane (shows all items with dimming)
		m.scriptsPane.SetFilterMode(true)
		// Reset and focus filter overlay
		m.filterOverlay.Reset()
		m.filterOverlay.SetCounts(m.scriptsPane.TotalCount(), m.scriptsPane.TotalCount())
		// Also reset legacy filter input for backward compatibility
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		return m, m.filterOverlay.Focus()
	case key.Matches(msg, m.keys.Refresh):
		return m, func() tea.Msg { return RefreshMsg{} }
	case key.Matches(msg, m.keys.Select):
		// Enter key - handle based on active panel
		return m.handlePanelSelect()
	}

	// Route to active panel
	return m.updateActivePanel(msg)
}

// switchToNextPanel moves focus to the next panel.
func (m *AppModel) switchToNextPanel() {
	minPanel := m.minPanelForLayout()
	maxPanel := m.maxPanelForLayout()

	// In 1-panel mode, don't switch panels
	if minPanel == maxPanel {
		return
	}

	m.setAllPanelsUnfocused()

	if m.activePanel >= maxPanel {
		m.activePanel = minPanel
	} else {
		m.activePanel = Panel(int(m.activePanel) + 1)
	}

	m.setActivePanelFocused()
	m.updateFooterContext()
}

// switchToPrevPanel moves focus to the previous panel.
func (m *AppModel) switchToPrevPanel() {
	minPanel := m.minPanelForLayout()
	maxPanel := m.maxPanelForLayout()

	// In 1-panel mode, don't switch panels
	if minPanel == maxPanel {
		return
	}

	m.setAllPanelsUnfocused()

	if m.activePanel <= minPanel {
		m.activePanel = maxPanel
	} else {
		m.activePanel = Panel(int(m.activePanel) - 1)
	}

	m.setActivePanelFocused()
	m.updateFooterContext()
}

// maxPanelForLayout returns the max panel index for the current layout.
func (m AppModel) maxPanelForLayout() Panel {
	switch m.layoutMode {
	case LayoutThreePanel:
		return PanelDetails
	case LayoutTwoPanel:
		return PanelScripts
	default:
		// In 1-panel mode, only scripts panel is visible
		return PanelScripts
	}
}

// minPanelForLayout returns the min panel index for the current layout.
func (m AppModel) minPanelForLayout() Panel {
	switch m.layoutMode {
	case LayoutOnePanel:
		// In 1-panel mode, only scripts panel is visible
		return PanelScripts
	default:
		return PanelSidebar
	}
}

// ensureValidActivePanel adjusts the active panel if it's not visible in the current layout.
func (m *AppModel) ensureValidActivePanel() {
	minPanel := m.minPanelForLayout()
	maxPanel := m.maxPanelForLayout()

	if m.activePanel < minPanel {
		m.activePanel = minPanel
	} else if m.activePanel > maxPanel {
		m.activePanel = maxPanel
	}

	m.setAllPanelsUnfocused()
	m.setActivePanelFocused()
}

// setAllPanelsUnfocused removes focus from all panels.
func (m *AppModel) setAllPanelsUnfocused() {
	m.sidebar.SetFocused(false)
	m.scriptsPane.SetFocused(false)
	m.detailsPane.SetFocused(false)
}

// setActivePanelFocused sets focus on the active panel.
func (m *AppModel) setActivePanelFocused() {
	switch m.activePanel {
	case PanelSidebar:
		m.sidebar.SetFocused(true)
	case PanelScripts:
		m.scriptsPane.SetFocused(true)
	case PanelDetails:
		m.detailsPane.SetFocused(true)
	}
}

// handlePanelSelect handles the enter key based on active panel.
func (m AppModel) handlePanelSelect() (tea.Model, tea.Cmd) {
	switch m.activePanel {
	case PanelSidebar:
		// Update scripts panel based on sidebar selection
		if m.sidebar.IsAllScriptsSelected() {
			m.scriptsPane.SetAllScripts(m.categories)
		} else if cat := m.sidebar.SelectedCategory(); cat != nil {
			m.scriptsPane.SetScripts(cat.Scripts)
		}
		// Move focus to scripts panel
		m.setAllPanelsUnfocused()
		m.activePanel = PanelScripts
		m.scriptsPane.SetFocused(true)
		// Update details with first script
		if script := m.scriptsPane.SelectedScript(); script != nil {
			m.detailsPane.SetScript(script)
		}
		return m, nil

	case PanelScripts:
		// Run the selected script
		if script := m.scriptsPane.SelectedScript(); script != nil {
			if len(script.Parameters) > 0 {
				// Script has params - show form
				m.state = StateForm
				m.formModel = NewFormModel(*script, m.width, m.height)
				return m, m.formModel.Init()
			}
			// No params - execute directly
			m.selectedScript = script
			return m, tea.Quit
		}
		return m, nil

	case PanelDetails:
		// In details panel, enter runs the script
		if script := m.detailsPane.Script(); script != nil {
			if len(script.Parameters) > 0 {
				m.state = StateForm
				m.formModel = NewFormModel(*script, m.width, m.height)
				return m, m.formModel.Init()
			}
			m.selectedScript = script
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

// updateActivePanel routes key messages to the active panel.
func (m AppModel) updateActivePanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.activePanel {
	case PanelSidebar:
		m.sidebar, cmd = m.sidebar.Update(msg)
		// Update scripts when sidebar selection changes
		if m.sidebar.IsAllScriptsSelected() {
			m.scriptsPane.SetAllScripts(m.categories)
		} else if cat := m.sidebar.SelectedCategory(); cat != nil {
			m.scriptsPane.SetScripts(cat.Scripts)
		}
		// Update details panel
		if script := m.scriptsPane.SelectedScript(); script != nil {
			m.detailsPane.SetScript(script)
		} else {
			m.detailsPane.SetScript(nil)
		}
		// Update footer (hasParams may have changed)
		m.updateFooterContext()

	case PanelScripts:
		m.scriptsPane, cmd = m.scriptsPane.Update(msg)
		// Update details panel when script selection changes
		if script := m.scriptsPane.SelectedScript(); script != nil {
			m.detailsPane.SetScript(script)
		} else {
			m.detailsPane.SetScript(nil)
		}
		// Update footer (hasParams may have changed)
		m.updateFooterContext()

	case PanelDetails:
		m.detailsPane, cmd = m.detailsPane.Update(msg)
	}

	return m, cmd
}

// updateForm handles input in the form state.
func (m AppModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to form model
	var cmd tea.Cmd
	model, cmd := m.formModel.Update(msg)
	if fm, ok := model.(FormModel); ok {
		m.formModel = fm
	}
	return m, cmd
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
		// Confirm filter - keep filtered results and exit filter overlay mode
		m.filterOverlay.Blur()
		m.filterInput.Blur()
		// Disable filter mode (items are now hidden, not dimmed)
		m.scriptsPane.SetFilterMode(false)
		m.state = m.prevState
		m.updateFooterContext()
		return m, nil
	case tea.KeyEscape:
		// Cancel filter - restore full list and exit filter mode
		m.filterOverlay.Blur()
		m.filterOverlay.Reset()
		m.filterInput.Blur()
		m.filterInput.SetValue("")
		m.menu.ClearFilter()
		m.scriptsPane.ClearFilter()
		m.state = m.prevState
		m.updateFooterContext()
		return m, nil
	}

	// Update both the overlay input and the legacy filter input
	var cmd tea.Cmd
	m.filterOverlay, cmd = m.filterOverlay.Update(msg)
	m.filterInput, _ = m.filterInput.Update(msg)

	// Apply filter to menu and scripts pane
	query := m.filterOverlay.Value()
	m.menu.ApplyFilter(query)
	m.scriptsPane.ApplyFilter(query)

	// Update match counts in overlay
	m.filterOverlay.SetCounts(m.scriptsPane.ScriptCount(), m.scriptsPane.TotalCount())

	return m, cmd
}

// updateHelp handles input in the help state.
func (m AppModel) updateHelp(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key exits help
	m.state = m.prevState
	m.updateFooterContext()
	return m, nil
}

// View implements tea.Model.
func (m AppModel) View() string {
	if m.err != nil {
		return Styles.Error.Render("Error: " + m.err.Error())
	}

	switch m.state {
	case StateLoading:
		return m.renderLoadingView()
	case StateHelp:
		return m.renderHelp()
	case StateFilter:
		return m.renderFilterView()
	case StateForm:
		return m.formModel.View()
	case StateBrowsing:
		return m.renderBrowsingView()
	default:
		// Legacy: CategoryList, ScriptList views are rendered by MenuModel
		return m.menu.View()
	}
}

// renderLoadingView renders a centered loading spinner.
func (m AppModel) renderLoadingView() string {
	// Calculate content area dimensions
	contentWidth := m.width
	contentHeight := m.height - 3 // Reserve space for footer
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Create the loading message and spinner
	message := "Scanning scripts..."
	spinnerFrame := m.loadingSpinner.View()

	// Create centered content
	messageStyle := lipgloss.NewStyle().
		Foreground(Theme.Foreground)
	spinnerStyle := lipgloss.NewStyle().
		Foreground(Theme.Primary)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		messageStyle.Render(message),
		spinnerStyle.Render(spinnerFrame),
	)

	// Create the panel box
	panelWidth := contentWidth - 4 // Account for border padding
	panelHeight := contentHeight - 2

	// Calculate vertical padding to center content
	contentLines := 2 // message + spinner
	topPadding := (panelHeight - contentLines) / 2
	if topPadding < 1 {
		topPadding = 1
	}

	// Build vertically centered content
	var lines []string
	for i := 0; i < topPadding; i++ {
		lines = append(lines, "")
	}
	lines = append(lines, content)

	centeredContent := strings.Join(lines, "\n")

	// Create the panel with rounded border
	panel := Styles.PanelActive.
		Width(panelWidth).
		Height(panelHeight).
		Align(lipgloss.Center).
		Render(centeredContent)

	// Center the panel horizontally
	panelView := lipgloss.Place(
		contentWidth,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		panel,
	)

	// Add footer
	footer := m.footer.View()

	return panelView + "\n" + footer
}

// renderBrowsingView renders the 3-panel layout.
func (m AppModel) renderBrowsingView() string {
	var s strings.Builder

	// Render panels based on layout mode
	switch m.layoutMode {
	case LayoutThreePanel:
		s.WriteString(m.renderThreePanels())
	case LayoutTwoPanel:
		s.WriteString(m.renderTwoPanels())
	case LayoutOnePanel:
		s.WriteString(m.renderOnePanel())
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(m.renderBrowsingFooter())

	return s.String()
}

// renderThreePanels renders the full 3-panel layout.
func (m AppModel) renderThreePanels() string {
	sidebarView := m.sidebar.View()
	scriptsView := m.scriptsPane.View()
	detailsView := m.detailsPane.View()

	// Join panels horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, " ", scriptsView, " ", detailsView)
}

// renderTwoPanels renders the 2-panel layout (sidebar + scripts).
func (m AppModel) renderTwoPanels() string {
	sidebarView := m.sidebar.View()
	scriptsView := m.scriptsPane.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, " ", scriptsView)
}

// renderOnePanel renders the single-panel layout (scripts only).
func (m AppModel) renderOnePanel() string {
	// In 1-panel mode, show a compact category indicator at the top
	var s strings.Builder

	// Category header
	catHeader := m.sidebar.RenderCompact()
	s.WriteString(Styles.Header.Render("tap - " + catHeader))
	s.WriteString("\n\n")

	// Scripts list
	s.WriteString(m.scriptsPane.View())

	return s.String()
}

// renderBrowsingFooter renders the footer for browsing mode.
func (m AppModel) renderBrowsingFooter() string {
	return m.footer.View()
}

// updateFooterContext updates the footer context based on current state.
func (m *AppModel) updateFooterContext() {
	hasParams := false
	if script := m.scriptsPane.SelectedScript(); script != nil {
		hasParams = len(script.Parameters) > 0
	}

	m.footer.SetContext(FooterContext{
		State:       m.state,
		ActivePanel: m.activePanel,
		LayoutMode:  m.layoutMode,
		HasParams:   hasParams,
	})
	m.footer.SetWidth(m.width)
}

// renderFilterView renders the view with the filter input overlay.
func (m AppModel) renderFilterView() string {
	// For browsing mode, render the superfile-style overlay
	if m.prevState == StateBrowsing {
		return m.renderFilterOverlayView()
	}

	// Legacy filter view for category/script list modes
	var s strings.Builder

	// Header
	if m.menu.ShowingScripts() && m.selectedCatIdx >= 0 && m.selectedCatIdx < len(m.categories) {
		s.WriteString(Styles.Header.Render("tap - " + m.categories[m.selectedCatIdx].Name))
	} else {
		s.WriteString(Styles.Header.Render("tap - Script Runner"))
	}
	s.WriteString("\n\n")

	// Filter input with match count
	filterText := "Filter: " + m.filterInput.View()
	s.WriteString(Styles.FilterInput.Render(filterText))
	s.WriteString("\n\n")

	// Show filtered list
	if m.menu.ShowingScripts() {
		s.WriteString(m.menu.scriptList.View())
	} else {
		s.WriteString(m.menu.categoryList.View())
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(FilterFooter(m.width))

	return s.String()
}

// renderFilterOverlayView renders the 3-panel layout with the filter overlay on top.
func (m AppModel) renderFilterOverlayView() string {
	// Render the background (scripts panel with dimmed non-matches)
	var background string
	switch m.layoutMode {
	case LayoutThreePanel:
		background = m.renderThreePanels()
	case LayoutTwoPanel:
		background = m.renderTwoPanels()
	case LayoutOnePanel:
		background = m.renderOnePanel()
	}

	// Calculate content area dimensions (without footer)
	contentHeight := m.height - 3

	// Render the filter overlay box
	overlayBox := m.filterOverlay.View()
	overlayLines := strings.Split(overlayBox, "\n")
	overlayWidth := lipgloss.Width(overlayBox)

	// Center the overlay horizontally
	horizPadding := (m.width - overlayWidth) / 2
	if horizPadding < 0 {
		horizPadding = 0
	}

	// Position overlay near the top (about 1/5 down)
	vertPadding := contentHeight / 5
	if vertPadding < 2 {
		vertPadding = 2
	}

	// Composite the overlay on top of the background
	bgLines := strings.Split(background, "\n")

	// Ensure we have enough lines for the overlay
	requiredLines := vertPadding + len(overlayLines)
	if len(bgLines) < requiredLines {
		// Pad background with empty lines
		for len(bgLines) < requiredLines {
			bgLines = append(bgLines, "")
		}
	}

	result := make([]string, len(bgLines))
	copy(result, bgLines)

	indent := strings.Repeat(" ", horizPadding)
	for i, overlayLine := range overlayLines {
		lineIdx := vertPadding + i
		if lineIdx >= 0 && lineIdx < len(result) {
			// Overlay the filter box on this line
			result[lineIdx] = indent + overlayLine
		}
	}

	// Add footer
	var s strings.Builder
	for i, line := range result {
		s.WriteString(line)
		if i < len(result)-1 {
			s.WriteString("\n")
		}
	}

	// Ensure we have enough lines for footer
	currentLines := len(result)
	neededLines := contentHeight
	for currentLines < neededLines {
		s.WriteString("\n")
		currentLines++
	}

	s.WriteString("\n")
	s.WriteString(FilterFooter(m.width))

	return s.String()
}

// renderHelp renders the help overlay.
func (m AppModel) renderHelp() string {
	return RenderHelp(m.width, m.height)
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

// SelectedParams returns the parameters from form submission, if any.
func (m AppModel) SelectedParams() map[string]string {
	return m.selectedParams
}

// SetCategories updates the categories (for testing).
func (m *AppModel) SetCategories(categories []core.Category) {
	m.categories = categories
	m.menu.SetCategories(categories)
	m.sidebar.SetCategories(categories)
	if len(categories) > 0 {
		m.scriptsPane.SetAllScripts(categories)
		if script := m.scriptsPane.SelectedScript(); script != nil {
			m.detailsPane.SetScript(script)
		}
		if m.state == StateLoading {
			m.state = StateBrowsing
		}
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
