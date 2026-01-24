package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

func TestNewScriptsModel(t *testing.T) {
	m := NewScriptsModel()

	if len(m.scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(m.scripts))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
	if m.focused {
		t.Error("expected scripts panel to be unfocused initially")
	}
}

func TestScriptsModel_SetScripts(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to prod"},
		{Name: "backup.sh", Description: "Backup database"},
		{Name: "test.sh", Description: "Run tests"},
	}

	m.SetScripts(scripts)

	if len(m.scripts) != 3 {
		t.Errorf("expected 3 scripts, got %d", len(m.scripts))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", m.cursor)
	}
}

func TestScriptsModel_SetScripts_CursorBounds(t *testing.T) {
	m := NewScriptsModel()
	m.cursor = 5 // Set cursor beyond bounds

	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
	}
	m.SetScripts(scripts)

	// Cursor should be clamped to last valid index
	if m.cursor != 1 {
		t.Errorf("expected cursor to be clamped to 1, got %d", m.cursor)
	}
}

func TestScriptsModel_SetAllScripts(t *testing.T) {
	m := NewScriptsModel()
	categories := []core.Category{
		{Name: "Deploy", Scripts: []core.Script{{Name: "deploy.sh"}, {Name: "rollback.sh"}}},
		{Name: "Database", Scripts: []core.Script{{Name: "backup.sh"}}},
	}

	m.SetAllScripts(categories)

	if len(m.scripts) != 3 {
		t.Errorf("expected 3 scripts (all), got %d", len(m.scripts))
	}
}

func TestScriptsModel_Navigation(t *testing.T) {
	m := NewScriptsModel()
	m.SetFocused(true)
	m.SetSize(40, 20)

	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
		{Name: "test.sh"},
	}
	m.SetScripts(scripts)

	// Initial cursor at 0
	if m.Cursor() != 0 {
		t.Errorf("expected initial cursor at 0, got %d", m.Cursor())
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.Cursor())
	}

	// Move down again
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor at 2 after down, got %d", m.Cursor())
	}

	// Can't go past end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", m.Cursor())
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Cursor() != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", m.Cursor())
	}

	// Move up to 0
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

func TestScriptsModel_NavigationWhenUnfocused(t *testing.T) {
	m := NewScriptsModel()
	m.SetFocused(false)
	m.SetSize(40, 20)

	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
	}
	m.SetScripts(scripts)

	// Should not move when unfocused
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to stay at 0 when unfocused, got %d", m.Cursor())
	}
}

func TestScriptsModel_SelectedScript(t *testing.T) {
	m := NewScriptsModel()
	m.SetFocused(true)
	m.SetSize(40, 20)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to prod"},
		{Name: "backup.sh", Description: "Backup database"},
	}
	m.SetScripts(scripts)

	// Get selected script
	selected := m.SelectedScript()
	if selected == nil {
		t.Fatal("expected selected script to not be nil")
	}
	if selected.Name != "deploy.sh" {
		t.Errorf("expected selected script to be deploy.sh, got %s", selected.Name)
	}

	// Move cursor and check again
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected = m.SelectedScript()
	if selected == nil {
		t.Fatal("expected selected script to not be nil")
	}
	if selected.Name != "backup.sh" {
		t.Errorf("expected selected script to be backup.sh, got %s", selected.Name)
	}
}

func TestScriptsModel_SelectedScript_Empty(t *testing.T) {
	m := NewScriptsModel()

	selected := m.SelectedScript()
	if selected != nil {
		t.Error("expected selected script to be nil when list is empty")
	}
}

func TestScriptsModel_ApplyFilter(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(40, 20)

	scripts := []core.Script{
		{Name: "deploy-staging.sh", Description: "Deploy to staging"},
		{Name: "deploy-production.sh", Description: "Deploy to production"},
		{Name: "backup-database.sh", Description: "Backup PostgreSQL"},
		{Name: "run-tests.sh", Description: "Execute test suite"},
	}
	m.SetScripts(scripts)

	// Apply filter
	m.ApplyFilter("deploy")

	// Should match 2 scripts
	if m.ScriptCount() != 2 {
		t.Errorf("expected 2 filtered scripts, got %d", m.ScriptCount())
	}
	if m.TotalCount() != 4 {
		t.Errorf("expected total count of 4, got %d", m.TotalCount())
	}
	if m.FilterQuery() != "deploy" {
		t.Errorf("expected filter query 'deploy', got '%s'", m.FilterQuery())
	}

	// Verify filtered list
	list := m.displayList()
	if len(list) != 2 {
		t.Errorf("expected filtered list length 2, got %d", len(list))
	}
	if list[0].Name != "deploy-staging.sh" {
		t.Errorf("expected first match to be deploy-staging.sh, got %s", list[0].Name)
	}
}

func TestScriptsModel_ApplyFilter_CaseInsensitive(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "Deploy.sh", Description: "Deploy to prod"},
		{Name: "backup.sh", Description: "Backup database"},
	}
	m.SetScripts(scripts)

	// Filter with lowercase should match uppercase
	m.ApplyFilter("DEPLOY")

	if m.ScriptCount() != 1 {
		t.Errorf("expected 1 filtered script, got %d", m.ScriptCount())
	}
}

func TestScriptsModel_ApplyFilter_MatchesDescription(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
		{Name: "backup.sh", Description: "Backup PostgreSQL database"},
	}
	m.SetScripts(scripts)

	// Filter by description content
	m.ApplyFilter("PostgreSQL")

	if m.ScriptCount() != 1 {
		t.Errorf("expected 1 filtered script, got %d", m.ScriptCount())
	}
	if m.displayList()[0].Name != "backup.sh" {
		t.Errorf("expected match to be backup.sh, got %s", m.displayList()[0].Name)
	}
}

func TestScriptsModel_ApplyFilter_MatchesTags(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy.sh", Tags: []string{"production", "critical"}},
		{Name: "backup.sh", Tags: []string{"database", "maintenance"}},
	}
	m.SetScripts(scripts)

	// Filter by tag
	m.ApplyFilter("critical")

	if m.ScriptCount() != 1 {
		t.Errorf("expected 1 filtered script, got %d", m.ScriptCount())
	}
	if m.displayList()[0].Name != "deploy.sh" {
		t.Errorf("expected match to be deploy.sh, got %s", m.displayList()[0].Name)
	}
}

func TestScriptsModel_ApplyFilter_CursorReset(t *testing.T) {
	m := NewScriptsModel()
	m.SetFocused(true)
	m.SetSize(40, 20)

	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
		{Name: "test.sh"},
	}
	m.SetScripts(scripts)

	// Move cursor to last item
	m.cursor = 2

	// Apply filter that matches only 1 item
	m.ApplyFilter("deploy")

	// Cursor should be reset/clamped
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", m.Cursor())
	}
}

func TestScriptsModel_ClearFilter(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
	}
	m.SetScripts(scripts)

	// Apply filter
	m.ApplyFilter("deploy")
	if m.ScriptCount() != 1 {
		t.Errorf("expected 1 filtered script, got %d", m.ScriptCount())
	}

	// Clear filter
	m.ClearFilter()

	if m.FilterQuery() != "" {
		t.Error("expected filter query to be empty after clear")
	}
	if m.ScriptCount() != 2 {
		t.Errorf("expected full list of 2 scripts, got %d", m.ScriptCount())
	}
}

func TestScriptsModel_Focus(t *testing.T) {
	m := NewScriptsModel()

	// Initially unfocused
	if m.IsFocused() {
		t.Error("expected scripts panel to be unfocused initially")
	}

	// Set focused
	m.SetFocused(true)
	if !m.IsFocused() {
		t.Error("expected scripts panel to be focused")
	}

	// Set unfocused
	m.SetFocused(false)
	if m.IsFocused() {
		t.Error("expected scripts panel to be unfocused")
	}
}

func TestScriptsModel_SetSize(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(80, 40)

	if m.width != 80 {
		t.Errorf("expected width 80, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestScriptsModel_View_Empty(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(40, 20)

	view := m.View()

	if !strings.Contains(view, "Scripts") {
		t.Error("expected view to contain 'Scripts' title")
	}
	if !strings.Contains(view, "No scripts found") {
		t.Error("expected view to contain 'No scripts found' message")
	}
}

func TestScriptsModel_View_WithScripts(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(60, 20)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production", Shell: "bash"},
		{Name: "backup.py", Description: "Backup database", Shell: "python"},
	}
	m.SetScripts(scripts)

	view := m.View()

	// Should contain script names
	if !strings.Contains(view, "deploy.sh") {
		t.Error("expected view to contain 'deploy.sh'")
	}
	if !strings.Contains(view, "backup.py") {
		t.Error("expected view to contain 'backup.py'")
	}

	// Should contain descriptions
	if !strings.Contains(view, "Deploy to production") {
		t.Error("expected view to contain description")
	}
}

func TestScriptsModel_View_WithFilter(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(60, 20)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
		{Name: "backup.sh", Description: "Backup database"},
		{Name: "test.sh", Description: "Run tests"},
	}
	m.SetScripts(scripts)

	// Apply filter
	m.ApplyFilter("deploy")

	view := m.View()

	// Should contain filter bar
	if !strings.Contains(view, "Filter:") {
		t.Error("expected view to contain 'Filter:' when filter is active")
	}

	// Should contain filter query
	if !strings.Contains(view, "deploy") {
		t.Error("expected view to contain filter query")
	}

	// Should contain match count [1/3]
	if !strings.Contains(view, "[1/3]") {
		t.Error("expected view to contain match count [1/3]")
	}
}

func TestScriptsModel_View_FilterNoMatches(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(60, 20)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
	}
	m.SetScripts(scripts)

	// Apply filter with no matches
	m.ApplyFilter("nonexistent")

	view := m.View()

	// Should show "No matching scripts" message
	if !strings.Contains(view, "No matching scripts") {
		t.Error("expected view to contain 'No matching scripts' when filter has no results")
	}

	// Should still show filter bar with [0/1]
	if !strings.Contains(view, "[0/1]") {
		t.Error("expected view to contain match count [0/1]")
	}
}

func TestScriptsModel_View_SelectionIndicator(t *testing.T) {
	m := NewScriptsModel()
	m.SetFocused(true)
	m.SetSize(60, 20)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
		{Name: "backup.sh", Description: "Backup database"},
	}
	m.SetScripts(scripts)

	view := m.View()

	// Should contain selection indicator for first item
	if !strings.Contains(view, "●") {
		t.Error("expected view to contain selection indicator (●)")
	}
}

func TestScriptsModel_PanelTitle(t *testing.T) {
	m := NewScriptsModel()

	if m.PanelTitle() != "Scripts" {
		t.Errorf("expected panel title 'Scripts', got '%s'", m.PanelTitle())
	}
}

func TestScriptsModel_RenderItem_NoDescription(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(60, 20)

	script := core.Script{
		Name:        "nodesc.sh",
		Description: "",
		Shell:       "bash",
	}

	rendered := m.renderItem(script, false)

	// Should show "No description" placeholder
	if !strings.Contains(rendered, "No description") {
		t.Error("expected 'No description' placeholder for script without description")
	}
}

func TestScriptsModel_RenderItem_TruncatesLongDescription(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(40, 20) // Narrow width

	longDesc := "This is a very long description that should definitely be truncated when rendered in a narrow panel width"
	script := core.Script{
		Name:        "test.sh",
		Description: longDesc,
		Shell:       "bash",
	}

	rendered := m.renderItem(script, false)

	// Should contain truncation indicator
	if !strings.Contains(rendered, "...") {
		t.Error("expected long description to be truncated with '...'")
	}

	// Should not contain the full description
	if strings.Contains(rendered, "narrow panel width") {
		t.Error("expected long description to be truncated")
	}
}

// Filter Mode Tests

func TestScriptsModel_FilterMode_Toggle(t *testing.T) {
	m := NewScriptsModel()

	// Initially not in filter mode
	if m.IsFilterMode() {
		t.Error("expected filter mode to be false initially")
	}

	// Enable filter mode
	m.SetFilterMode(true)
	if !m.IsFilterMode() {
		t.Error("expected filter mode to be true after enabling")
	}

	// Disable filter mode
	m.SetFilterMode(false)
	if m.IsFilterMode() {
		t.Error("expected filter mode to be false after disabling")
	}
}

func TestScriptsModel_FilterMode_ClearedOnClearFilter(t *testing.T) {
	m := NewScriptsModel()
	m.SetFilterMode(true)
	m.ApplyFilter("test")

	m.ClearFilter()

	if m.IsFilterMode() {
		t.Error("expected filter mode to be cleared after ClearFilter")
	}
	if m.FilterQuery() != "" {
		t.Error("expected filter query to be empty after ClearFilter")
	}
}

func TestScriptsModel_IsMatched_WithNoFilter(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy.sh"},
		{Name: "backup.sh"},
	}
	m.SetScripts(scripts)

	// With no filter, all items should be "matched"
	if !m.IsMatched(0) {
		t.Error("expected IsMatched(0) to be true when no filter")
	}
	if !m.IsMatched(1) {
		t.Error("expected IsMatched(1) to be true when no filter")
	}
}

func TestScriptsModel_IsMatched_WithFilter(t *testing.T) {
	m := NewScriptsModel()
	scripts := []core.Script{
		{Name: "deploy-staging.sh", Description: "Deploy to staging"},
		{Name: "backup-database.sh", Description: "Backup PostgreSQL"},
		{Name: "deploy-production.sh", Description: "Deploy to production"},
	}
	m.SetScripts(scripts)

	m.ApplyFilter("deploy")

	// Index 0 and 2 should match (deploy-staging and deploy-production)
	if !m.IsMatched(0) {
		t.Error("expected IsMatched(0) to be true for deploy-staging.sh")
	}
	if m.IsMatched(1) {
		t.Error("expected IsMatched(1) to be false for backup-database.sh")
	}
	if !m.IsMatched(2) {
		t.Error("expected IsMatched(2) to be true for deploy-production.sh")
	}
}

func TestScriptsModel_FilterMode_DisplayListShowsAllScripts(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(60, 20)

	scripts := []core.Script{
		{Name: "deploy-staging.sh", Description: "Deploy to staging"},
		{Name: "backup-database.sh", Description: "Backup PostgreSQL"},
		{Name: "deploy-production.sh", Description: "Deploy to production"},
	}
	m.SetScripts(scripts)

	// Enable filter mode before applying filter
	m.SetFilterMode(true)
	m.ApplyFilter("deploy")

	// displayListForView should return all scripts in filter mode
	list := m.displayListForView()
	if len(list) != 3 {
		t.Errorf("expected displayListForView to return all 3 scripts in filter mode, got %d", len(list))
	}

	// displayList (filtered) should still show only matches
	filteredList := m.displayList()
	if len(filteredList) != 2 {
		t.Errorf("expected displayList to return 2 matching scripts, got %d", len(filteredList))
	}
}

func TestScriptsModel_FilterMode_ViewShowsAllScripts(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(80, 25)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
		{Name: "backup.sh", Description: "Backup database"},
		{Name: "test.sh", Description: "Run tests"},
	}
	m.SetScripts(scripts)

	// Enable filter mode and apply filter
	m.SetFilterMode(true)
	m.ApplyFilter("deploy")

	view := m.View()

	// All scripts should be visible in filter mode (non-matches dimmed)
	if !strings.Contains(view, "deploy.sh") {
		t.Error("expected view to contain 'deploy.sh' in filter mode")
	}
	if !strings.Contains(view, "backup.sh") {
		t.Error("expected view to contain 'backup.sh' (dimmed) in filter mode")
	}
	if !strings.Contains(view, "test.sh") {
		t.Error("expected view to contain 'test.sh' (dimmed) in filter mode")
	}

	// Filter bar should NOT be shown in filter mode (overlay handles it)
	if strings.Contains(view, "Filter:") {
		t.Error("expected filter bar to NOT be shown when filterMode is true (overlay mode)")
	}
}

func TestScriptsModel_NoFilterMode_ViewHidesNonMatches(t *testing.T) {
	m := NewScriptsModel()
	m.SetSize(80, 25)

	scripts := []core.Script{
		{Name: "deploy.sh", Description: "Deploy to production"},
		{Name: "backup.sh", Description: "Backup database"},
		{Name: "test.sh", Description: "Run tests"},
	}
	m.SetScripts(scripts)

	// Apply filter WITHOUT enabling filter mode (normal behavior)
	m.SetFilterMode(false)
	m.ApplyFilter("deploy")

	view := m.View()

	// Only matching script should be visible
	if !strings.Contains(view, "deploy.sh") {
		t.Error("expected view to contain 'deploy.sh' when filtered")
	}

	// Non-matching scripts should NOT be visible (hidden, not dimmed)
	if strings.Contains(view, "backup.sh") {
		t.Error("expected 'backup.sh' to be hidden (not shown) when not in filter mode")
	}
	if strings.Contains(view, "test.sh") {
		t.Error("expected 'test.sh' to be hidden (not shown) when not in filter mode")
	}

	// Filter bar SHOULD be shown when not in overlay mode
	if !strings.Contains(view, "Filter:") {
		t.Error("expected filter bar to be shown when filterMode is false")
	}
}
