# tap Implementation Plan

> Last updated: 2026-01-23
> Status: Phase 3 - Complete

## Phase 1: Foundation (MVP)

### Completed
- [x] Initialize Go module and project structure
  - Created go.mod with module path github.com/shaiungar/tap
  - Created directory structure: cmd/tap/, internal/core/, internal/tui/, internal/config/, internal/cli/, internal/templates/
  - Added .gitignore for Go projects
  - Added minimal main.go entry point

### Current Focus

### Core Models
- [x] Create script and parameter data models
  - internal/core/script.go: Script, Parameter, Example structs
  - internal/core/category.go: Category struct
  - Spec: specs/01-scanner.md#data-structures

### Metadata Parser
- [x] Implement YAML front matter parser
  - internal/core/parser.go: ParseScript function
  - Handle bash scripts (# --- delimiters)
  - Handle python scripts (docstring with ---)
  - Spec: specs/01-scanner.md#metadata-extraction

- [x] Add parser tests
  - internal/core/parser_test.go
  - testdata/valid_bash.sh
  - testdata/valid_python.py
  - testdata/minimal.sh
  - testdata/no_metadata.sh
  - testdata/invalid_yaml.sh

### Directory Scanner
- [x] Implement directory scanner
  - internal/core/scanner.go: Scanner interface, ScanDirectory function
  - Use filepath.WalkDir for efficiency
  - Filter by extensions (.sh, .bash, .py)
  - Skip ignored directories (.git, node_modules, etc.)
  - Spec: specs/01-scanner.md#directory-scanning

- [x] Add scanner tests
  - internal/core/scanner_test.go
  - testdata/scripts/ directory with test scripts

### Configuration
- [x] Implement config manager
  - internal/config/config.go: Config struct, Load/Save functions
  - XDG path handling with adrg/xdg
  - Default config creation on first run
  - Spec: specs/02-config.md

- [x] Add config tests
  - internal/config/config_test.go

### TUI Menu
- [x] Implement base TUI app model
  - internal/tui/app.go: AppModel, state machine
  - internal/tui/styles.go: lipgloss styles
  - internal/tui/keys.go: key bindings
  - Spec: specs/03-tui-menu.md

- [x] Implement category list view
  - internal/tui/menu.go: MenuModel with bubbles/list
  - Category navigation with enter to drill down
  - Back navigation with esc/backspace

- [x] Implement script list view
  - Script list within selected category
  - Custom delegate for script rendering
  - Enter to select script

- [x] Implement filter overlay
  - Textinput for fuzzy filtering
  - / to activate, esc to cancel
  - Filter both categories and scripts

### Executor
- [x] Implement script executor
  - internal/core/executor.go: Executor interface, Execute function
  - Real-time stdout/stderr streaming
  - Exit code pass-through
  - Spec: specs/05-executor.md

- [x] Add executor tests
  - internal/core/executor_test.go
  - testdata/echo.sh, testdata/exit_code.sh, testdata/stderr.sh, testdata/workdir.sh

### CLI Wiring
- [x] Implement root command
  - internal/cli/root.go: Cobra root command
  - Mode detection (interactive vs headless)
  - cmd/tap/main.go: entry point
  - Spec: specs/06-cli.md

- [x] Implement list command
  - internal/cli/list.go: tap list
  - Category filtering with --category
  - JSON output with --json

### Integration
- [x] Wire everything together
  - TUI launches from root command
  - Script selection triggers executor
  - Exit TUI before execution (tea.ExitAltScreen)
  - Pass through exit code

- [x] End-to-end testing
  - Manual test: tap → navigate → select → execute
  - Verify real script execution works
  - Added testdata/e2e/ with test scripts (hello.sh, greet.sh, show-env.sh, exit-code.sh)
  - Added internal/cli/integration_test.go with comprehensive E2E tests

---

## Phase 2: Parameters & Headless
_(completed)_

- [x] Parameter parsing in metadata ← DONE
  - Added validateParameters() function to parser.go
  - Validates: name identifiers, type values, short flag length, duplicate names/shorts, default in choices
- [x] TUI parameter form (huh library) ← DONE
  - Added internal/tui/form.go with FormModel using huh library
  - Supports string, int, float, bool params and choices
  - Integrated StateForm into app.go with form messages
  - Added Required style and comprehensive tests
- [x] `tap run <script>` command ← DONE
  - Added internal/cli/run.go with full run command
  - Supports --param/-p flags and inline key=value params
  - Finds scripts by name or category/name prefix
  - Validates parameter types and choices
  - Applies defaults for missing optional params
  - Shows interactive form when required params missing
  - Added comprehensive tests in run_test.go
- [x] --param flag handling ← (included in tap run)
- [x] Parameter validation ← (included in tap run)
- [x] Integration tests for run command ← DONE
  - Added testdata/e2e/demo/with-params.sh script with 4 parameters
  - Added 12 integration tests covering:
    - Basic run command flow (find script, execute)
    - Run with params (inline and flag)
    - Param merge priority (inline > flag)
    - Missing required param validation
    - Invalid param type validation
    - Invalid choice validation
    - Default value application
    - Interactive param detection
    - Full execution with all params
    - Script not found error

---

## Phase 3: Script Management
_(in progress)_

- [x] `tap new` interactive scaffolding ← DONE
  - Added internal/cli/new.go with NewScriptConfig, ParameterConfig, TemplateData structs
  - Interactive mode using huh forms for name, description, category, shell, parameters
  - Headless mode with flags: -d description, -c category, -o output, --shell, -p params
  - Parameter wizard for interactive parameter creation
  - Template-based script generation with embedded templates
  - Created internal/templates/templates.go with embedded bash.tmpl and python.tmpl
  - Comprehensive tests in new_test.go (validation, parsing, generation)
- [x] `tap add <path>` registration ← DONE
  - Added internal/cli/add.go with addCmd, addSingleScript, addDirectory functions
  - Supports --alias/-a for script aliases
  - Supports --recursive/-r for adding all scripts in a directory
  - Validates scripts have proper YAML metadata before registering
  - Added comprehensive tests in add_test.go
- [x] `tap remove <script>` unregistration ← DONE
  - Added internal/cli/remove.go with removeCmd
  - Supports removal by path (absolute or relative) or alias
  - Added comprehensive tests in remove_test.go
- [x] Script templates (bash.tmpl, python.tmpl) ← (included in tap new)

## Phase 4: Superfile-Inspired UI Refactor 🎨

> Reference: specs/08-ui-design.md

### Theme & Colors
- [x] Create theme system with Catppuccin Mocha palette
  - internal/tui/theme.go: Color definitions
  - Background #1e1e2e, Primary #89b4fa, etc.
  - Support for future theme switching

- [x] Refactor styles to use theme colors ← DONE
  - internal/tui/styles.go: Updated all lipgloss styles to use Theme colors
  - Added Panel/PanelActive styles with rounded borders
  - Active border (#b4befe) vs inactive (#6c7086)
  - Added Item/ItemSelected/ItemDesc for list styling
  - Added Key/Action for footer hints
  - Added styles_test.go with comprehensive tests

### Icons & Visual Polish
- [x] Add Nerd Font icons ← DONE
  - internal/tui/icons.go: IconSet struct, NerdFontIcons, ASCIIIcons
  - 󰉋 categories,  scripts,  bash,  python
  - Graceful fallback via UseASCIIIcons() function
  - IconForShell() helper for shell-specific icons
  - Comprehensive tests in icons_test.go

- [x] Update list item rendering ← DONE
  - 2-line format: name (bold) + description (dimmed)
  - Icon prefix based on script type (IconForShell)
  - Selection indicator (●) for both categories and scripts

### 3-Panel Layout
- [x] Implement responsive 3-panel layout ← DONE
  - Created Panel type with PanelSidebar, PanelScripts, PanelDetails
  - Created SidebarModel, ScriptsModel, DetailsModel components
  - Tab/Shift+Tab to switch between panels
  - Active panel border highlighting with Catppuccin theme
  - Responsive breakpoints: ≥120 cols (3 panels), 80-119 (2 panels), <80 (1 panel)
  - Updated AppModel state machine to use StateBrowsing for 3-panel mode
  - Updated all tests to use new StateBrowsing state

- [x] Refactor sidebar panel ← DONE
  - Category list with icons and counts
  - "All Scripts" option at top
  - Pinned scripts section with separator
  - Added SidebarItemType enum for item classification
  - Added pinned scripts support with visual separator
  - Navigation skips pinned header (non-selectable)
  - Added sidebar_test.go with comprehensive tests
  - Spec: specs/08-ui-design.md#sidebar-panel

- [x] Refactor scripts panel ← DONE
  - 2-line item display (name + description)
  - Improved selection styling
  - Filter bar integration
  - Match count during filtering
  - Added FilterQuery and FilterCount styles
  - Added scripts_test.go with comprehensive tests
  - Spec: specs/08-ui-design.md#scripts-panel

- [x] Implement details panel (NEW) ← DONE
  - Script name with icon
  - Full description (text wrapped)
  - Metadata: category, shell, path
  - Parameters list with types/defaults
  - Tags section
  - Comprehensive tests in details_test.go
  - Spec: specs/08-ui-design.md#details-panel

### Footer & Help
- [x] Refactor footer bar ← DONE
  - Single-line format
  - Key (highlighted) + action (dimmed) pairs
  - Context-aware hints based on state
  - Created internal/tui/footer.go with FooterModel, FooterContext, KeyHint
  - Footer shows different hints based on state (Browsing, Filter, Help, Form)
  - Context-aware actions (select vs run based on panel)
  - Shows params hint when selected script has parameters
  - Added comprehensive tests in footer_test.go
  - Spec: specs/08-ui-design.md#footer-bar

- [ ] Improve help overlay ← CURRENT
  - Categorized keybinding sections
  - Styled with theme colors
  - ? to toggle

### Filter Improvements
- [ ] Enhance filter overlay
  - Overlay box styling (superfile-style)
  - Real-time match highlighting
  - Non-matching items dimmed (not hidden)
  - Match count [3/12] display

### Responsive Design
- [ ] Implement width breakpoints
  - ≥120 cols: 3 panels
  - 80-119 cols: 2 panels (no details)
  - <80 cols: 1 panel (scripts only)
  - Handle tea.WindowSizeMsg properly

### Final Polish
- [ ] Add loading state with spinner
  - "Scanning scripts..." message
  - Centered spinner animation

- [ ] Visual feedback for actions
  - Brief highlight on script run
  - Error display in footer

- [ ] Test on various terminal sizes
  - Verify responsive breakpoints
  - Check color rendering in different terminals

---

## Discovered Issues
_(add bugs/issues found during implementation)_

---

## Notes
- Each task should be completable in one Claude Code session
- Don't skip ahead to future phases
- If a task is too big, split it before implementing
