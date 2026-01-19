# CLAUDE.md

> Instructions for implementing tap with Claude Code

## Project Overview

**tap** is a terminal-based script runner that gives developers quick access to their shell and Python scripts through an interactive TUI menu or direct CLI invocation.

## Key Design Decisions

1. **Bubble Tea for TUI** — Use the charmbracelet ecosystem (bubbletea, bubbles, huh, lipgloss)
2. **Cobra for CLI** — Standard Go CLI framework with flag parsing and help generation
3. **YAML front matter** — Script metadata embedded in comments, no separate config files
4. **XDG compliance** — Config in `~/.config/tap/`, data in `~/.local/share/tap/`
5. **Real-time execution** — Exit TUI before script runs, pass through exit codes

## Spec Files

Read these specs in order before implementing each phase:

| Phase | Primary Specs |
|-------|---------------|
| Phase 1 | [01-scanner.md](specs/01-scanner.md), [02-config.md](specs/02-config.md), [03-tui-menu.md](specs/03-tui-menu.md), [05-executor.md](specs/05-executor.md) |
| Phase 2 | [04-tui-params.md](specs/04-tui-params.md), [06-cli.md](specs/06-cli.md) |
| Phase 3 | [07-scaffolding.md](specs/07-scaffolding.md) |

## Implementation Order

### Phase 1: Foundation (MVP)

**Goal:** Browse and run scripts with filtering

```
1. Create project structure
   - Initialize Go module: `go mod init github.com/user/tap`
   - Create directory structure per SPEC.md
   
2. Implement core/parser.go
   - YAML front matter extraction from bash/python scripts
   - See specs/01-scanner.md for metadata schema
   
3. Implement core/scanner.go
   - Directory scanning with filepath.WalkDir
   - Extension filtering (.sh, .bash, .py)
   - Category organization
   
4. Implement config/config.go
   - Viper-based configuration loading
   - XDG path handling with adrg/xdg
   - Default config creation
   
5. Implement tui/menu.go
   - Category list view with bubbles/list
   - Script list view with drill-down
   - Fuzzy filtering with textinput
   
6. Implement core/executor.go
   - Script execution with os/exec
   - Real-time output streaming
   - Exit code pass-through
   
7. Implement cli/root.go
   - Basic `tap` command that launches TUI
   - Mode detection (interactive vs headless)
   
8. Implement cli/list.go
   - `tap list` command
```

**Test after Phase 1:**
```bash
# Setup
mkdir -p ~/scripts
cat > ~/scripts/hello.sh << 'EOF'
#!/bin/bash
# ---
# name: hello
# description: Say hello
# category: demo
# ---
echo "Hello, world!"
EOF
chmod +x ~/scripts/hello.sh

# Configure
tap config add-dir ~/scripts

# Test
tap           # Should show TUI with hello script
tap list      # Should list hello script
```

### Phase 2: Parameters & Headless

**Goal:** Handle script parameters in both modes

```
1. Extend core/parser.go
   - Parameter parsing from metadata
   - Validation rules
   
2. Implement tui/form.go
   - Parameter input form using huh
   - Type-specific form fields
   - Validation display
   
3. Implement cli/run.go
   - `tap run <script>` command
   - --param flag parsing
   - Inline param=value syntax
   
4. Update core/executor.go
   - Pass parameters as TAP_PARAM_* env vars
```

**Test after Phase 2:**
```bash
cat > ~/scripts/greet.sh << 'EOF'
#!/bin/bash
# ---
# name: greet
# description: Greet someone
# category: demo
# parameters:
#   - name: name
#     type: string
#     required: true
#     description: Name to greet
#   - name: loud
#     type: bool
#     default: false
# ---
msg="Hello, $TAP_PARAM_NAME!"
if [ "$TAP_PARAM_LOUD" = "true" ]; then
    echo "${msg^^}"
else
    echo "$msg"
fi
EOF

# Test interactive params
tap run greet

# Test headless
tap run greet --param name=World --param loud=true
tap run greet name=Claude
```

### Phase 3: Script Management

**Goal:** Create and register scripts easily

```
1. Implement templates/
   - Embed bash.tmpl and python.tmpl
   
2. Implement cli/new.go
   - Interactive script creation
   - Headless creation with flags
   
3. Implement cli/add.go
   - Script registration
   - Directory scanning with --recursive
   
4. Implement cli/remove.go
   - Script unregistration
```

## Code Style Guidelines

### Go Conventions

```go
// Use meaningful names
type Scanner interface { ... }  // Not: type S interface { ... }

// Return errors, don't panic
func (s *scanner) ParseScript(path string) (*Script, error) {
    // ...
    if err != nil {
        return nil, fmt.Errorf("parsing %s: %w", path, err)
    }
}

// Use context for cancellation
func (s *scanner) Scan(ctx context.Context) ([]Script, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // ...
}
```

### Bubble Tea Patterns

```go
// Keep Update() fast - use Cmd for slow operations
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case loadScriptsMsg:
        // Fast: just update state
        m.scripts = msg.scripts
        return m, nil
    case tea.KeyMsg:
        if msg.String() == "r" {
            // Slow: return Cmd
            return m, loadScriptsCmd
        }
    }
    return m, nil
}

// Define clear message types
type loadScriptsMsg struct {
    scripts []Script
}

type errorMsg struct {
    err error
}
```

### Error Messages

```go
// User-friendly errors
fmt.Errorf("script not found: %s", name)
fmt.Errorf("no scan directories configured. Run: tap config add-dir <path>")
fmt.Errorf("invalid parameter %q: expected %s", name, expectedType)

// Don't expose internal details
// Bad: fmt.Errorf("yaml.Unmarshal failed: %w", err)
// Good: fmt.Errorf("invalid metadata in %s: %w", path, err)
```

## Testing Strategy

### Unit Tests

Every package should have `*_test.go` files:

```go
// core/parser_test.go
func TestParseScript_ValidBash(t *testing.T) {
    script, err := ParseScript("testdata/valid.sh")
    require.NoError(t, err)
    assert.Equal(t, "deploy", script.Name)
    assert.Equal(t, "deployment", script.Category)
}
```

### Test Fixtures

Create `testdata/` directories with sample scripts:

```
testdata/
├── valid.sh           # Full metadata
├── minimal.sh         # Minimal metadata
├── no_metadata.sh     # No metadata (should be skipped)
├── invalid_yaml.sh    # Invalid YAML (should error)
└── params.sh          # With parameters
```

### Integration Tests

Test full workflows:

```go
func TestTUI_SelectAndRun(t *testing.T) {
    // Use teatest for TUI testing
}

func TestCLI_RunWithParams(t *testing.T) {
    // Test CLI end-to-end
}
```

## Dependencies

Add to go.mod:

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/huh v0.3.0
    github.com/charmbracelet/lipgloss v0.10.0
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    github.com/adrg/xdg v0.4.0
    github.com/mattn/go-isatty v0.0.20
    gopkg.in/yaml.v3 v3.0.1
)
```

## Common Pitfalls

### Bubble Tea

1. **Don't block in Update()** — Use Cmd for I/O
2. **Handle WindowSizeMsg** — Always track terminal size
3. **Exit before script execution** — Use tea.ExitAltScreen

### Metadata Parsing

1. **Check for shebang first** — Skip line 1 if it starts with `#!`
2. **Handle Python docstrings** — Different comment style
3. **Validate early** — Reject bad metadata at scan time

### Execution

1. **Preserve working directory** — Run scripts from their directory
2. **Pass through signals** — Forward SIGINT/SIGTERM to child
3. **Exit with script's code** — Don't swallow exit codes

## Quick Reference

### Commands to Implement

| Command | Phase | Spec |
|---------|-------|------|
| `tap` | 1 | 03-tui-menu.md |
| `tap list` | 1 | 06-cli.md |
| `tap run <script>` | 2 | 06-cli.md |
| `tap new` | 3 | 07-scaffolding.md |
| `tap add <path>` | 3 | 06-cli.md |
| `tap remove <script>` | 3 | 06-cli.md |
| `tap config` | 1 | 02-config.md |

### Key Files

| File | Purpose |
|------|---------|
| `cmd/tap/main.go` | Entry point |
| `internal/core/script.go` | Script model |
| `internal/core/parser.go` | Metadata extraction |
| `internal/core/scanner.go` | Directory scanning |
| `internal/core/executor.go` | Script execution |
| `internal/config/config.go` | Configuration management |
| `internal/tui/app.go` | Root TUI model |
| `internal/tui/menu.go` | Menu navigation |
| `internal/tui/form.go` | Parameter forms |
| `internal/cli/root.go` | Cobra root command |

## Debugging Tips

### TUI Issues

```go
// Add debug logging
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    log.Printf("Update: %T %+v", msg, msg)
    // ...
}
```

### Script Execution

```bash
# Test script directly first
TAP_PARAM_NAME=test ./script.sh

# Check environment
env | grep TAP_
```

### Config Issues

```bash
# Show config location
tap config show

# Check XDG paths
echo $XDG_CONFIG_HOME
echo $XDG_DATA_HOME
```
