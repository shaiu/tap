package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCategories() []core.Category {
	return []core.Category{
		{
			Name: "deployment",
			Scripts: []core.Script{
				{Name: "deploy", Description: "Deploy application"},
				{Name: "rollback", Description: "Rollback deployment"},
			},
		},
		{
			Name: "data",
			Scripts: []core.Script{
				{Name: "backup", Description: "Backup database"},
			},
		},
	}
}

func TestNewMenuModel(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	assert.False(t, menu.ShowingScripts())
	assert.Equal(t, -1, menu.SelectedCategoryIndex())
	assert.Nil(t, menu.SelectedScript())
}

func TestMenuModel_EmptyCategories(t *testing.T) {
	menu := NewMenuModel(nil, 80, 24)

	assert.False(t, menu.ShowingScripts())
	view := menu.View()
	assert.Contains(t, view, "tap")
}

func TestMenuModel_CategoryNavigation_Down(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Press down to move to second category
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := menu.Update(msg)

	// Should still be in category view
	assert.False(t, updated.ShowingScripts())
}

func TestMenuModel_CategoryNavigation_UpDown(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Press j (vim-style down)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	menu, _ = menu.Update(msg)
	assert.False(t, menu.ShowingScripts())

	// Press k (vim-style up)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	menu, _ = menu.Update(msg)
	assert.False(t, menu.ShowingScripts())
}

func TestMenuModel_DrillIntoCategory(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Press enter to drill into first category
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := menu.Update(msg)

	assert.True(t, updated.ShowingScripts())
	assert.Equal(t, 0, updated.SelectedCategoryIndex())
}

func TestMenuModel_BackFromScripts(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Drill into category
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	menu, _ = menu.Update(msg)
	require.True(t, menu.ShowingScripts())

	// Press esc to go back
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	menu, _ = menu.Update(msg)

	assert.False(t, menu.ShowingScripts())
	assert.Equal(t, -1, menu.SelectedCategoryIndex())
}

func TestMenuModel_BackFromScripts_Backspace(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Drill into category
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	menu, _ = menu.Update(msg)
	require.True(t, menu.ShowingScripts())

	// Press backspace to go back
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	menu, _ = menu.Update(msg)

	assert.False(t, menu.ShowingScripts())
}

func TestMenuModel_SelectScript(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Drill into category
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	menu, _ = menu.Update(msg)
	require.True(t, menu.ShowingScripts())

	// Press enter to select script
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	menu, cmd := menu.Update(msg)

	// Should have a selected script
	assert.NotNil(t, menu.SelectedScript())
	assert.Equal(t, "deploy", menu.SelectedScript().Name)

	// Should return a command that produces ScriptSelectedMsg
	assert.NotNil(t, cmd)
	result := cmd()
	selectedMsg, ok := result.(ScriptSelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, "deploy", selectedMsg.Script.Name)
}

func TestMenuModel_WindowResize(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := menu.Update(msg)

	// Should handle resize without error
	view := updated.View()
	assert.NotEmpty(t, view)
}

func TestMenuModel_SetCategories(t *testing.T) {
	menu := NewMenuModel(nil, 80, 24)

	categories := testCategories()
	menu.SetCategories(categories)

	// Should reset navigation state
	assert.False(t, menu.ShowingScripts())
	assert.Equal(t, -1, menu.SelectedCategoryIndex())
}

func TestMenuModel_SetSize(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	menu.SetSize(100, 30)

	// Should handle resize without error
	view := menu.View()
	assert.NotEmpty(t, view)
}

func TestMenuModel_View_CategoryList(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	view := menu.View()

	// Should show header
	assert.Contains(t, view, "tap")

	// Should show categories
	assert.Contains(t, view, "deployment")
	assert.Contains(t, view, "data")

	// Should show footer with keybindings
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "quit")
}

func TestMenuModel_View_ScriptList(t *testing.T) {
	categories := testCategories()
	menu := NewMenuModel(categories, 80, 24)

	// Drill into category
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	menu, _ = menu.Update(msg)

	view := menu.View()

	// Should show category name in header
	assert.Contains(t, view, "deployment")

	// Should show scripts
	assert.Contains(t, view, "deploy")
	assert.Contains(t, view, "Deploy application")

	// Should show footer with back option
	assert.Contains(t, view, "back")
}

func TestCategoryItem_Interface(t *testing.T) {
	category := core.Category{
		Name: "test",
		Scripts: []core.Script{
			{Name: "script1"},
			{Name: "script2"},
		},
	}
	item := CategoryItem{Category: category}

	assert.Equal(t, "test", item.Title())
	assert.Equal(t, "2 scripts", item.Description())
	assert.Equal(t, "test", item.FilterValue())
}

func TestScriptItem_Interface(t *testing.T) {
	script := core.Script{
		Name:        "deploy",
		Description: "Deploy the app",
		Tags:        []string{"prod", "deploy"},
	}
	item := ScriptItem{Script: script}

	assert.Equal(t, "deploy", item.Title())
	assert.Equal(t, "Deploy the app", item.Description())

	filterValue := item.FilterValue()
	assert.Contains(t, filterValue, "deploy")
	assert.Contains(t, filterValue, "Deploy the app")
	assert.Contains(t, filterValue, "prod")
}

func TestCategoryDelegate_Height(t *testing.T) {
	delegate := CategoryDelegate{}
	assert.Equal(t, 1, delegate.Height())
	assert.Equal(t, 0, delegate.Spacing())
}

func TestScriptDelegate_Height(t *testing.T) {
	delegate := ScriptDelegate{}
	assert.Equal(t, 2, delegate.Height())
	assert.Equal(t, 1, delegate.Spacing())
}
