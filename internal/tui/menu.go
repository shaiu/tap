package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

// CategoryItem implements list.Item for categories.
type CategoryItem struct {
	Category core.Category
}

// Title returns the category name.
func (i CategoryItem) Title() string {
	return i.Category.Name
}

// Description returns the script count.
func (i CategoryItem) Description() string {
	return fmt.Sprintf("%d scripts", len(i.Category.Scripts))
}

// FilterValue returns the string used for filtering.
func (i CategoryItem) FilterValue() string {
	return i.Category.Name
}

// ScriptItem implements list.Item for scripts.
type ScriptItem struct {
	Script core.Script
}

// Title returns the script name.
func (i ScriptItem) Title() string {
	return i.Script.Name
}

// Description returns the script description.
func (i ScriptItem) Description() string {
	return i.Script.Description
}

// FilterValue returns the string used for filtering.
func (i ScriptItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s",
		i.Script.Name,
		i.Script.Description,
		strings.Join(i.Script.Tags, " "))
}

// CategoryDelegate renders category items in the list.
type CategoryDelegate struct{}

// Height returns the number of lines a category item takes.
func (d CategoryDelegate) Height() int { return 1 }

// Spacing returns the space between items.
func (d CategoryDelegate) Spacing() int { return 0 }

// Update handles delegate-specific updates.
func (d CategoryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a category item.
func (d CategoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	cat, ok := item.(CategoryItem)
	if !ok {
		return
	}

	// Build category display with icon and count
	icon := Icons.Category
	name := cat.Title()
	count := len(cat.Category.Scripts)

	var str string
	if index == m.Index() {
		// Selected: ● icon name (count)
		str = Styles.ItemSelected.Render(fmt.Sprintf("● %s %s (%d)", icon, name, count))
	} else {
		// Normal: icon name (count) with dimmed count
		nameStr := Styles.Item.Render(fmt.Sprintf("  %s %s ", icon, name))
		countStr := Styles.ItemDesc.Render(fmt.Sprintf("(%d)", count))
		str = nameStr + countStr
	}
	fmt.Fprint(w, str)
}

// ScriptDelegate renders script items in the list.
type ScriptDelegate struct{}

// Height returns the number of lines a script item takes.
func (d ScriptDelegate) Height() int { return 2 }

// Spacing returns the space between items.
func (d ScriptDelegate) Spacing() int { return 1 }

// Update handles delegate-specific updates.
func (d ScriptDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a script item.
func (d ScriptDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	script, ok := item.(ScriptItem)
	if !ok {
		return
	}

	var s strings.Builder

	// Get shell-specific icon
	icon := IconForShell(script.Script.Shell)

	// Title line with icon prefix
	name := script.Title()
	if index == m.Index() {
		// Selected: ● icon name (bold, highlighted)
		title := Styles.ItemSelected.Render(fmt.Sprintf("● %s %s", icon, name))
		s.WriteString(title)
	} else {
		// Normal: icon name
		title := Styles.Item.Render(fmt.Sprintf("  %s %s", icon, name))
		s.WriteString(title)
	}
	s.WriteString("\n")

	// Description line (indented to align with name)
	desc := Styles.ItemDesc.Render("      " + script.Description())
	s.WriteString(desc)

	fmt.Fprint(w, s.String())
}

// MenuModel handles the category and script list navigation.
type MenuModel struct {
	// State
	categories     []core.Category
	selectedCatIdx int

	// Lists
	categoryList list.Model
	scriptList   list.Model

	// Dimensions
	width  int
	height int

	// Results
	selectedScript *core.Script

	// View mode
	showingScripts bool
}

// NewMenuModel creates a new MenuModel with the given categories.
func NewMenuModel(categories []core.Category, width, height int) MenuModel {
	// Create category list items
	catItems := make([]list.Item, len(categories))
	for i, cat := range categories {
		catItems[i] = CategoryItem{Category: cat}
	}

	// Calculate list height (leaving room for header and footer)
	listHeight := height - 6
	if listHeight < 5 {
		listHeight = 5
	}

	// Create category list
	catList := list.New(catItems, CategoryDelegate{}, width, listHeight)
	catList.Title = "Categories"
	catList.SetShowStatusBar(false)
	catList.SetShowHelp(false)
	catList.SetFilteringEnabled(false) // We handle filtering in AppModel

	// Create empty script list (will be populated when drilling into a category)
	scriptList := list.New([]list.Item{}, ScriptDelegate{}, width, listHeight)
	scriptList.SetShowStatusBar(false)
	scriptList.SetShowHelp(false)
	scriptList.SetFilteringEnabled(false)

	return MenuModel{
		categories:     categories,
		selectedCatIdx: -1,
		categoryList:   catList,
		scriptList:     scriptList,
		width:          width,
		height:         height,
		showingScripts: false,
	}
}

// Init implements tea.Model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := m.height - 6
		if listHeight < 5 {
			listHeight = 5
		}
		m.categoryList.SetSize(msg.Width, listHeight)
		m.scriptList.SetSize(msg.Width, listHeight)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

// handleKeyMsg processes keyboard input.
func (m MenuModel) handleKeyMsg(msg tea.KeyMsg) (MenuModel, tea.Cmd) {
	if m.showingScripts {
		return m.updateScriptList(msg)
	}
	return m.updateCategoryList(msg)
}

// updateCategoryList handles input in the category list view.
func (m MenuModel) updateCategoryList(msg tea.KeyMsg) (MenuModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Drill into selected category
		if item, ok := m.categoryList.SelectedItem().(CategoryItem); ok {
			m.selectedCatIdx = m.categoryList.Index()
			m.scriptList = m.buildScriptList(item.Category)
			m.showingScripts = true
		}
		return m, nil
	}

	// Pass to list component for navigation (up/down/j/k)
	var cmd tea.Cmd
	m.categoryList, cmd = m.categoryList.Update(msg)
	return m, cmd
}

// updateScriptList handles input in the script list view.
func (m MenuModel) updateScriptList(msg tea.KeyMsg) (MenuModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Select script
		if item, ok := m.scriptList.SelectedItem().(ScriptItem); ok {
			script := item.Script
			m.selectedScript = &script
			return m, func() tea.Msg {
				return ScriptSelectedMsg{Script: script}
			}
		}
		return m, nil

	case "esc", "backspace":
		// Go back to categories
		m.showingScripts = false
		m.selectedCatIdx = -1
		return m, nil
	}

	// Pass to list component for navigation
	var cmd tea.Cmd
	m.scriptList, cmd = m.scriptList.Update(msg)
	return m, cmd
}

// buildScriptList creates a list.Model for the scripts in a category.
func (m MenuModel) buildScriptList(category core.Category) list.Model {
	items := make([]list.Item, len(category.Scripts))
	for i, script := range category.Scripts {
		items[i] = ScriptItem{Script: script}
	}

	listHeight := m.height - 6
	if listHeight < 5 {
		listHeight = 5
	}

	scriptList := list.New(items, ScriptDelegate{}, m.width, listHeight)
	scriptList.Title = category.Name
	scriptList.SetShowStatusBar(false)
	scriptList.SetShowHelp(false)
	scriptList.SetFilteringEnabled(false)

	return scriptList
}

// View renders the menu.
func (m MenuModel) View() string {
	var s strings.Builder

	// Header
	if m.showingScripts && m.selectedCatIdx >= 0 && m.selectedCatIdx < len(m.categories) {
		s.WriteString(Styles.Header.Render(fmt.Sprintf("tap - %s", m.categories[m.selectedCatIdx].Name)))
	} else {
		s.WriteString(Styles.Header.Render("tap - Script Runner"))
	}
	s.WriteString("\n\n")

	// Main content
	if m.showingScripts {
		s.WriteString(m.scriptList.View())
	} else {
		s.WriteString(m.categoryList.View())
	}

	// Footer
	s.WriteString("\n")
	s.WriteString(m.renderFooter())

	return s.String()
}

// renderFooter renders the key binding hints.
func (m MenuModel) renderFooter() string {
	var keys []string
	if m.showingScripts {
		keys = []string{"↑/↓ navigate", "enter run", "esc back", "/ filter", "q quit"}
	} else {
		keys = []string{"↑/↓ navigate", "enter select", "/ filter", "q quit"}
	}
	return Styles.Footer.Render(strings.Join(keys, "  "))
}

// ShowingScripts returns true if the menu is showing the script list.
func (m MenuModel) ShowingScripts() bool {
	return m.showingScripts
}

// SelectedCategoryIndex returns the index of the selected category (-1 if none).
func (m MenuModel) SelectedCategoryIndex() int {
	return m.selectedCatIdx
}

// SelectedScript returns the selected script, if any.
func (m MenuModel) SelectedScript() *core.Script {
	return m.selectedScript
}

// SetCategories updates the categories.
func (m *MenuModel) SetCategories(categories []core.Category) {
	m.categories = categories

	// Rebuild category list
	catItems := make([]list.Item, len(categories))
	for i, cat := range categories {
		catItems[i] = CategoryItem{Category: cat}
	}
	m.categoryList.SetItems(catItems)

	// Reset navigation state
	m.showingScripts = false
	m.selectedCatIdx = -1
}

// SetSize updates the menu dimensions.
func (m *MenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listHeight := height - 6
	if listHeight < 5 {
		listHeight = 5
	}
	m.categoryList.SetSize(width, listHeight)
	m.scriptList.SetSize(width, listHeight)
}

// ApplyFilter filters the current list based on the query string.
func (m *MenuModel) ApplyFilter(query string) {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		m.ClearFilter()
		return
	}

	if m.showingScripts {
		// Filter scripts in the selected category
		if m.selectedCatIdx >= 0 && m.selectedCatIdx < len(m.categories) {
			var filtered []list.Item
			for _, script := range m.categories[m.selectedCatIdx].Scripts {
				if fuzzyMatch(script, query) {
					filtered = append(filtered, ScriptItem{Script: script})
				}
			}
			m.scriptList.SetItems(filtered)
		}
	} else {
		// Filter categories
		var filtered []list.Item
		for _, cat := range m.categories {
			if strings.Contains(strings.ToLower(cat.Name), query) {
				filtered = append(filtered, CategoryItem{Category: cat})
			}
		}
		m.categoryList.SetItems(filtered)
	}
}

// ClearFilter restores the full list of items.
func (m *MenuModel) ClearFilter() {
	if m.showingScripts {
		// Restore full script list for selected category
		if m.selectedCatIdx >= 0 && m.selectedCatIdx < len(m.categories) {
			items := make([]list.Item, len(m.categories[m.selectedCatIdx].Scripts))
			for i, script := range m.categories[m.selectedCatIdx].Scripts {
				items[i] = ScriptItem{Script: script}
			}
			m.scriptList.SetItems(items)
		}
	} else {
		// Restore full category list
		catItems := make([]list.Item, len(m.categories))
		for i, cat := range m.categories {
			catItems[i] = CategoryItem{Category: cat}
		}
		m.categoryList.SetItems(catItems)
	}
}

// fuzzyMatch checks if a script matches the query string.
func fuzzyMatch(script core.Script, query string) bool {
	searchText := strings.ToLower(fmt.Sprintf("%s %s %s",
		script.Name,
		script.Description,
		strings.Join(script.Tags, " ")))

	// Simple contains match - could be upgraded to true fuzzy
	return strings.Contains(searchText, query)
}
