# tap Implementation Plan

> Last updated: Initial seed
> Status: Phase 1 - Foundation

## Phase 1: Foundation (MVP)

### Completed
_(none yet)_

### Current Focus
- [ ] Initialize Go module and project structure ← CURRENT
  - Create go.mod with module path
  - Create directory structure: cmd/tap/, internal/core/, internal/tui/, internal/config/, internal/cli/
  - Add .gitignore for Go projects

### Core Models
- [ ] Create script and parameter data models
  - internal/core/script.go: Script, Parameter, Example structs
  - internal/core/category.go: Category struct
  - Spec: specs/01-scanner.md#data-structures

### Metadata Parser  
- [ ] Implement YAML front matter parser
  - internal/core/parser.go: ExtractMetadata function
  - Handle bash scripts (# --- delimiters)
  - Handle python scripts (docstring with ---)
  - Spec: specs/01-scanner.md#metadata-extraction

- [ ] Add parser tests
  - internal/core/parser_test.go
  - testdata/valid_bash.sh
  - testdata/valid_python.py  
  - testdata/minimal.sh
  - testdata/no_metadata.sh
  - testdata/invalid_yaml.sh

### Directory Scanner
- [ ] Implement directory scanner
  - internal/core/scanner.go: Scanner interface, ScanDirectory function
  - Use filepath.WalkDir for efficiency
  - Filter by extensions (.sh, .bash, .py)
  - Skip ignored directories (.git, node_modules, etc.)
  - Spec: specs/01-scanner.md#directory-scanning

- [ ] Add scanner tests
  - internal/core/scanner_test.go
  - testdata/scripts/ directory with test scripts

### Configuration
- [ ] Implement config manager
  - internal/config/config.go: Config struct, Load/Save functions
  - XDG path handling with adrg/xdg
  - Default config creation on first run
  - Spec: specs/02-config.md

- [ ] Add config tests
  - internal/config/config_test.go

### TUI Menu
- [ ] Implement base TUI app model
  - internal/tui/app.go: AppModel, state machine
  - internal/tui/styles.go: lipgloss styles
  - internal/tui/keys.go: key bindings
  - Spec: specs/03-tui-menu.md

- [ ] Implement category list view
  - internal/tui/menu.go: MenuModel with bubbles/list
  - Category navigation with enter to drill down
  - Back navigation with esc/backspace

- [ ] Implement script list view
  - Script list within selected category
  - Custom delegate for script rendering
  - Enter to select script

- [ ] Implement filter overlay
  - Textinput for fuzzy filtering
  - / to activate, esc to cancel
  - Filter both categories and scripts

### Executor
- [ ] Implement script executor
  - internal/core/executor.go: Executor interface, Execute function
  - Real-time stdout/stderr streaming
  - Exit code pass-through
  - Spec: specs/05-executor.md

- [ ] Add executor tests
  - internal/core/executor_test.go
  - testdata/echo.sh, testdata/exit_code.sh

### CLI Wiring
- [ ] Implement root command
  - internal/cli/root.go: Cobra root command
  - Mode detection (interactive vs headless)
  - cmd/tap/main.go: entry point
  - Spec: specs/06-cli.md

- [ ] Implement list command
  - internal/cli/list.go: tap list
  - Category filtering with --category
  - JSON output with --json

### Integration
- [ ] Wire everything together
  - TUI launches from root command
  - Script selection triggers executor
  - Exit TUI before execution (tea.ExitAltScreen)
  - Pass through exit code

- [ ] End-to-end testing
  - Manual test: tap → navigate → select → execute
  - Verify real script execution works

---

## Phase 2: Parameters & Headless
_(not started - complete Phase 1 first)_

- [ ] Parameter parsing in metadata
- [ ] TUI parameter form (huh library)
- [ ] `tap run <script>` command
- [ ] --param flag handling
- [ ] Parameter validation

---

## Phase 3: Script Management  
_(not started - complete Phase 2 first)_

- [ ] `tap new` interactive scaffolding
- [ ] `tap add <path>` registration
- [ ] `tap remove <script>` unregistration
- [ ] Script templates (bash.tmpl, python.tmpl)

---

## Discovered Issues
_(add bugs/issues found during implementation)_

---

## Notes
- Each task should be completable in one Claude Code session
- Don't skip ahead to future phases
- If a task is too big, split it before implementing
