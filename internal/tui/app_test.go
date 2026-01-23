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
	assert.Equal(t, StateCategoryList, model.State())
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

	assert.Equal(t, StateCategoryList, m.State())
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
	assert.Equal(t, StateCategoryList, model.State())

	// Press ? to open help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateHelp, m.State())

	// Press any key to close help
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)
	assert.Equal(t, StateCategoryList, m.State())
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
	assert.Equal(t, StateCategoryList, model.State())

	// Press / to activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())
}

func TestAppModel_BackFromScriptList(t *testing.T) {
	categories := []core.Category{
		{Name: "test", Scripts: []core.Script{{Name: "script1"}}},
	}
	model := NewAppModel(categories)

	// Manually set state to script list
	model.state = StateScriptList
	model.selectedCatIdx = 0

	// Press esc to go back
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateCategoryList, m.State())
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

	assert.Equal(t, StateCategoryList, model.State())
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
	assert.Equal(t, StateCategoryList, model.State())

	// Press / to activate filter
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, cmd := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, StateFilter, m.State())
	assert.Equal(t, StateCategoryList, m.prevState)
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

	// Press esc to cancel
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	assert.Equal(t, StateCategoryList, m.State())
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

	// Press enter to confirm
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)

	// Should return to previous state with filter applied
	assert.Equal(t, StateCategoryList, m.State())
	// Filter value is preserved
	assert.Equal(t, "dep", m.filterInput.Value())
}

func TestAppModel_FilterViewRendering(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{{Name: "deploy"}}},
	}
	model := NewAppModel(categories)
	model.state = StateFilter
	model.prevState = StateCategoryList

	view := model.View()

	// Should show filter input
	assert.Contains(t, view, "Filter:")
	// Should show footer with filter keys
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "esc")
}

func TestAppModel_FilterInScriptList(t *testing.T) {
	categories := []core.Category{
		{Name: "deployment", Scripts: []core.Script{
			{Name: "deploy", Description: "Deploy app"},
			{Name: "rollback", Description: "Rollback deployment"},
		}},
	}
	model := NewAppModel(categories)

	// Drill into category first
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)
	assert.Equal(t, StateScriptList, m.State())

	// Press / to activate filter
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ = m.Update(msg)
	m = updated.(AppModel)
	assert.Equal(t, StateFilter, m.State())
	assert.Equal(t, StateScriptList, m.prevState)
}
