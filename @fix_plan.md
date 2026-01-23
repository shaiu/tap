# tap Implementation Plan

> Last updated: 2026-01-19
> Status: Phase 1 - Foundation

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
_(in progress)_

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
- [ ] Integration tests for run command ← CURRENT

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
