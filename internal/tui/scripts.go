package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

// ScriptsModel handles the scripts list panel.
type ScriptsModel struct {
	scripts       []core.Script
	allScripts    []core.Script // All scripts for "All Scripts" view
	cursor        int
	focused       bool
	width         int
	height        int
	filterQuery   string
	filteredList  []core.Script
	matchedSet    map[int]bool // Indices of scripts that match the filter
	filterMode    bool         // True when filter overlay is active (show all with dimming)
}

// NewScriptsModel creates a new ScriptsModel.
func NewScriptsModel() ScriptsModel {
	return ScriptsModel{
		scripts:      []core.Script{},
		allScripts:   []core.Script{},
		cursor:       0,
		focused:      false,
		filteredList: []core.Script{},
		matchedSet:   make(map[int]bool),
		filterMode:   false,
	}
}

// Init implements tea.Model.
func (m ScriptsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ScriptsModel) Update(msg tea.Msg) (ScriptsModel, tea.Cmd) {
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
			list := m.displayList()
			if m.cursor < len(list)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the scripts panel.
func (m ScriptsModel) View() string {
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

	// Title (takes 1 line)
	title := Styles.Title.Render(fmt.Sprintf("%s Scripts", Icons.Script))
	lines = append(lines, title)

	// Track header lines for item area calculation
	headerLines := 1

	// Filter bar (shown when filter is active but NOT in overlay mode)
	if m.filterQuery != "" && !m.filterMode {
		filterBar := m.renderFilterBar()
		lines = append(lines, filterBar)
		headerLines++
	}

	// Blank line after header
	lines = append(lines, "")
	headerLines++

	// In filter mode, show ALL scripts with non-matches dimmed
	// Otherwise, show filtered list
	list := m.displayListForView()

	if len(list) == 0 {
		if m.filterQuery != "" {
			lines = append(lines, Styles.ItemDesc.Render("  No matching scripts"))
		} else {
			lines = append(lines, Styles.ItemDesc.Render("  No scripts found"))
		}
	} else {
		// Available height for items (each item takes 2 lines: name + desc)
		itemAreaHeight := innerHeight - headerLines
		if itemAreaHeight < 2 {
			itemAreaHeight = 2
		}
		// Number of items that fit (2 lines per item + 1 blank between items = 3 lines)
		visibleItems := itemAreaHeight / 3
		if visibleItems < 1 {
			visibleItems = 1
		}

		// Calculate scroll offset to keep cursor visible
		scrollOffset := 0
		if m.cursor >= visibleItems {
			scrollOffset = m.cursor - visibleItems + 1
		}

		// Render visible items
		for i := scrollOffset; i < len(list) && i < scrollOffset+visibleItems; i++ {
			script := list[i]
			isMatch := m.IsMatched(i)
			rendered := m.renderItemWithFilter(script, i == m.cursor, isMatch)
			itemLines := strings.Split(rendered, "\n")
			lines = append(lines, itemLines...)
			// Add blank line between items (but not after last)
			if i < len(list)-1 && i < scrollOffset+visibleItems-1 {
				lines = append(lines, "")
			}
		}
	}

	// Pad content to exact inner height
	content := BuildPanelContent(lines, innerHeight)

	// Apply panel style - only set Width, NOT Height
	return panelStyle.
		Width(m.width - BorderWidth).
		Render(content)
}

// displayListForView returns the list to display for rendering.
// In filter mode, returns all scripts (non-matches will be dimmed).
// Otherwise, returns the filtered list.
func (m ScriptsModel) displayListForView() []core.Script {
	if m.filterMode {
		return m.scripts // Show all, dimming handled in renderItem
	}
	return m.displayList()
}

// renderFilterBar renders the filter bar with query and match count.
func (m ScriptsModel) renderFilterBar() string {
	// Format: "Filter: query                [3/12]"
	query := Styles.FilterQuery.Render(m.filterQuery)
	count := Styles.FilterCount.Render(fmt.Sprintf("[%d/%d]", len(m.filteredList), len(m.scripts)))

	// Calculate padding to right-align the count
	filterPrefix := fmt.Sprintf("%s Filter: %s", Icons.Search, query)
	prefixLen := len(fmt.Sprintf("%s Filter: %s", Icons.Search, m.filterQuery))
	countLen := len(fmt.Sprintf("[%d/%d]", len(m.filteredList), len(m.scripts)))

	// Available width for spacing (accounting for panel padding)
	availableWidth := m.width - 6 // Account for borders and padding
	spacingLen := availableWidth - prefixLen - countLen
	if spacingLen < 1 {
		spacingLen = 1
	}

	spacing := strings.Repeat(" ", spacingLen)
	return filterPrefix + spacing + count
}

// renderItem renders a single script item (2-line format).
func (m ScriptsModel) renderItem(script core.Script, selected bool) string {
	return m.renderItemWithFilter(script, selected, true) // true = matches (normal rendering)
}

// renderItemWithFilter renders a single script item with filter awareness.
// If isMatch is false and filter is active, the item is dimmed.
func (m ScriptsModel) renderItemWithFilter(script core.Script, selected, isMatch bool) string {
	var s strings.Builder

	// Get shell-specific icon
	icon := IconForShell(script.Shell)

	// Determine styling based on filter state
	isDimmed := m.filterMode && m.filterQuery != "" && !isMatch

	// Title line with icon
	if selected && !isDimmed {
		s.WriteString(Styles.ItemSelected.Render(fmt.Sprintf("● %s %s", icon, script.Name)))
	} else if isDimmed {
		s.WriteString(Styles.ItemDimmed.Render(fmt.Sprintf("  %s %s", icon, script.Name)))
	} else if isMatch && m.filterMode && m.filterQuery != "" {
		// Highlighted match in filter mode
		s.WriteString(Styles.ItemMatch.Render(fmt.Sprintf("  %s %s", icon, script.Name)))
	} else {
		s.WriteString(Styles.Item.Render(fmt.Sprintf("  %s %s", icon, script.Name)))
	}
	s.WriteString("\n")

	// Description line (indented)
	desc := script.Description
	if desc == "" {
		desc = "No description"
	}
	// Truncate description if too long
	maxDescLen := m.width - 10
	if maxDescLen < 20 {
		maxDescLen = 20
	}
	if len(desc) > maxDescLen {
		desc = desc[:maxDescLen-3] + "..."
	}

	if isDimmed {
		s.WriteString(Styles.ItemDimmed.Render("      " + desc))
	} else if isMatch && m.filterMode && m.filterQuery != "" {
		s.WriteString(Styles.ItemMatchDesc.Render("      " + desc))
	} else {
		s.WriteString(Styles.ItemDesc.Render("      " + desc))
	}

	return s.String()
}

// displayList returns the list to display (filtered or full).
func (m ScriptsModel) displayList() []core.Script {
	if m.filterQuery != "" {
		return m.filteredList // May be empty if no matches
	}
	return m.scripts
}

// SetFocused sets whether the panel is focused.
func (m *ScriptsModel) SetFocused(focused bool) {
	m.focused = focused
}

// IsFocused returns whether the panel is focused.
func (m ScriptsModel) IsFocused() bool {
	return m.focused
}

// SetSize updates the panel dimensions.
func (m *ScriptsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetScripts updates the scripts list for a specific category.
func (m *ScriptsModel) SetScripts(scripts []core.Script) {
	m.scripts = scripts
	m.filteredList = nil
	m.filterQuery = ""
	if m.cursor >= len(scripts) {
		m.cursor = len(scripts) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// SetAllScripts sets all scripts (used when "All Scripts" is selected).
func (m *ScriptsModel) SetAllScripts(categories []core.Category) {
	var all []core.Script
	for _, cat := range categories {
		all = append(all, cat.Scripts...)
	}
	m.allScripts = all
	m.SetScripts(all)
}

// ApplyFilter filters the scripts based on the query.
func (m *ScriptsModel) ApplyFilter(query string) {
	m.filterQuery = strings.ToLower(strings.TrimSpace(query))
	m.matchedSet = make(map[int]bool)

	if m.filterQuery == "" {
		m.filteredList = nil
		return
	}

	m.filteredList = nil
	for i, script := range m.scripts {
		if m.matchesFilter(script) {
			m.filteredList = append(m.filteredList, script)
			m.matchedSet[i] = true
		}
	}

	// In filter mode, cursor stays within matched items only
	if !m.filterMode {
		// Reset cursor if out of bounds
		if m.cursor >= len(m.filteredList) {
			m.cursor = len(m.filteredList) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}

// ClearFilter removes the filter.
func (m *ScriptsModel) ClearFilter() {
	m.filterQuery = ""
	m.filteredList = nil
	m.matchedSet = make(map[int]bool)
	m.filterMode = false
}

// SetFilterMode enables or disables filter mode.
// In filter mode, all scripts are shown with non-matches dimmed.
func (m *ScriptsModel) SetFilterMode(enabled bool) {
	m.filterMode = enabled
}

// IsFilterMode returns whether filter mode is active.
func (m ScriptsModel) IsFilterMode() bool {
	return m.filterMode
}

// IsMatched checks if a script at the given index matches the current filter.
func (m ScriptsModel) IsMatched(idx int) bool {
	if m.filterQuery == "" {
		return true // No filter means all match
	}
	return m.matchedSet[idx]
}

// matchesFilter checks if a script matches the current filter.
func (m ScriptsModel) matchesFilter(script core.Script) bool {
	searchText := strings.ToLower(fmt.Sprintf("%s %s %s",
		script.Name,
		script.Description,
		strings.Join(script.Tags, " ")))
	return strings.Contains(searchText, m.filterQuery)
}

// SelectedScript returns the currently selected script, or nil if none.
func (m ScriptsModel) SelectedScript() *core.Script {
	list := m.displayList()
	if m.cursor >= 0 && m.cursor < len(list) {
		script := list[m.cursor]
		return &script
	}
	return nil
}

// Cursor returns the current cursor position.
func (m ScriptsModel) Cursor() int {
	return m.cursor
}

// ScriptCount returns the number of scripts in the current view.
func (m ScriptsModel) ScriptCount() int {
	return len(m.displayList())
}

// TotalCount returns the total number of scripts (before filtering).
func (m ScriptsModel) TotalCount() int {
	return len(m.scripts)
}

// FilterQuery returns the current filter query.
func (m ScriptsModel) FilterQuery() string {
	return m.filterQuery
}

// PanelTitle returns the title for this panel.
func (m ScriptsModel) PanelTitle() string {
	return "Scripts"
}
