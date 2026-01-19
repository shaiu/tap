# TUI Menu Spec

> Interactive menu navigation with filtering

## Overview

The TUI menu is tap's primary interface. It displays categories and scripts in a navigable list with fuzzy filtering, allowing users to quickly find and select scripts to run.

## User Flow

```
┌─────────────────────────────────────────┐
│  tap - Script Runner                    │
│                                         │
│  Categories:                            │
│  > deployment (5)                       │
│    data (3)                             │
│    maintenance (2)                      │
│    uncategorized (1)                    │
│                                         │
│  ↑/↓ navigate  enter select  / filter  │
│  q quit                                 │
└─────────────────────────────────────────┘
           │ enter
           ▼
┌─────────────────────────────────────────┐
│  tap - deployment                       │
│                                         │
│  > deploy                               │
│    Deploy application to environment    │
│                                         │
│    rollback                             │
│    Rollback to previous version         │
│                                         │
│    scale                                │
│    Scale deployment replicas            │
│                                         │
│  ↑/↓ navigate  enter run  esc back     │
│  / filter  q quit                       │
└─────────────────────────────────────────┘
           │ /
           ▼
┌─────────────────────────────────────────┐
│  tap - deployment                       │
│                                         │
│  Filter: dep█                           │
│                                         │
│  > deploy                               │
│    Deploy application to environment    │
│                                         │
│  enter select  esc cancel               │
└─────────────────────────────────────────┘
```

## Features

### Core Navigation
- **↑/↓ or j/k** — Move selection up/down
- **Enter** — Select category → drill into scripts; Select script → run (or show params)
- **Esc or Backspace** — Go back to parent view
- **q** — Quit tap

### Filtering
- **/** — Activate filter mode
- **Type** — Fuzzy filter visible items
- **Enter** — Select top match and proceed
- **Esc** — Cancel filter, restore full list

### Quick Actions
- **?** — Show help overlay
- **r** — Refresh (rescan directories)

## State Machine

```
                    ┌─────────────┐
                    │   Loading   │
                    └──────┬──────┘
                           │ scripts loaded
                           ▼
                    ┌─────────────┐
            ┌──────►│  Category   │◄──────┐
            │       │    List     │       │
            │       └──────┬──────┘       │
            │              │ enter        │ esc/backspace
            │              ▼              │
            │       ┌─────────────┐       │
            │       │   Script    │───────┘
            │       │    List     │
            │       └──────┬──────┘
            │              │ enter (no params)
            │              ▼
            │       ┌─────────────┐
            │       │  Executing  │
            │       └──────┬──────┘
            │              │ done
            │              ▼
            │       ┌─────────────┐
            └───────┤    Exit     │
                    └─────────────┘

Note: If script has parameters, goes to Form view first (see 04-tui-params.md)
```

## Data Structures

### View State

```go
type viewState int

const (
    stateLoading viewState = iota
    stateCategoryList
    stateScriptList
    stateFilter
    stateHelp
)
```

### Menu Model

```go
type MenuModel struct {
    // State
    state           viewState
    categories      []core.Category
    selectedCatIdx  int
    
    // Lists
    categoryList    list.Model
    scriptList      list.Model
    
    // Filter
    filterInput     textinput.Model
    filterActive    bool
    
    // Dimensions
    width           int
    height          int
    
    // Results
    selectedScript  *core.Script
}
```

### List Items

```go
// CategoryItem implements list.Item for categories
type CategoryItem struct {
    category core.Category
}

func (i CategoryItem) Title() string {
    return i.category.Name
}

func (i CategoryItem) Description() string {
    return fmt.Sprintf("%d scripts", len(i.category.Scripts))
}

func (i CategoryItem) FilterValue() string {
    return i.category.Name
}

// ScriptItem implements list.Item for scripts
type ScriptItem struct {
    script core.Script
}

func (i ScriptItem) Title() string {
    return i.script.Name
}

func (i ScriptItem) Description() string {
    return i.script.Description
}

func (i ScriptItem) FilterValue() string {
    // Include name, description, and tags for filtering
    return fmt.Sprintf("%s %s %s", 
        i.script.Name, 
        i.script.Description,
        strings.Join(i.script.Tags, " "))
}
```

## Implementation

### Initialization

```go
func NewMenuModel(categories []core.Category, width, height int) MenuModel {
    // Create category list
    catItems := make([]list.Item, len(categories))
    for i, cat := range categories {
        catItems[i] = CategoryItem{category: cat}
    }
    
    catList := list.New(catItems, CategoryDelegate{}, width, height-4)
    catList.Title = "Categories"
    catList.SetShowHelp(false)
    catList.SetFilteringEnabled(false) // We handle filtering ourselves
    
    // Create filter input
    filter := textinput.New()
    filter.Placeholder = "Filter..."
    filter.CharLimit = 50
    
    return MenuModel{
        state:        stateCategoryList,
        categories:   categories,
        categoryList: catList,
        filterInput:  filter,
        width:        width,
        height:       height,
    }
}
```

### Update Logic

```go
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.categoryList.SetSize(msg.Width, msg.Height-4)
        m.scriptList.SetSize(msg.Width, msg.Height-4)
        return m, nil
        
    case tea.KeyMsg:
        // Global keys
        switch msg.String() {
        case "ctrl+c", "q":
            if !m.filterActive {
                return m, tea.Quit
            }
        case "?":
            if !m.filterActive {
                m.state = stateHelp
                return m, nil
            }
        }
        
        // State-specific handling
        switch m.state {
        case stateCategoryList:
            return m.updateCategoryList(msg)
        case stateScriptList:
            return m.updateScriptList(msg)
        case stateFilter:
            return m.updateFilter(msg)
        case stateHelp:
            return m.updateHelp(msg)
        }
    }
    
    return m, nil
}

func (m MenuModel) updateCategoryList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        // Drill into selected category
        if item, ok := m.categoryList.SelectedItem().(CategoryItem); ok {
            m.selectedCatIdx = m.categoryList.Index()
            m.scriptList = m.buildScriptList(item.category)
            m.state = stateScriptList
        }
        return m, nil
        
    case "/":
        // Activate filter
        m.filterActive = true
        m.filterInput.SetValue("")
        m.filterInput.Focus()
        m.state = stateFilter
        return m, textinput.Blink
    }
    
    // Pass to list component
    var cmd tea.Cmd
    m.categoryList, cmd = m.categoryList.Update(msg)
    return m, cmd
}

func (m MenuModel) updateScriptList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        // Select script
        if item, ok := m.scriptList.SelectedItem().(ScriptItem); ok {
            m.selectedScript = &item.script
            return m, func() tea.Msg {
                return ScriptSelectedMsg{Script: item.script}
            }
        }
        return m, nil
        
    case "esc", "backspace":
        // Go back to categories
        m.state = stateCategoryList
        return m, nil
        
    case "/":
        // Activate filter
        m.filterActive = true
        m.filterInput.SetValue("")
        m.filterInput.Focus()
        m.state = stateFilter
        return m, textinput.Blink
    }
    
    var cmd tea.Cmd
    m.scriptList, cmd = m.scriptList.Update(msg)
    return m, cmd
}
```

### Filtering

```go
func (m MenuModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        // Select top match
        m.filterActive = false
        m.filterInput.Blur()
        // Keep filtered results, select first item
        return m, nil
        
    case "esc":
        // Cancel filter, restore full list
        m.filterActive = false
        m.filterInput.Blur()
        m.restoreFullList()
        return m, nil
    }
    
    // Update filter input
    var cmd tea.Cmd
    m.filterInput, cmd = m.filterInput.Update(msg)
    
    // Apply filter to current list
    m.applyFilter(m.filterInput.Value())
    
    return m, cmd
}

func (m *MenuModel) applyFilter(query string) {
    if query == "" {
        m.restoreFullList()
        return
    }
    
    query = strings.ToLower(query)
    
    switch m.state {
    case stateFilter:
        // Filter the appropriate list based on previous state
        if m.selectedCatIdx >= 0 {
            // Filtering scripts
            var filtered []list.Item
            for _, script := range m.categories[m.selectedCatIdx].Scripts {
                if fuzzyMatch(script, query) {
                    filtered = append(filtered, ScriptItem{script: script})
                }
            }
            m.scriptList.SetItems(filtered)
        } else {
            // Filtering categories
            var filtered []list.Item
            for _, cat := range m.categories {
                if strings.Contains(strings.ToLower(cat.Name), query) {
                    filtered = append(filtered, CategoryItem{category: cat})
                }
            }
            m.categoryList.SetItems(filtered)
        }
    }
}

func fuzzyMatch(script core.Script, query string) bool {
    searchText := strings.ToLower(fmt.Sprintf("%s %s %s",
        script.Name,
        script.Description,
        strings.Join(script.Tags, " ")))
    
    // Simple contains match - could be upgraded to true fuzzy
    return strings.Contains(searchText, query)
}
```

### View Rendering

```go
func (m MenuModel) View() string {
    var s strings.Builder
    
    // Header
    header := m.renderHeader()
    s.WriteString(header)
    s.WriteString("\n\n")
    
    // Filter bar (if active)
    if m.filterActive {
        s.WriteString(m.filterInput.View())
        s.WriteString("\n\n")
    }
    
    // Main content
    switch m.state {
    case stateLoading:
        s.WriteString("Loading scripts...")
    case stateCategoryList, stateFilter:
        if m.selectedCatIdx < 0 {
            s.WriteString(m.categoryList.View())
        } else {
            s.WriteString(m.scriptList.View())
        }
    case stateScriptList:
        s.WriteString(m.scriptList.View())
    case stateHelp:
        s.WriteString(m.renderHelp())
    }
    
    // Footer with keybindings
    s.WriteString("\n")
    s.WriteString(m.renderFooter())
    
    return s.String()
}

func (m MenuModel) renderHeader() string {
    title := "tap"
    if m.state == stateScriptList && m.selectedCatIdx >= 0 {
        title = fmt.Sprintf("tap - %s", m.categories[m.selectedCatIdx].Name)
    }
    
    return styles.Header.Render(title)
}

func (m MenuModel) renderFooter() string {
    var keys []string
    
    switch m.state {
    case stateCategoryList:
        keys = []string{"↑/↓ navigate", "enter select", "/ filter", "q quit"}
    case stateScriptList:
        keys = []string{"↑/↓ navigate", "enter run", "esc back", "/ filter", "q quit"}
    case stateFilter:
        keys = []string{"enter select", "esc cancel"}
    }
    
    return styles.Footer.Render(strings.Join(keys, "  "))
}

func (m MenuModel) renderHelp() string {
    help := `
Keyboard Shortcuts

Navigation
  ↑/k       Move up
  ↓/j       Move down
  enter     Select / Run
  esc       Go back
  
Filtering
  /         Start filtering
  esc       Cancel filter
  enter     Select match
  
Other
  ?         Show this help
  r         Refresh scripts
  q         Quit
`
    return styles.Help.Render(help)
}
```

## Styles

Using Lipgloss for consistent styling:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var styles = struct {
    Header    lipgloss.Style
    Footer    lipgloss.Style
    Help      lipgloss.Style
    Selected  lipgloss.Style
    Dimmed    lipgloss.Style
}{
    Header: lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("212")).
        MarginBottom(1),
        
    Footer: lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")).
        MarginTop(1),
        
    Help: lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")).
        Padding(1, 2),
        
    Selected: lipgloss.NewStyle().
        Foreground(lipgloss.Color("212")).
        Bold(true),
        
    Dimmed: lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")),
}
```

## Custom List Delegate

For prettier script rendering:

```go
type ScriptDelegate struct{}

func (d ScriptDelegate) Height() int { return 2 }
func (d ScriptDelegate) Spacing() int { return 1 }
func (d ScriptDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d ScriptDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    script, ok := item.(ScriptItem)
    if !ok {
        return
    }
    
    var s strings.Builder
    
    // Title line
    title := script.Title()
    if index == m.Index() {
        title = styles.Selected.Render("> " + title)
    } else {
        title = "  " + title
    }
    s.WriteString(title)
    s.WriteString("\n")
    
    // Description line (indented)
    desc := styles.Dimmed.Render("  " + script.Description())
    s.WriteString(desc)
    
    fmt.Fprint(w, s.String())
}
```

## Messages

```go
// ScriptSelectedMsg is sent when a script is chosen
type ScriptSelectedMsg struct {
    Script core.Script
}

// RefreshMsg triggers a rescan
type RefreshMsg struct{}

// ErrorMsg carries error information
type ErrorMsg struct {
    Err error
}
```

## Testing

### Unit Tests

```go
func TestMenuModel_Navigation(t *testing.T)
func TestMenuModel_CategoryToDrillDown(t *testing.T)
func TestMenuModel_BackNavigation(t *testing.T)
func TestMenuModel_FilterActivation(t *testing.T)
func TestMenuModel_FilterMatching(t *testing.T)
func TestMenuModel_FilterCancel(t *testing.T)
func TestMenuModel_ScriptSelection(t *testing.T)
func TestMenuModel_EmptyCategories(t *testing.T)
func TestMenuModel_WindowResize(t *testing.T)
```

### Visual Testing

Consider using `teatest` for snapshot testing of rendered output:

```go
func TestMenuModel_Render(t *testing.T) {
    model := NewMenuModel(testCategories, 80, 24)
    
    // Snapshot initial render
    golden := "testdata/menu_initial.txt"
    actual := model.View()
    
    if *update {
        os.WriteFile(golden, []byte(actual), 0644)
    }
    
    expected, _ := os.ReadFile(golden)
    if actual != string(expected) {
        t.Errorf("render mismatch")
    }
}
```
