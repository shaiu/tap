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
	assert.Contains(t, view, "Loading")
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
