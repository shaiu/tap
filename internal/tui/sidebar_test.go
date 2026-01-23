package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

func TestNewSidebarModel(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}, {Name: "rollback.sh"}}},
		{Name: "Database", Scripts: []core.Script{{Name: "backup.sh"}}},
	}

	m := NewSidebarModel(categories)

	// Should have 3 items: All Scripts + 2 categories
	if len(m.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(m.items))
	}

	// First item should be "All Scripts" with total count
	if m.items[0].Type != SidebarItemAllScripts {
		t.Errorf("expected first item to be AllScripts, got %v", m.items[0].Type)
	}
	if m.items[0].Count != 3 {
		t.Errorf("expected All Scripts count to be 3, got %d", m.items[0].Count)
	}

	// Remaining items should be categories
	if m.items[1].Type != SidebarItemCategory {
		t.Errorf("expected second item to be Category, got %v", m.items[1].Type)
	}
	if m.items[1].Name != "Deploy" {
		t.Errorf("expected second item name to be Deploy, got %s", m.items[1].Name)
	}
}

func TestNewSidebarModelWithPinned(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
	}
	pinnedScripts := []core.Script{
		{Name: "quick-deploy", Path: "/scripts/quick-deploy.sh"},
		{Name: "backup-prod", Path: "/scripts/backup-prod.sh"},
	}

	m := NewSidebarModelWithPinned(categories, pinnedScripts)

	// Should have 5 items: All Scripts + 1 category + Pinned header + 2 pinned scripts
	if len(m.items) != 5 {
		t.Errorf("expected 5 items, got %d", len(m.items))
	}

	// Check pinned header position
	if m.items[2].Type != SidebarItemPinnedHeader {
		t.Errorf("expected item at index 2 to be PinnedHeader, got %v", m.items[2].Type)
	}
	if m.items[2].Name != "Pinned" {
		t.Errorf("expected PinnedHeader name to be 'Pinned', got %s", m.items[2].Name)
	}

	// Check pinned scripts
	if m.items[3].Type != SidebarItemPinnedScript {
		t.Errorf("expected item at index 3 to be PinnedScript, got %v", m.items[3].Type)
	}
	if m.items[3].Name != "quick-deploy" {
		t.Errorf("expected first pinned script name to be 'quick-deploy', got %s", m.items[3].Name)
	}
	if m.items[3].Script == nil {
		t.Error("expected first pinned script to have Script reference")
	}

	// Check stored pinned scripts
	if len(m.PinnedScripts()) != 2 {
		t.Errorf("expected 2 pinned scripts stored, got %d", len(m.PinnedScripts()))
	}
}

func TestSidebarModel_Navigation(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
		{Name: "Database", Scripts: []core.Script{{Name: "backup.sh"}}},
	}

	m := NewSidebarModel(categories)
	m.SetSize(30, 20)

	// Initial cursor at 0
	if m.Cursor() != 0 {
		t.Errorf("expected initial cursor at 0, got %d", m.Cursor())
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.Cursor())
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.Cursor())
	}

	// Can't go above 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
	}
}

func TestSidebarModel_NavigationSkipsPinnedHeader(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
	}
	pinnedScripts := []core.Script{
		{Name: "quick-deploy"},
	}

	m := NewSidebarModelWithPinned(categories, pinnedScripts)
	m.SetSize(30, 20)

	// Items: [0] All Scripts, [1] Deploy, [2] Pinned header, [3] quick-deploy
	// Navigate to Deploy (index 1)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", m.Cursor())
	}

	// Navigate down - should skip header (index 2) and land on pinned script (index 3)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 3 {
		t.Errorf("expected cursor at 3 (skipping header), got %d", m.Cursor())
	}

	// Navigate back up - should skip header (index 2) and land on Deploy (index 1)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1 (skipping header), got %d", m.Cursor())
	}
}

func TestSidebarModel_Selection(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
	}
	pinnedScripts := []core.Script{
		{Name: "quick-deploy", Path: "/scripts/quick-deploy.sh"},
	}

	m := NewSidebarModelWithPinned(categories, pinnedScripts)
	m.SetSize(30, 20)

	// At All Scripts
	if !m.IsAllScriptsSelected() {
		t.Error("expected All Scripts to be selected initially")
	}
	if m.SelectedCategory() != nil {
		t.Error("expected no category to be selected when All Scripts is selected")
	}
	if m.IsPinnedScriptSelected() {
		t.Error("expected pinned script not to be selected")
	}

	// Move to Deploy category
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.IsAllScriptsSelected() {
		t.Error("expected All Scripts not to be selected")
	}
	if m.SelectedCategory() == nil {
		t.Error("expected a category to be selected")
	}
	if m.SelectedCategory().Name != "Deploy" {
		t.Errorf("expected selected category to be Deploy, got %s", m.SelectedCategory().Name)
	}

	// Move to pinned script (skips header)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !m.IsPinnedScriptSelected() {
		t.Error("expected pinned script to be selected")
	}
	pinnedScript := m.SelectedPinnedScript()
	if pinnedScript == nil {
		t.Error("expected SelectedPinnedScript to return a script")
	}
	if pinnedScript.Name != "quick-deploy" {
		t.Errorf("expected selected pinned script to be quick-deploy, got %s", pinnedScript.Name)
	}
}

func TestSidebarModel_SetPinnedScripts(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
	}

	m := NewSidebarModel(categories)

	// Initially no pinned scripts
	if len(m.PinnedScripts()) != 0 {
		t.Error("expected no pinned scripts initially")
	}

	// Set pinned scripts
	pinnedScripts := []core.Script{
		{Name: "my-script"},
	}
	m.SetPinnedScripts(pinnedScripts)

	// Should now have pinned scripts
	if len(m.PinnedScripts()) != 1 {
		t.Errorf("expected 1 pinned script, got %d", len(m.PinnedScripts()))
	}

	// Items should include pinned section
	// [0] All Scripts, [1] Deploy, [2] Pinned header, [3] my-script
	if len(m.items) != 4 {
		t.Errorf("expected 4 items after setting pinned scripts, got %d", len(m.items))
	}
}

func TestSidebarModel_ViewRendersPinned(t *testing.T) {
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}}},
	}
	pinnedScripts := []core.Script{
		{Name: "quick-deploy"},
	}

	m := NewSidebarModelWithPinned(categories, pinnedScripts)
	m.SetSize(40, 20)

	view := m.View()

	// Should contain separator
	if !strings.Contains(view, "─") {
		t.Error("expected view to contain separator line")
	}

	// Should contain "Pinned"
	if !strings.Contains(view, "Pinned") {
		t.Error("expected view to contain 'Pinned' header")
	}

	// Should contain pinned script name
	if !strings.Contains(view, "quick-deploy") {
		t.Error("expected view to contain pinned script name")
	}
}

func TestSidebarModel_Focus(t *testing.T) {
	m := NewSidebarModel(nil)

	// Initially focused
	if !m.IsFocused() {
		t.Error("expected sidebar to be focused initially")
	}

	// Set unfocused
	m.SetFocused(false)
	if m.IsFocused() {
		t.Error("expected sidebar to be unfocused")
	}

	// When unfocused, navigation should not work
	m.SetSize(30, 20)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Cursor should stay at 0 since unfocused
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to stay at 0 when unfocused, got %d", m.Cursor())
	}
}

func TestBuildSidebarItems_NoPinned(t *testing.T) {
	categories := []core.Category{
		{Name: "Cat1", Scripts: []core.Script{{Name: "s1"}, {Name: "s2"}}},
		{Name: "Cat2", Scripts: []core.Script{{Name: "s3"}}},
	}

	items := buildSidebarItems(categories, nil)

	// Should have: All Scripts + Cat1 + Cat2
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Verify All Scripts total count
	if items[0].Count != 3 {
		t.Errorf("expected All Scripts count to be 3, got %d", items[0].Count)
	}
}

func TestBuildSidebarItems_WithPinned(t *testing.T) {
	categories := []core.Category{
		{Name: "Cat1", Scripts: []core.Script{{Name: "s1"}}},
	}
	pinned := []core.Script{
		{Name: "p1"},
		{Name: "p2"},
	}

	items := buildSidebarItems(categories, pinned)

	// Should have: All Scripts + Cat1 + Pinned header + p1 + p2
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}

	// Verify structure
	types := []SidebarItemType{
		SidebarItemAllScripts,
		SidebarItemCategory,
		SidebarItemPinnedHeader,
		SidebarItemPinnedScript,
		SidebarItemPinnedScript,
	}
	for i, expectedType := range types {
		if items[i].Type != expectedType {
			t.Errorf("item %d: expected type %v, got %v", i, expectedType, items[i].Type)
		}
	}

	// Verify pinned header count
	if items[2].Count != 2 {
		t.Errorf("expected pinned header count to be 2, got %d", items[2].Count)
	}
}
