# tap

> Just tap and go — your personal script runner

**tap** is a terminal-based script runner that gives developers quick access to their shell and Python scripts through an interactive TUI menu or direct CLI invocation.

## Vision

Developers accumulate scripts for repeated tasks: deployments, data migrations, environment setup, debugging routines. These scripts end up scattered across directories, forgotten, or requiring manual lookup. **tap** solves this by:

1. **Discovering** scripts automatically from configured directories
2. **Organizing** them into browsable categories
3. **Documenting** them via embedded metadata (no separate config files)
4. **Running** them with interactive parameter prompts or headless flags

## Core Principles

- **Zero friction** — Scripts work as-is; just add a metadata header
- **Embedded metadata** — Documentation lives in the script, not alongside it
- **Hybrid discovery** — Auto-scan directories + explicit registration
- **Dual mode** — Interactive TUI for exploration, headless for automation
- **Real-time output** — Scripts stream output directly, then tap exits

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer (Cobra)                        │
│  tap / tap run <script> / tap list / tap new / tap add          │
└─────────────────────────────────────────────────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
         ┌──────────────────┐    ┌──────────────────┐
         │   TUI (Bubble    │    │  Headless Mode   │
         │      Tea)        │    │  (Direct Exec)   │
         └──────────────────┘    └──────────────────┘
                    │                       │
                    └───────────┬───────────┘
                                ▼
         ┌──────────────────────────────────────────┐
         │              Core Services               │
         │  ┌────────────┐  ┌────────────────────┐  │
         │  │  Scanner   │  │  Metadata Parser   │  │
         │  └────────────┘  └────────────────────┘  │
         │  ┌────────────┐  ┌────────────────────┐  │
         │  │  Executor  │  │  Config Manager    │  │
         │  └────────────┘  └────────────────────┘  │
         └──────────────────────────────────────────┘
                                │
                                ▼
         ┌──────────────────────────────────────────┐
         │              File System                 │
         │  ~/.config/tap/    ~/.local/share/tap/   │
         │  ~/scripts/        (user directories)    │
         └──────────────────────────────────────────┘
```

## Key Components

| Component | Spec | Description |
|-----------|------|-------------|
| Scanner | [specs/01-scanner.md](specs/01-scanner.md) | Script discovery and directory scanning |
| Config | [specs/02-config.md](specs/02-config.md) | Configuration and state management |
| TUI Menu | [specs/03-tui-menu.md](specs/03-tui-menu.md) | Interactive menu navigation with filtering |
| TUI Params | [specs/04-tui-params.md](specs/04-tui-params.md) | Parameter input forms |
| Executor | [specs/05-executor.md](specs/05-executor.md) | Script execution engine |
| CLI | [specs/06-cli.md](specs/06-cli.md) | Cobra commands and headless mode |
| Scaffolding | [specs/07-scaffolding.md](specs/07-scaffolding.md) | Script generation (`tap new`) |

## Implementation Phases

### Phase 1: Foundation (MVP)
**Goal:** Browse and run scripts with filtering

- [ ] Metadata parser (YAML front matter in bash/python)
- [ ] Directory scanner with extension filtering
- [ ] Basic configuration (scan directories)
- [ ] TUI: Category list → Script list with fuzzy filter
- [ ] Executor: Run script with real-time streaming output
- [ ] CLI: `tap` (interactive), `tap list`

**Deliverable:** User can configure directories, browse discovered scripts by category, filter by name, and run them.

### Phase 2: Parameters & Headless
**Goal:** Handle script parameters in both modes

- [ ] Parameter parsing from metadata
- [ ] TUI: Parameter input form (text, select, bool)
- [ ] Headless execution: `tap run <script> --param key=value`
- [ ] CLI: `tap run` with flag-based parameters
- [ ] Parameter validation (required, choices, types)

**Deliverable:** Scripts with parameters work in both interactive and headless modes.

### Phase 3: Script Management
**Goal:** Create and register scripts easily

- [ ] `tap new` — Interactive script scaffolding
- [ ] `tap add <path>` — Register external scripts
- [ ] `tap remove <script>` — Unregister scripts
- [ ] Template system for new scripts
- [ ] Metadata cache for faster startup

**Deliverable:** Full script lifecycle management.

### Phase 4: Polish & Power Features
**Goal:** Quality-of-life improvements

- [ ] Favorites / recently used
- [ ] Script aliases
- [ ] Execution history
- [ ] `tap edit <script>` — Open in $EDITOR
- [ ] Shell completions (bash, zsh, fish)
- [ ] `tap config` — Interactive configuration

**Deliverable:** Production-ready tool with great UX.

## Project Structure

```
tap/
├── cmd/tap/
│   └── main.go                 # Entry point
├── internal/
│   ├── cli/                    # Cobra commands
│   │   ├── root.go
│   │   ├── run.go
│   │   ├── list.go
│   │   ├── new.go
│   │   └── add.go
│   ├── tui/                    # Bubble Tea components
│   │   ├── app.go              # Root model & state machine
│   │   ├── menu.go             # Category/script list
│   │   ├── filter.go           # Fuzzy filter overlay
│   │   ├── form.go             # Parameter input
│   │   ├── styles.go           # Lipgloss styling
│   │   └── keys.go             # Key bindings
│   ├── core/                   # Business logic
│   │   ├── scanner.go          # Directory scanning
│   │   ├── parser.go           # Metadata extraction
│   │   ├── executor.go         # Script execution
│   │   ├── script.go           # Script model
│   │   └── category.go         # Category model
│   ├── config/                 # Configuration
│   │   └── config.go
│   └── templates/              # Script templates
│       ├── embed.go
│       └── templates/
│           └── bash.tmpl
├── SPEC.md                     # This file
├── CLAUDE.md                   # Claude Code instructions
├── specs/                      # Detailed component specs
│   ├── 01-scanner.md
│   ├── 02-config.md
│   ├── 03-tui-menu.md
│   ├── 04-tui-params.md
│   ├── 05-executor.md
│   ├── 06-cli.md
│   └── 07-scaffolding.md
├── go.mod
└── go.sum
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| TUI Components | [Bubbles](https://github.com/charmbracelet/bubbles) (list, textinput) |
| TUI Forms | [Huh](https://github.com/charmbracelet/huh) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Config | [Viper](https://github.com/spf13/viper) |
| YAML Parsing | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| TTY Detection | [go-isatty](https://github.com/mattn/go-isatty) |
| XDG Paths | [adrg/xdg](https://github.com/adrg/xdg) |

## Metadata Format

Scripts embed their metadata as YAML front matter in comments:

```bash
#!/bin/bash
# ---
# name: deploy
# description: Deploy application to environment
# category: deployment
# parameters:
#   - name: environment
#     type: string
#     required: true
#     choices: [staging, production]
#     description: Target environment
#   - name: dry_run
#     type: bool
#     default: false
#     description: Preview without executing
# ---

set -euo pipefail
# Implementation...
```

See [specs/01-scanner.md](specs/01-scanner.md) for full metadata schema.

## Usage Examples

```bash
# Interactive mode (default)
tap

# List all discovered scripts
tap list
tap list --category deployment

# Run script directly (headless)
tap run deploy
tap run deploy --param environment=production --param dry_run=true

# Create new script
tap new
tap new backup --category maintenance

# Register existing script
tap add ~/scripts/custom.sh
tap add ~/scripts/tools/ --recursive

# Configuration
tap config                    # Interactive config
tap config show               # Show current config
tap config add-dir ~/scripts  # Add scan directory
```

## Success Criteria

1. **Discoverability** — Adding a metadata header to any script makes it appear in tap
2. **Speed** — TUI launches in <100ms, even with hundreds of scripts
3. **Reliability** — Scripts execute exactly as if run directly
4. **Flexibility** — Works equally well interactive and headless
5. **Simplicity** — Zero config to start, progressive enhancement
