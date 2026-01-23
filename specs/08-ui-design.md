# tap UI Design Spec

> Inspired by [superfile](https://github.com/yorukot/superfile) - a modern terminal file manager

## Design Philosophy

tap adopts superfile's approach to terminal UIs: **beautiful by default, functional always**. The interface should feel polished and modern while remaining keyboard-driven and efficient.

## Visual Reference

```
┌─ Categories ──────────┬─ Scripts ─────────────────────────────────┬─ Details ─────────────────┐
│                       │                                           │                           │
│  󰉋  All Scripts (12)  │   deploy-staging.sh                      │  deploy-staging.sh       │
│                       │   Deploy to staging environment           │                           │
│  󰒍  Deploy (4)        │                                           │  Deploy application to   │
│  󰊤  Database (3)      │ ● backup-database.sh                      │  the staging server      │
│  󰑮  Testing (2)       │   Backup PostgreSQL database              │                           │
│  󰒲  Utils (3)         │                                           │  Category: Deploy         │
│                       │   run-tests.sh                            │  Shell: bash              │
│ ─────────────────────│   Execute test suite                      │                           │
│  󰐕  Pinned            │                                           │  Parameters:              │
│                       │   seed-database.sh                        │  • branch (required)      │
│     quick-deploy      │   Seed development database               │  • skip-tests (optional)  │
│                       │                                           │                           │
│                       │                                           │                           │
├───────────────────────┴───────────────────────────────────────────┴───────────────────────────┤
│  ↑↓ navigate  enter select  / filter  tab switch panel  ? help  q quit                       │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Layout Structure

### Three-Panel Layout (Default)

| Panel | Width | Purpose |
|-------|-------|---------|
| **Sidebar** | 20-25% | Categories, pinned scripts |
| **Main** | 40-50% | Script list with selection |
| **Details** | 30-35% | Selected script info, parameters |

### Two-Panel Layout (Compact)

For narrower terminals (<100 cols), collapse to sidebar + main only.

### Single Panel (Minimal)

For very narrow terminals (<60 cols), show only main list.

## Color Palette

Based on Catppuccin Mocha (superfile's default):

```go
// Theme colors - Catppuccin Mocha inspired
var Theme = struct {
    // Base
    Background    lipgloss.Color  // #1e1e2e - dark background
    Foreground    lipgloss.Color  // #cdd6f4 - light text
    Subtle        lipgloss.Color  // #6c7086 - muted text
    
    // Accent colors
    Primary       lipgloss.Color  // #89b4fa - blue (active/selected)
    Secondary     lipgloss.Color  // #cba6f7 - mauve (secondary accent)
    Success       lipgloss.Color  // #a6e3a1 - green (success states)
    Warning       lipgloss.Color  // #f9e2af - yellow (warnings)
    Error         lipgloss.Color  // #f38ba8 - red (errors)
    
    // UI Elements
    Border        lipgloss.Color  // #6c7086 - inactive borders
    BorderActive  lipgloss.Color  // #b4befe - active panel border
    Selection     lipgloss.Color  // #313244 - selection background
    
    // Gradient (for title bars)
    GradientStart lipgloss.Color  // #89b4fa
    GradientEnd   lipgloss.Color  // #cba6f7
}{
    Background:    lipgloss.Color("#1e1e2e"),
    Foreground:    lipgloss.Color("#cdd6f4"),
    Subtle:        lipgloss.Color("#6c7086"),
    Primary:       lipgloss.Color("#89b4fa"),
    Secondary:     lipgloss.Color("#cba6f7"),
    Success:       lipgloss.Color("#a6e3a1"),
    Warning:       lipgloss.Color("#f9e2af"),
    Error:         lipgloss.Color("#f38ba8"),
    Border:        lipgloss.Color("#6c7086"),
    BorderActive:  lipgloss.Color("#b4befe"),
    Selection:     lipgloss.Color("#313244"),
    GradientStart: lipgloss.Color("#89b4fa"),
    GradientEnd:   lipgloss.Color("#cba6f7"),
}
```

## Component Styles

### Panel Style

```go
// Inactive panel
var PanelStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(Theme.Border).
    Padding(0, 1)

// Active panel (focused)
var PanelActiveStyle = PanelStyle.Copy().
    BorderForeground(Theme.BorderActive)
```

### Title Style

Panel titles with subtle gradient effect:

```go
var TitleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(Theme.Primary).
    Padding(0, 1)
```

### List Item Styles

```go
// Normal item
var ItemStyle = lipgloss.NewStyle().
    Foreground(Theme.Foreground)

// Selected item (cursor on it)
var ItemSelectedStyle = lipgloss.NewStyle().
    Foreground(Theme.Primary).
    Background(Theme.Selection).
    Bold(true)

// Item description (secondary text)
var ItemDescStyle = lipgloss.NewStyle().
    Foreground(Theme.Subtle)
```

### Footer Style

```go
var FooterStyle = lipgloss.NewStyle().
    Foreground(Theme.Subtle).
    Border(lipgloss.NormalBorder(), true, false, false, false).
    BorderForeground(Theme.Border).
    Padding(0, 1)

// Key hint in footer
var KeyStyle = lipgloss.NewStyle().
    Foreground(Theme.Primary).
    Bold(true)

// Action text in footer
var ActionStyle = lipgloss.NewStyle().
    Foreground(Theme.Foreground)
```

## Icons (Nerd Fonts)

Use Nerd Font icons for visual polish:

```go
var Icons = struct {
    Script      string  // 
    Category    string  // 󰉋
    Folder      string  // 󰉋
    Pin         string  // 󰐕
    Search      string  // 
    Running     string  // 
    Success     string  // 
    Error       string  // 
    Arrow       string  // 
    Bash        string  // 
    Python      string  // 
    Zsh         string  // 
}{
    Script:   "",
    Category: "󰉋",
    Folder:   "󰉋",
    Pin:      "󰐕",
    Search:   "",
    Running:  "",
    Success:  "",
    Error:    "",
    Arrow:    "",
    Bash:     "",
    Python:   "",
    Zsh:      "",
}
```

## Panel Components

### Sidebar Panel

```
┌─ Categories ─────────┐
│                      │
│  󰉋  All Scripts (12) │  ← Total count
│                      │
│  󰒍  Deploy (4)       │  ← Category with count
│  󰊤  Database (3)     │
│  󰑮  Testing (2)      │
│  󰒲  Utils (3)        │
│                      │
│ ──────────────────── │  ← Separator
│  󰐕  Pinned           │  ← Pinned section
│                      │
│     quick-deploy     │
│     backup-prod      │
│                      │
└──────────────────────┘
```

Features:
- Category icon + name + script count
- Active category highlighted
- Separator between categories and pinned
- Pinned scripts for quick access

### Scripts Panel

```
┌─ Scripts ─────────────────────────────────────┐
│                                               │
│  Filter: deploy                     [3/12]   │  ← Filter bar (when active)
│                                               │
│   deploy-staging.sh                          │
│   Deploy to staging environment               │
│                                               │
│ ● backup-database.sh                ← selected│
│   Backup PostgreSQL database                  │
│                                               │
│   run-tests.sh                               │
│   Execute test suite                          │
│                                               │
└───────────────────────────────────────────────┘
```

Features:
- Script name (bold when selected)
- Description on second line (dimmed)
- Selection indicator (● or highlight)
- Filter input overlays when typing `/`
- Match count during filtering

### Details Panel

```
┌─ Details ────────────────────────────┐
│                                      │
│   backup-database.sh                │
│                                      │
│  Backup PostgreSQL database to S3    │
│  with compression and encryption.    │
│                                      │
│  ─────────────────────────────────── │
│                                      │
│  Category    Database                │
│  Shell       bash                    │
│  Path        ~/scripts/db/           │
│                                      │
│  ─────────────────────────────────── │
│                                      │
│  Parameters                          │
│                                      │
│  • database (required)               │
│    Database name to backup           │
│                                      │
│  • compress (default: true)          │
│    Enable gzip compression           │
│                                      │
└──────────────────────────────────────┘
```

Features:
- Script name with icon
- Full description (wrapped)
- Metadata section
- Parameters with types and defaults

### Footer Bar

```
┌───────────────────────────────────────────────────────────────────────────────┐
│  ↑↓ navigate  enter run  / filter  tab panel  p params  ? help  q quit       │
└───────────────────────────────────────────────────────────────────────────────┘
```

Features:
- Context-aware hints (changes based on state)
- Key highlighted, action dimmed
- Compact single-line format

## Interaction States

### Focus States

| State | Visual Treatment |
|-------|------------------|
| Panel focused | Bright border (#b4befe) |
| Panel unfocused | Dim border (#6c7086) |
| Item selected | Background highlight + bold |
| Item hovered | Subtle background |

### Filter Overlay

When user presses `/`:

```
┌─ Scripts ─────────────────────────────────────┐
│ ┌─ Filter ──────────────────────────────────┐ │
│ │   deploy█                                │ │
│ └───────────────────────────────────────────┘ │
│                                               │
│   deploy-staging.sh                 [match]  │
│   deploy-production.sh              [match]  │
│   backup-database.sh                         │  ← dimmed non-matches
│                                               │
└───────────────────────────────────────────────┘
```

### Loading State

```
┌─ Scripts ─────────────────────────────────────┐
│                                               │
│              Scanning scripts...              │
│                    ⠋                          │
│                                               │
└───────────────────────────────────────────────┘
```

## Keyboard Navigation

### Global Keys

| Key | Action |
|-----|--------|
| `tab` | Switch between panels |
| `shift+tab` | Switch panels (reverse) |
| `q` / `ctrl+c` | Quit |
| `?` | Toggle help overlay |
| `/` | Open filter input |
| `esc` | Close overlay / cancel |

### List Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `home` / `g` | Go to top |
| `end` / `G` | Go to bottom |
| `pgup` / `ctrl+u` | Page up |
| `pgdn` / `ctrl+d` | Page down |

### Actions

| Key | Action |
|-----|--------|
| `enter` | Run selected script |
| `p` | Edit parameters (if any) |
| `e` | Edit script in $EDITOR |
| `y` | Copy script path |
| `space` | Toggle pin |

## Responsive Behavior

### Width Breakpoints

| Width | Layout |
|-------|--------|
| ≥120 cols | 3 panels (sidebar + main + details) |
| 80-119 cols | 2 panels (sidebar + main) |
| 60-79 cols | 2 panels (compact) |
| <60 cols | 1 panel (main only) |

### Height Handling

- Minimum usable height: 10 rows
- Footer always visible
- Lists scroll when content exceeds panel height
- Details panel content truncates with "..." indicator

## Animation & Polish

### Subtle Animations (Optional)

- Cursor movement: instant (no delay)
- Panel focus: instant border color change
- Filter results: immediate update (debounced 50ms)
- Loading spinner: smooth rotation

### Visual Feedback

- Selection change: immediate highlight
- Action success: brief green flash or checkmark
- Error: red flash + error message in footer

## Implementation Notes

### File Structure

```
internal/tui/
├── app.go           # Main Bubble Tea model
├── styles.go        # All lipgloss styles
├── theme.go         # Color theme definitions
├── icons.go         # Nerd font icons
├── keys.go          # Key bindings
├── sidebar.go       # Sidebar panel component
├── scripts.go       # Scripts list component
├── details.go       # Details panel component
├── footer.go        # Footer component
├── filter.go        # Filter overlay component
└── help.go          # Help overlay component
```

### Bubble Tea Model Structure

```go
type AppModel struct {
    // State
    state       AppState
    activePanel Panel
    
    // Panels
    sidebar  SidebarModel
    scripts  ScriptsModel
    details  DetailsModel
    
    // Overlays
    filter   FilterModel
    help     HelpModel
    
    // Data
    categories []Category
    scripts    []Script
    selected   *Script
    
    // Layout
    width  int
    height int
}

type AppState int

const (
    StateLoading AppState = iota
    StateBrowsing
    StateFiltering
    StateHelp
    StateParams
)

type Panel int

const (
    PanelSidebar Panel = iota
    PanelScripts
    PanelDetails
)
```

### Style Application Pattern

```go
func (m ScriptsModel) View() string {
    // Determine panel style based on focus
    style := styles.Panel
    if m.focused {
        style = styles.PanelActive
    }
    
    // Build content
    var items []string
    for i, script := range m.scripts {
        itemStyle := styles.Item
        if i == m.cursor {
            itemStyle = styles.ItemSelected
        }
        
        name := itemStyle.Render(script.Name)
        desc := styles.ItemDesc.Render(script.Description)
        items = append(items, name + "\n" + desc)
    }
    
    content := strings.Join(items, "\n\n")
    
    // Apply panel style with title
    return style.
        Width(m.width).
        Height(m.height).
        Render(styles.Title.Render("Scripts") + "\n\n" + content)
}
```

## Theme Customization (Future)

Support loading custom themes from config:

```yaml
# ~/.config/tap/theme.yaml
theme:
  background: "#1e1e2e"
  foreground: "#cdd6f4"
  primary: "#89b4fa"
  secondary: "#cba6f7"
  # ... etc
```

Or use preset names:

```yaml
theme: catppuccin-mocha  # default
# theme: catppuccin-latte
# theme: dracula
# theme: nord
```
