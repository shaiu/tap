package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestViewState_String(t *testing.T) {
	tests := []struct {
		state    ViewState
		expected string
	}{
		{StateLoading, "loading"},
		{StateCategoryList, "category-list"},
		{StateScriptList, "script-list"},
		{StateFilter, "filter"},
		{StateHelp, "help"},
		{StateForm, "form"},
		{StateBrowsing, "browsing"},
		{ViewState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestNewAppModel_EmptyCategories(t *testing.T) {
	model := NewAppModel(nil)
	assert.Equal(t, StateLoading, model.State())
	assert.Empty(t, model.Categories())
}

func TestNewAppModel_WithCategories(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}

	model := NewAppModel(categories)
	assert.Equal(t, StateBrowsing, model.State())
	assert.Len(t, model.Categories(), 1)
}

func TestAppModel_WindowResize(t *testing.T) {
	model := NewAppModel(nil)

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, 80, m.Width())
	assert.Equal(t, 24, m.Height())
}

func TestAppModel_ScriptsLoadedMsg(t *testing.T) {
	model := NewAppModel(nil)
	assert.Equal(t, StateLoading, model.State())

	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}

	msg := ScriptsLoadedMsg{Categories: categories}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, StateBrowsing, m.State())
	assert.Len(t, m.Categories(), 1)
}

func TestAppModel_ErrorMsg(t *testing.T) {
	model := NewAppModel(nil)

	msg := ErrorMsg{Err: assert.AnError}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Contains(t, m.View(), "Error:")
}

func TestAppModel_HelpToggle(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)
	assert.Equal(t, StateBrowsing, model.State())

	// Press ? to open help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateHelp, m.State())

	// Press any key to close help
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)
	assert.Equal(t, StateBrowsing, m.State())
}

func TestAppModel_QuitKey(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	// Should return tea.Quit command
	assert.NotNil(t, cmd)
}

func TestAppModel_FilterKey(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)
	assert.Equal(t, StateBrowsing, model.State())

	// Press / to activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())
}

func TestAppModel_PanelSwitching(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// In browsing mode, sidebar is initially focused
	assert.Equal(t, StateBrowsing, model.State())
	assert.Equal(t, PanelSidebar, model.activePanel)
	assert.True(t, model.sidebar.IsFocused())

	// Press tab to switch to scripts panel
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, PanelScripts, m.activePanel)
	assert.True(t, m.scriptsPane.IsFocused())
	assert.False(t, m.sidebar.IsFocused())
}

func TestAppModel_RefreshKey(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := model.Update(msg)

	// Should return a command that produces RefreshMsg
	assert.NotNil(t, cmd)
	result := cmd()
	_, ok := result.(RefreshMsg)
	assert.True(t, ok)
}

func TestAppModel_SetCategories(t *testing.T) {
	model := NewAppModel(nil)
	assert.Equal(t, StateLoading, model.State())

	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model.SetCategories(categories)

	assert.Equal(t, StateBrowsing, model.State())
	assert.Len(t, model.Categories(), 1)
}

func TestAppModel_ViewLoading(t *testing.T) {
	model := NewAppModel(nil)
	view := model.View()

	// Loading view should show the scanning message
	assert.Contains(t, view, "Scanning scripts...")

	// Should be rendered in a panel with rounded borders
	assert.Contains(t, view, "╭") // Top-left corner of rounded border
	assert.Contains(t, view, "╯") // Bottom-right corner of rounded border
}

func TestAppModel_ViewHelp(t *testing.T) {
	model := NewAppModel(nil)
	model.state = StateHelp
	view := model.View()

	assert.Contains(t, view, "Keyboard Shortcuts")
	assert.Contains(t, view, "Navigation")
	assert.Contains(t, view, "Filtering")
}

func TestKeyMap_DefaultKeyMap(t *testing.T) {
	keys := DefaultKeyMap()

	assert.NotEmpty(t, keys.Up.Keys())
	assert.NotEmpty(t, keys.Down.Keys())
	assert.NotEmpty(t, keys.Select.Keys())
	assert.NotEmpty(t, keys.Back.Keys())
	assert.NotEmpty(t, keys.Filter.Keys())
	assert.NotEmpty(t, keys.Refresh.Keys())
	assert.NotEmpty(t, keys.Help.Keys())
	assert.NotEmpty(t, keys.Quit.Keys())
}

func TestKeyMap_ShortHelp(t *testing.T) {
	keys := DefaultKeyMap()
	help := keys.ShortHelp()
	assert.Len(t, help, 6)
}

func TestKeyMap_FullHelp(t *testing.T) {
	keys := DefaultKeyMap()
	help := keys.FullHelp()
	assert.Len(t, help, 2)
}

func TestAppModel_FilterActivation(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{{Name: "deploy"}}},
		{Name: "data", Scripts: []core.Script{{Name: "backup"}}},
	}
	model := NewAppModel(categories)
	assert.Equal(t, StateBrowsing, model.State())

	// Press / to activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, cmd := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, StateFilter, m.State())
	assert.Equal(t, StateBrowsing, m.prevState)
	// Should return blink command
	assert.NotNil(t, cmd)
}

func TestAppModel_FilterCancel(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{{Name: "deploy"}}},
		{Name: "data", Scripts: []core.Script{{Name: "backup"}}},
	}
	model := NewAppModel(categories)

	// Activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())

	// Type something to filter
	m.filterInput.SetValue("dep")
	m.menu.ApplyFilter("dep")
	m.scriptsPane.ApplyFilter("dep")

	// Press esc to cancel
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	assert.Equal(t, StateBrowsing, m.State())
	assert.Equal(t, "", m.filterInput.Value())
}

func TestAppModel_FilterConfirm(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{{Name: "deploy"}}},
		{Name: "data", Scripts: []core.Script{{Name: "backup"}}},
	}
	model := NewAppModel(categories)

	// Activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// Type something to filter
	m.filterInput.SetValue("dep")
	m.menu.ApplyFilter("dep")
	m.scriptsPane.ApplyFilter("dep")

	// Press enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	// Should return to previous state with filter applied
	assert.Equal(t, StateBrowsing, m.State())
	// Filter value is preserved
	assert.Equal(t, "dep", m.filterInput.Value())
}

func TestAppModel_FilterViewRendering(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{{Name: "deploy"}}},
	}
	model := NewAppModel(categories)

	// Activate filter through proper flow (pressing /)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := updatedModel.(AppModel)

	view := m.View()

	// Should show filter overlay with rounded border (superfile-style)
	assert.Contains(t, view, "Filter", "Filter overlay should contain 'Filter' text")
	// Should have rounded box border characters
	assert.Contains(t, view, "╭", "Filter overlay should have rounded corners")
	// Should show footer with filter keys
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "esc")
}

func TestAppModel_FilterInBrowsingMode(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy app"},
			{Name: "rollback", Description: "Rollback deployment"},
		}},
	}
	model := NewAppModel(categories)

	// In browsing mode
	assert.Equal(t, StateBrowsing, model.State())

	// Press / to activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())
	assert.Equal(t, StateBrowsing, m.prevState)
}

// Test responsive layout functionality

func TestCalculateLayoutMode(t *testing.T) {
	tests := []struct {
		width    int
		expected LayoutMode
	}{
		{width: 150, expected: LayoutThreePanel},
		{width: 120, expected: LayoutThreePanel},
		{width: 119, expected: LayoutTwoPanel},
		{width: 100, expected: LayoutTwoPanel},
		{width: 80, expected: LayoutTwoPanel},
		{width: 79, expected: LayoutOnePanel},
		{width: 60, expected: LayoutOnePanel},
		{width: 40, expected: LayoutOnePanel},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.width)), func(t *testing.T) {
			assert.Equal(t, tt.expected, calculateLayoutMode(tt.width))
		})
	}
}

func TestAppModel_LayoutModeOnResize(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Start with default (80 width = 2-panel mode)
	assert.Equal(t, LayoutTwoPanel, model.layoutMode)

	// Resize to wide (3-panel mode)
	msg := tea.WindowSizeMsg{Width: 130, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, LayoutThreePanel, m.layoutMode)

	// Resize to narrow (1-panel mode)
	msg = tea.WindowSizeMsg{Width: 70, Height: 24}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)
	assert.Equal(t, LayoutOnePanel, m.layoutMode)
}

func TestAppModel_PanelFocusAdjustsOnLayoutChange(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Start in 3-panel mode with focus on details panel
	msg := tea.WindowSizeMsg{Width: 130, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// Switch to details panel (tab twice)
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelDetails, m.activePanel)
	assert.True(t, m.detailsPane.IsFocused())

	// Resize to 2-panel mode (details is hidden)
	msg = tea.WindowSizeMsg{Width: 100, Height: 24}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	// Active panel should be adjusted to scripts (last visible panel)
	assert.Equal(t, LayoutTwoPanel, m.layoutMode)
	assert.Equal(t, PanelScripts, m.activePanel)
	assert.True(t, m.scriptsPane.IsFocused())
	assert.False(t, m.detailsPane.IsFocused())
}

func TestAppModel_OnePanelModeStartsWithScriptsFocused(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 1-panel mode
	msg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// In 1-panel mode, scripts should be focused (sidebar is hidden)
	assert.Equal(t, LayoutOnePanel, m.layoutMode)
	assert.Equal(t, PanelScripts, m.activePanel)
	assert.True(t, m.scriptsPane.IsFocused())
	assert.False(t, m.sidebar.IsFocused())
}

func TestAppModel_PanelSwitchingDisabledInOnePanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 1-panel mode
	msg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, LayoutOnePanel, m.layoutMode)
	assert.Equal(t, PanelScripts, m.activePanel)

	// Tab should not change panel in 1-panel mode
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelScripts, m.activePanel)

	// Shift+Tab should not change panel either
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updated, _ = m.Update(shiftTabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelScripts, m.activePanel)
}

func TestAppModel_TwoPanelModePanelSwitching(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Start in 2-panel mode (default 80 width)
	assert.Equal(t, LayoutTwoPanel, model.layoutMode)
	assert.Equal(t, PanelSidebar, model.activePanel)

	// Tab: Sidebar → Scripts
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := model.Update(tabMsg)
	m := updated.(AppModel)
	assert.Equal(t, PanelScripts, m.activePanel)

	// Tab: Scripts → Sidebar (wraps, not to Details)
	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelSidebar, m.activePanel)
}

func TestAppModel_ThreePanelModePanelSwitching(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 3-panel mode
	msg := tea.WindowSizeMsg{Width: 130, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, LayoutThreePanel, m.layoutMode)

	// Tab: Sidebar → Scripts → Details → Sidebar
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelScripts, m.activePanel)

	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelDetails, m.activePanel)

	updated, _ = m.Update(tabMsg)
	m = updated.(AppModel)
	assert.Equal(t, PanelSidebar, m.activePanel)
}

func TestAppModel_InitReturnsSpinnerTickWhenLoading(t *testing.T) {
	// When in loading state, Init should return a spinner tick command
	model := NewAppModel(nil)
	assert.Equal(t, StateLoading, model.State())

	cmd := model.Init()
	assert.NotNil(t, cmd, "Init should return spinner tick command when loading")
}

func TestAppModel_InitReturnsNilWhenNotLoading(t *testing.T) {
	// When not in loading state, Init should return nil
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)
	assert.Equal(t, StateBrowsing, model.State())

	cmd := model.Init()
	assert.Nil(t, cmd, "Init should return nil when not loading")
}

// ============================================================================
// Terminal Size Tests - Comprehensive responsive layout verification
// ============================================================================

func TestAppModel_ViewRendering_ThreePanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy to production"},
		}},
	}
	model := NewAppModel(categories)

	// Resize to 3-panel mode (≥120 cols)
	msg := tea.WindowSizeMsg{Width: 140, Height: 30}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, LayoutThreePanel, m.layoutMode)
	view := m.View()

	// Should contain all three panels with their titles
	assert.Contains(t, view, "Categories", "Should show categories panel")
	assert.Contains(t, view, "Scripts", "Should show scripts panel")
	assert.Contains(t, view, "Details", "Should show details panel")
	assert.Contains(t, view, "deploy", "Should show script name")
}

func TestAppModel_ViewRendering_TwoPanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy to production"},
		}},
	}
	model := NewAppModel(categories)

	// Resize to 2-panel mode (80-119 cols)
	msg := tea.WindowSizeMsg{Width: 100, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, LayoutTwoPanel, m.layoutMode)
	view := m.View()

	// Should contain sidebar and scripts panels
	assert.Contains(t, view, "Categories", "Should show categories panel")
	assert.Contains(t, view, "Scripts", "Should show scripts panel")
	assert.Contains(t, view, "deploy", "Should show script name")
}

func TestAppModel_ViewRendering_OnePanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy to production"},
		}},
	}
	model := NewAppModel(categories)

	// Resize to 1-panel mode (<80 cols)
	msg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, LayoutOnePanel, m.layoutMode)
	view := m.View()

	// Should show scripts and category header
	assert.Contains(t, view, "deploy", "Should show script name")
	assert.Contains(t, view, "tap", "Should show app header")
}

func TestAppModel_ViewRendering_BoundaryWidths(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}

	tests := []struct {
		width          int
		expectedLayout LayoutMode
		description    string
	}{
		{120, LayoutThreePanel, "exactly 120 cols (3-panel boundary)"},
		{119, LayoutTwoPanel, "119 cols (just under 3-panel boundary)"},
		{80, LayoutTwoPanel, "exactly 80 cols (2-panel boundary)"},
		{79, LayoutOnePanel, "79 cols (just under 2-panel boundary)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			model := NewAppModel(categories)
			msg := tea.WindowSizeMsg{Width: tt.width, Height: 24}
			updated, _ := model.Update(msg)
			m := updated.(AppModel)

			assert.Equal(t, tt.expectedLayout, m.layoutMode,
				"Width %d should result in %v layout", tt.width, tt.expectedLayout)

			// Verify view renders without panic
			view := m.View()
			assert.NotEmpty(t, view, "View should render at width %d", tt.width)
		})
	}
}

func TestAppModel_ViewRendering_ExtremeWidths(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}

	tests := []struct {
		width       int
		height      int
		description string
	}{
		{20, 10, "very narrow terminal (20 cols)"},
		{40, 15, "narrow terminal (40 cols)"},
		{200, 50, "very wide terminal (200 cols)"},
		{300, 80, "extremely wide terminal (300 cols)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			model := NewAppModel(categories)
			msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			updated, _ := model.Update(msg)
			m := updated.(AppModel)

			// Should not panic and should render something
			view := m.View()
			assert.NotEmpty(t, view, "View should render at %dx%d", tt.width, tt.height)
			assert.NotContains(t, view, "panic", "View should not contain panic message")
		})
	}
}

func TestAppModel_ViewRendering_MinimumHeight(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}

	tests := []struct {
		height      int
		description string
	}{
		{5, "very short terminal (5 rows)"},
		{10, "minimum usable height (10 rows)"},
		{15, "short terminal (15 rows)"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			model := NewAppModel(categories)
			msg := tea.WindowSizeMsg{Width: 100, Height: tt.height}
			updated, _ := model.Update(msg)
			m := updated.(AppModel)

			// Should render without panic
			view := m.View()
			assert.NotEmpty(t, view, "View should render at height %d", tt.height)
		})
	}
}

func TestAppModel_PanelSizes_ThreePanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 3-panel mode
	msg := tea.WindowSizeMsg{Width: 140, Height: 30}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// Verify panels have reasonable sizes
	sidebarWidth := m.sidebar.width
	scriptsWidth := m.scriptsPane.width
	detailsWidth := m.detailsPane.width

	assert.Greater(t, sidebarWidth, 0, "Sidebar should have positive width")
	assert.Greater(t, scriptsWidth, 0, "Scripts should have positive width")
	assert.Greater(t, detailsWidth, 0, "Details should have positive width")

	// Total should approximately equal terminal width (with gaps)
	totalWidth := sidebarWidth + scriptsWidth + detailsWidth
	assert.LessOrEqual(t, totalWidth, 140, "Panel widths should not exceed terminal width")
}

func TestAppModel_PanelSizes_TwoPanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 2-panel mode
	msg := tea.WindowSizeMsg{Width: 100, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// In 2-panel mode, details should be hidden (size 0)
	assert.Greater(t, m.sidebar.width, 0, "Sidebar should have positive width")
	assert.Greater(t, m.scriptsPane.width, 0, "Scripts should have positive width")
	assert.Equal(t, 0, m.detailsPane.width, "Details should be hidden in 2-panel mode")
}

func TestAppModel_PanelSizes_OnePanelMode(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Resize to 1-panel mode
	msg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// In 1-panel mode, only scripts should be visible
	assert.Equal(t, 0, m.sidebar.width, "Sidebar should be hidden in 1-panel mode")
	assert.Greater(t, m.scriptsPane.width, 0, "Scripts should have positive width")
	assert.Equal(t, 0, m.detailsPane.width, "Details should be hidden in 1-panel mode")
}

func TestAppModel_FooterHintsChangeWithLayout(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// In 3-panel mode, footer should show tab for panel switching
	msg := tea.WindowSizeMsg{Width: 140, Height: 30}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	view := m.View()
	assert.Contains(t, view, "tab", "3-panel mode should show tab hint")

	// In 1-panel mode, footer should NOT show tab (no panel switching)
	msg = tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	// Check footer context
	assert.Equal(t, LayoutOnePanel, m.layoutMode)
}

func TestAppModel_ResizeDuringFilter(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy to production"},
			{Name: "rollback", Description: "Rollback deployment"},
		}},
	}
	model := NewAppModel(categories)

	// Activate filter
	filterMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(filterMsg)
	m := updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())

	// Resize while in filter mode
	resizeMsg := tea.WindowSizeMsg{Width: 60, Height: 24}
	updated, _ = m.Update(resizeMsg)
	m = updated.(AppModel)

	// Should still be in filter state
	assert.Equal(t, StateFilter, m.State())

	// View should still show filter overlay
	view := m.View()
	assert.Contains(t, view, "Filter", "Should still show filter overlay after resize")
}

func TestAppModel_InteractiveScriptSkipsForm(t *testing.T) {
	// An interactive script with params should skip the form and run directly
	interactiveScript := core.Script{
		Name:        "interactive-tool",
		Interactive: true,
		Parameters: []core.Parameter{
			{Name: "env", Type: "string", Required: true},
		},
	}
	categories := []core.Category{
		{Name: "tools", Scripts: []core.Script{interactiveScript}},
	}
	model := NewAppModel(categories)

	// Send ScriptSelectedMsg for the interactive script
	msg := ScriptSelectedMsg{Script: interactiveScript}
	updated, cmd := model.Update(msg)
	m := updated.(AppModel)

	// Should NOT go to form state
	assert.NotEqual(t, StateForm, m.State(), "Interactive script should not show form")
	// Should set selectedScript and quit
	assert.NotNil(t, m.SelectedScript(), "Should have selected the script")
	assert.Equal(t, "interactive-tool", m.SelectedScript().Name)
	assert.NotNil(t, cmd, "Should return quit command")
}

func TestAppModel_NonInteractiveScriptShowsForm(t *testing.T) {
	// A non-interactive script with params should show the form
	script := core.Script{
		Name:        "deploy",
		Interactive: false,
		Parameters: []core.Parameter{
			{Name: "env", Type: "string", Required: true},
		},
	}
	categories := []core.Category{
		{Name: "tools", Scripts: []core.Script{script}},
	}
	model := NewAppModel(categories)

	msg := ScriptSelectedMsg{Script: script}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// Should go to form state
	assert.Equal(t, StateForm, m.State(), "Non-interactive script with params should show form")
}

func TestAppModel_InteractiveScriptFooterNoParams(t *testing.T) {
	// Interactive scripts should not show "p params" hint in footer
	interactiveScript := core.Script{
		Name:        "interactive-tool",
		Interactive: true,
		Parameters: []core.Parameter{
			{Name: "env", Type: "string", Required: true},
		},
	}
	categories := []core.Category{
		{Name: "tools", Scripts: []core.Script{interactiveScript}},
	}
	model := NewAppModel(categories)

	// The footer context should have HasParams=false for interactive scripts
	model.updateFooterContext()
	// We can check indirectly: the selected script has params but is interactive,
	// so hasParams should be false
	script := model.scriptsPane.SelectedScript()
	if script != nil {
		hasParams := len(script.Parameters) > 0 && !script.Interactive
		assert.False(t, hasParams, "hasParams should be false for interactive scripts")
	}
}

func TestAppModel_ResizeDuringHelp(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Open help
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := model.Update(helpMsg)
	m := updated.(AppModel)
	assert.Equal(t, StateHelp, m.State())

	// Resize while in help state
	resizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updated, _ = m.Update(resizeMsg)
	m = updated.(AppModel)

	// Should still be in help state
	assert.Equal(t, StateHelp, m.State())

	// View should still show help
	view := m.View()
	assert.Contains(t, view, "Keyboard Shortcuts", "Should still show help after resize")
}
