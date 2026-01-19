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
