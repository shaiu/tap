package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiungar/tap/internal/core"
)

// SidebarItem represents an item in the sidebar (category or special entry).
type SidebarItem struct {
	Name        string
	Count       int
	IsAllScripts bool
	Category    *core.Category
}

// SidebarModel handles the category sidebar panel.
type SidebarModel struct {
	items    []SidebarItem
	cursor   int
	focused  bool
	width    int
	height   int
}

// NewSidebarModel creates a new SidebarModel with the given categories.
func NewSidebarModel(categories []core.Category) SidebarModel {
	items := buildSidebarItems(categories)
	return SidebarModel{
		items:   items,
		cursor:  0,
		focused: true,
	}
}

// buildSidebarItems creates sidebar items from categories.
func buildSidebarItems(categories []core.Category) []SidebarItem {
	// Count total scripts
	totalScripts := 0
	for _, cat := range categories {
		totalScripts += len(cat.Scripts)
	}

	items := make([]SidebarItem, 0, len(categories)+1)

	// Add "All Scripts" at the top
	items = append(items, SidebarItem{
		Name:         "All Scripts",
		Count:        totalScripts,
		IsAllScripts: true,
	})

	// Add categories
	for i := range categories {
		items = append(items, SidebarItem{
			Name:     categories[i].Name,
			Count:    len(categories[i].Scripts),
			Category: &categories[i],
		})
	}

	return items
}

// Init implements tea.Model.
func (m SidebarModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap().Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the sidebar panel.
func (m SidebarModel) View() string {
	// Determine panel style based on focus
	panelStyle := Styles.Panel
	if m.focused {
		panelStyle = Styles.PanelActive
	}

	// Build content
	var content strings.Builder

	// Title
	title := Styles.Title.Render(fmt.Sprintf("%s Categories", Icons.Category))
	content.WriteString(title)
	content.WriteString("\n\n")

	// Visible height for items (accounting for title, padding, borders)
	visibleHeight := m.height - 6
	if visibleHeight < 3 {
		visibleHeight = 3
	}

	// Calculate scroll offset to keep cursor visible
	scrollOffset := 0
	if m.cursor >= visibleHeight {
		scrollOffset = m.cursor - visibleHeight + 1
	}

	// Render visible items
	for i := scrollOffset; i < len(m.items) && i < scrollOffset+visibleHeight; i++ {
		item := m.items[i]
		content.WriteString(m.renderItem(item, i == m.cursor))
		if i < len(m.items)-1 && i < scrollOffset+visibleHeight-1 {
			content.WriteString("\n")
		}
	}

	// Apply panel style
	innerWidth := m.width - 4 // Account for border and padding
	if innerWidth < 10 {
		innerWidth = 10
	}

	return panelStyle.
		Width(m.width).
		Height(m.height).
		Render(content.String())
}

// renderItem renders a single sidebar item.
func (m SidebarModel) renderItem(item SidebarItem, selected bool) string {
	icon := Icons.Category
	if item.IsAllScripts {
		icon = Icons.Script
	}

	name := item.Name
	count := fmt.Sprintf("(%d)", item.Count)

	if selected {
		// Selected: ● icon name (count)
		return Styles.ItemSelected.Render(fmt.Sprintf("● %s %s %s", icon, name, count))
	}

	// Normal: icon name (count) with dimmed count
	nameStr := Styles.Item.Render(fmt.Sprintf("  %s %s ", icon, name))
	countStr := Styles.ItemDesc.Render(count)
	return nameStr + countStr
}

// SetFocused sets whether the panel is focused.
func (m *SidebarModel) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the panel is focused.
func (m SidebarModel) IsFocused() bool {
	return m.focused
}

// SetSize updates the panel dimensions.
func (m *SidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetCategories updates the categories.
func (m *SidebarModel) SetCategories(categories []core.Category) {
	m.items = buildSidebarItems(categories)
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// SelectedItem returns the currently selected item.
func (m SidebarModel) SelectedItem() SidebarItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return m.items[m.cursor]
	}
	return SidebarItem{}
}

// SelectedCategory returns the selected category, or nil if "All Scripts" is selected.
func (m SidebarModel) SelectedCategory() *core.Category {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return m.items[m.cursor].Category
	}
	return nil
}

// IsAllScriptsSelected returns true if "All Scripts" is selected.
func (m SidebarModel) IsAllScriptsSelected() bool {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return m.items[m.cursor].IsAllScripts
	}
	return false
}

// Cursor returns the current cursor position.
func (m SidebarModel) Cursor() int {
	return m.cursor
}

// PanelTitle returns the title for this panel.
func (m SidebarModel) PanelTitle() string {
	return "Categories"
}

// Width returns the panel width.
func (m SidebarModel) Width() int {
	return m.width
}

// Height returns the panel height.
func (m SidebarModel) Height() int {
	return m.height
}

// RenderCompact renders a compact version for narrow terminals.
func (m SidebarModel) RenderCompact() string {
	selected := m.SelectedItem()
	if selected.IsAllScripts {
		return lipgloss.NewStyle().
			Foreground(Theme.Primary).
			Bold(true).
			Render("All Scripts")
	}
	return lipgloss.NewStyle().
		Foreground(Theme.Primary).
		Bold(true).
		Render(selected.Name)
}
