package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiungar/tap/internal/core"
)

// SidebarItemType indicates the type of sidebar item.
type SidebarItemType int

const (
	SidebarItemAllScripts SidebarItemType = iota
	SidebarItemCategory
	SidebarItemPinnedHeader
	SidebarItemPinnedScript
)

// SidebarItem represents an item in the sidebar (category, pinned script, or special entry).
type SidebarItem struct {
	Name     string
	Count    int
	Type     SidebarItemType
	Category *core.Category
	Script   *core.Script // For pinned scripts
}

// SidebarModel handles the category sidebar panel.
type SidebarModel struct {
	items         []SidebarItem
	cursor        int
	focused       bool
	width         int
	height        int
	pinnedScripts []core.Script
}

// NewSidebarModel creates a new SidebarModel with the given categories.
func NewSidebarModel(categories []core.Category) SidebarModel {
	items := buildSidebarItems(categories, nil)
	return SidebarModel{
		items:   items,
		cursor:  0,
		focused: true,
	}
}

// NewSidebarModelWithPinned creates a new SidebarModel with categories and pinned scripts.
func NewSidebarModelWithPinned(categories []core.Category, pinnedScripts []core.Script) SidebarModel {
	items := buildSidebarItems(categories, pinnedScripts)
	return SidebarModel{
		items:         items,
		cursor:        0,
		focused:       true,
		pinnedScripts: pinnedScripts,
	}
}

// buildSidebarItems creates sidebar items from categories and pinned scripts.
func buildSidebarItems(categories []core.Category, pinnedScripts []core.Script) []SidebarItem {
	// Count total scripts
	totalScripts := 0
	for _, cat := range categories {
		totalScripts += len(cat.Scripts)
	}

	// Estimate capacity: all scripts + categories + pinned header + pinned scripts
	capacity := len(categories) + 1
	if len(pinnedScripts) > 0 {
		capacity += 1 + len(pinnedScripts) // header + scripts
	}
	items := make([]SidebarItem, 0, capacity)

	// Add "All Scripts" at the top
	items = append(items, SidebarItem{
		Name:  "All Scripts",
		Count: totalScripts,
		Type:  SidebarItemAllScripts,
	})

	// Add categories
	for i := range categories {
		items = append(items, SidebarItem{
			Name:     categories[i].Name,
			Count:    len(categories[i].Scripts),
			Type:     SidebarItemCategory,
			Category: &categories[i],
		})
	}

	// Add pinned scripts section if there are any
	if len(pinnedScripts) > 0 {
		// Add pinned header (non-selectable visual separator)
		items = append(items, SidebarItem{
			Name:  "Pinned",
			Count: len(pinnedScripts),
			Type:  SidebarItemPinnedHeader,
		})

		// Add pinned scripts
		for i := range pinnedScripts {
			items = append(items, SidebarItem{
				Name:   pinnedScripts[i].Name,
				Type:   SidebarItemPinnedScript,
				Script: &pinnedScripts[i],
			})
		}
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
				// Skip the pinned header (not selectable)
				if m.cursor >= 0 && m.cursor < len(m.items) && m.items[m.cursor].Type == SidebarItemPinnedHeader {
					if m.cursor > 0 {
						m.cursor--
					} else {
						m.cursor++ // Can't go up, revert
					}
				}
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Skip the pinned header (not selectable)
				if m.cursor < len(m.items) && m.items[m.cursor].Type == SidebarItemPinnedHeader {
					if m.cursor < len(m.items)-1 {
						m.cursor++
					} else {
						m.cursor-- // Can't go down, revert
					}
				}
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

	// Calculate inner dimensions (content area)
	innerWidth, innerHeight := InnerDimensions(m.width, m.height)
	_ = innerWidth // Used for future width constraints

	// Build lines array
	var lines []string

	// Title (takes 2 lines: title + blank)
	title := Styles.Title.Render(fmt.Sprintf("%s Categories", Icons.Category))
	lines = append(lines, title)
	lines = append(lines, "")

	// Available height for items (inner height minus title lines)
	itemAreaHeight := innerHeight - 2
	if itemAreaHeight < 1 {
		itemAreaHeight = 1
	}

	// Calculate scroll offset to keep cursor visible
	scrollOffset := 0
	if m.cursor >= itemAreaHeight {
		scrollOffset = m.cursor - itemAreaHeight + 1
	}

	// Render visible items
	for i := scrollOffset; i < len(m.items) && i < scrollOffset+itemAreaHeight; i++ {
		item := m.items[i]
		rendered := m.renderItem(item, i == m.cursor)
		// renderItem may return multiple lines for pinned header
		itemLines := strings.Split(rendered, "\n")
		lines = append(lines, itemLines...)
	}

	// Pad content to exact inner height
	content := BuildPanelContent(lines, innerHeight)

	// Apply panel style - only set Width, NOT Height
	return panelStyle.
		Width(m.width - BorderWidth).
		Render(content)
}

// renderItem renders a single sidebar item.
func (m SidebarModel) renderItem(item SidebarItem, selected bool) string {
	// Handle pinned header specially (with separator)
	if item.Type == SidebarItemPinnedHeader {
		// Calculate separator width based on available space
		separatorWidth := m.width - 6 // Account for border and padding
		if separatorWidth < 5 {
			separatorWidth = 5
		}
		separator := strings.Repeat("─", separatorWidth)
		headerLine := fmt.Sprintf("  %s %s", Icons.Pin, item.Name)

		// Separator + newline + header
		return Styles.ItemDesc.Render(separator) + "\n" + Styles.Item.Render(headerLine)
	}

	// Handle pinned scripts (no count, indented)
	if item.Type == SidebarItemPinnedScript {
		if selected {
			return Styles.ItemSelected.Render(fmt.Sprintf("●     %s", item.Name))
		}
		return Styles.Item.Render(fmt.Sprintf("      %s", item.Name))
	}

	// Regular items (All Scripts, Categories)
	icon := Icons.Category
	if item.Type == SidebarItemAllScripts {
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

// SetCategories updates the categories while preserving pinned scripts.
func (m *SidebarModel) SetCategories(categories []core.Category) {
	m.items = buildSidebarItems(categories, m.pinnedScripts)
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// SetPinnedScripts updates the pinned scripts.
func (m *SidebarModel) SetPinnedScripts(scripts []core.Script) {
	m.pinnedScripts = scripts
	// Rebuild items with current categories (extracted from existing items)
	categories := m.extractCategories()
	m.items = buildSidebarItems(categories, scripts)
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// extractCategories extracts categories from existing sidebar items.
func (m SidebarModel) extractCategories() []core.Category {
	var categories []core.Category
	for _, item := range m.items {
		if item.Type == SidebarItemCategory && item.Category != nil {
			categories = append(categories, *item.Category)
		}
	}
	return categories
}

// PinnedScripts returns the current pinned scripts.
func (m SidebarModel) PinnedScripts() []core.Script {
	return m.pinnedScripts
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
		return m.items[m.cursor].Type == SidebarItemAllScripts
	}
	return false
}

// IsPinnedScriptSelected returns true if a pinned script is selected.
func (m SidebarModel) IsPinnedScriptSelected() bool {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return m.items[m.cursor].Type == SidebarItemPinnedScript
	}
	return false
}

// SelectedPinnedScript returns the selected pinned script, or nil if none is selected.
func (m SidebarModel) SelectedPinnedScript() *core.Script {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		if m.items[m.cursor].Type == SidebarItemPinnedScript {
			return m.items[m.cursor].Script
		}
	}
	return nil
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
	if selected.Type == SidebarItemAllScripts {
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
