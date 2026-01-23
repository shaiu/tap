# tap

A terminal-based script runner that gives developers quick access to their shell and Python scripts through an interactive TUI menu or direct CLI invocation.

## Quick Start

### 1. Install tap

```bash
# Build from source
go build -o tap ./cmd/tap

# Move to your PATH (optional)
mv tap /usr/local/bin/
```

### 2. Create your first script

```bash
# Create a scripts directory
mkdir -p ~/scripts

# Create a script with tap
tap new hello -d "Say hello" -c demo -o ~/scripts/hello.sh
```

Or manually create a script with YAML metadata:

```bash
cat > ~/scripts/hello.sh << 'EOF'
#!/bin/bash
# ---
# name: hello
# description: Say hello world
# category: demo
# ---
echo "Hello, World!"
EOF
chmod +x ~/scripts/hello.sh
```

### 3. Tell tap where to find scripts

**Option A: Add to config (recommended)**

Edit your config file to add scan directories:

```bash
# macOS
nano ~/Library/Application\ Support/tap/config.yaml

# Linux
nano ~/.config/tap/config.yaml
```

Add your scripts folder:
```yaml
scan_dirs:
  - ~/scripts
```

**Option B: Register explicitly**

```bash
tap add ~/scripts --recursive
```

### 4. Launch the TUI

```bash
tap
```

Use arrow keys to navigate, `/` to filter, and `Enter` to run a script.

---

## Installation

### From Source

```bash
git clone https://github.com/shaiungar/tap.git
cd tap
go build -o tap ./cmd/tap
```

### Requirements

- Go 1.21 or later
- A terminal that supports ANSI escape codes

---

## Configuration

tap follows the XDG Base Directory specification, with platform-specific defaults:

| File | macOS | Linux |
|------|-------|-------|
| Config | `~/Library/Application Support/tap/config.yaml` | `~/.config/tap/config.yaml` |
| Registry | `~/Library/Application Support/tap/registry.json` | `~/.local/share/tap/registry.json` |
| Cache | `~/Library/Caches/tap/metadata.json` | `~/.cache/tap/metadata.json` |

You can override the config location with `TAP_CONFIG` environment variable.

### Configuration Options

```yaml
# config.yaml
scan_dirs:
  - ~/scripts
  - ~/work/scripts
extensions:
  - .sh
  - .bash
  - .py
ignore_dirs:
  - .git
  - node_modules
  - __pycache__
  - .venv
  - vendor
max_depth: 10
```

---

## Commands

### `tap` - Launch TUI

Run tap without arguments to launch the interactive TUI:

```bash
tap
```

**TUI Controls:**
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate list |
| `Enter` | Select item / Run script |
| `Esc` or `Backspace` | Go back |
| `/` | Open filter |
| `q` | Quit |

### `tap list` - List Scripts

Display all discovered scripts:

```bash
# List all scripts grouped by category
tap list

# Filter by category
tap list --category demo

# Flat list (no grouping)
tap list --flat

# JSON output (for scripting)
tap list --json
```

### `tap run` - Run a Script

Execute a script directly without the TUI:

```bash
# Run by name
tap run hello

# Run with category prefix
tap run demo/hello

# Pass parameters
tap run greet --param name=World --param loud=true
tap run greet -p name=World -p loud=true

# Inline parameters (no --param flag needed)
tap run greet name=World loud=true

# Mix flags and inline parameters
tap run greet -p name=Alice count=3
```

### `tap new` - Create a Script

Scaffold a new script with proper metadata:

```bash
# Interactive mode (guided form)
tap new

# Interactive with name pre-filled
tap new backup-db

# Headless mode with flags
tap new backup-db -d "Backup database" -c maintenance

# Specify shell and output path
tap new deploy --shell python -o ~/scripts/deploy.py

# With parameters
tap new greet -d "Greet a user" -p "name:string:Name to greet:required"
```

**Parameter format:** `name:type:description:default`
- `name` - Parameter name (required)
- `type` - One of: `string`, `int`, `float`, `bool`, `path`
- `description` - Help text (optional)
- `default` - Default value, or `required` for required params

### `tap add` - Register Scripts

Register external scripts or directories:

```bash
# Register a single script
tap add ~/scripts/my-tool.sh

# Register with an alias
tap add ~/scripts/my-tool.sh --alias tool

# Register all scripts in a directory
tap add ~/company-scripts/ --recursive
```

### `tap remove` - Unregister Scripts

Remove scripts from tap's registry:

```bash
# Remove by path
tap remove ~/scripts/my-tool.sh

# Remove by alias
tap remove tool
```

---

## Script Metadata Format

tap discovers scripts by parsing YAML front matter in comments.

### Bash/Shell Scripts

```bash
#!/bin/bash
# ---
# name: deploy
# description: Deploy application to environment
# category: deployment
# author: platform-team
# version: 1.0.0
# tags: [kubernetes, production]
# parameters:
#   - name: environment
#     type: string
#     required: true
#     choices: [staging, production]
#     short: e
#     description: Target environment
#   - name: version
#     type: string
#     default: latest
#     short: v
#     description: Version to deploy
#   - name: dry_run
#     type: bool
#     default: false
#     description: Preview without executing
# examples:
#   - command: deploy -e production -v v2.1.0
#     description: Deploy v2.1.0 to production
# ---

set -euo pipefail
echo "Deploying $TAP_PARAM_VERSION to $TAP_PARAM_ENVIRONMENT"
```

### Python Scripts

```python
#!/usr/bin/env python3
"""
---
name: process-data
description: Transform and validate data files
category: data
parameters:
  - name: input_file
    type: path
    required: true
    description: Input data file
  - name: format
    type: string
    default: json
    choices: [json, csv, parquet]
---
"""

import os

input_file = os.environ.get("TAP_PARAM_INPUT_FILE")
format = os.environ.get("TAP_PARAM_FORMAT", "json")

print(f"Processing {input_file} as {format}")
```

### Metadata Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier for `tap run <name>` |
| `description` | Yes | One-line description shown in list |
| `category` | No | Grouping for menu (default: "uncategorized") |
| `author` | No | Script author |
| `version` | No | Script version |
| `tags` | No | Additional searchable tags |
| `parameters` | No | Input parameters |
| `examples` | No | Usage examples |

### Parameter Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text value | `name=Alice` |
| `int` | Integer number | `count=5` |
| `float` | Decimal number | `threshold=0.95` |
| `bool` | Boolean | `verbose=true` |
| `path` | File/directory path | `input=/data/file.csv` |

---

## How Parameters Work

When you run a script, tap passes parameters as environment variables with the prefix `TAP_PARAM_`:

| Parameter | Environment Variable |
|-----------|---------------------|
| `name` | `TAP_PARAM_NAME` |
| `dry_run` | `TAP_PARAM_DRY_RUN` |
| `input_file` | `TAP_PARAM_INPUT_FILE` |

### Example Script

```bash
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
#     description: Shout the greeting
# ---

msg="Hello, $TAP_PARAM_NAME!"

if [ "$TAP_PARAM_LOUD" = "true" ]; then
    echo "${msg^^}"
else
    echo "$msg"
fi
```

### Running It

```bash
# Interactive (TUI shows parameter form)
tap run greet

# Headless with params
tap run greet name=World
# Output: Hello, World!

tap run greet name=World loud=true
# Output: HELLO, WORLD!
```

---

## Execution Modes

### Interactive Mode (Default)

When run from a terminal, tap launches the TUI:

```bash
tap           # Launch TUI
tap run greet # Shows form for required params
```

### Headless Mode

For scripts and CI/CD, tap runs without TUI:

```bash
# Explicitly headless
tap --headless run greet name=World

# Auto-detected headless (no TTY)
echo "name=World" | tap run greet

# CI environment (CI or TAP_HEADLESS env vars)
CI=true tap run greet name=World
```

---

## Workflow Examples

### Developer Script Library

```bash
# Create a scripts directory
mkdir -p ~/scripts/{dev,deploy,utils}

# Add it to your config (one-time setup)
# macOS: ~/Library/Application Support/tap/config.yaml
# Linux: ~/.config/tap/config.yaml
#
# scan_dirs:
#   - ~/scripts

# Create some scripts (auto-discovered from scan_dirs)
tap new dev-setup -d "Setup dev environment" -c dev -o ~/scripts/dev/setup.sh
tap new db-migrate -d "Run database migrations" -c dev -o ~/scripts/dev/migrate.sh
tap new deploy-staging -d "Deploy to staging" -c deploy -o ~/scripts/deploy/staging.sh
tap new clean-logs -d "Clean old log files" -c utils -o ~/scripts/utils/clean-logs.sh

# Use tap for quick access
tap                      # Browse all scripts
tap run dev-setup        # Run directly
tap list --category dev  # List dev scripts
```

### CI/CD Integration

```bash
# In your CI pipeline
tap --headless run deploy-staging \
    --param version=$CI_COMMIT_TAG \
    --param dry_run=false
```

### Team Script Sharing

```bash
# Clone team scripts repo
git clone git@company.com:team/scripts.git ~/team-scripts

# Option 1: Add to scan_dirs in config
# Option 2: Register explicitly
tap add ~/team-scripts --recursive

# Everyone has access to the same scripts
tap run team/onboarding
```

### Environment Variable Setup

Add scan directories via environment variable (useful for dotfiles):

```bash
# In your .bashrc or .zshrc
export TAP_SCAN_DIRS="~/scripts:~/work/scripts:~/team-scripts"
```

---

## Tips & Tricks

### Quick Filtering

Press `/` in the TUI to filter scripts by name:

```
/ deploy          # Find scripts with "deploy" in name
/ staging         # Find staging-related scripts
```

### Script Aliases

Give frequently-used scripts short names:

```bash
tap add ~/long/path/to/complex-script.sh --alias cs
tap run cs  # Much easier!
```

### JSON Output for Scripting

```bash
# Get all script names
tap list --json | jq -r '.[].scripts[].name'

# Filter scripts by category
tap list --json | jq -r '.[] | select(.category=="deploy") | .scripts[].name'
```

### Debug Mode

```bash
# See what's happening
tap -v list
tap -v run deploy env=staging
```

---

## Troubleshooting

### "No scripts found"

1. Check your config has scan directories:
   ```bash
   # macOS
   cat ~/Library/Application\ Support/tap/config.yaml

   # Linux
   cat ~/.config/tap/config.yaml
   ```

2. Verify scripts have proper metadata:
   ```bash
   head -10 ~/scripts/your-script.sh
   ```

3. Make sure the `name` and `description` fields are present.

### "Script not found"

Check if the script is registered:
```bash
tap list --json | jq -r '.[].name' | grep your-script
```

### Parameters Not Working

Ensure you're using the correct environment variable format:
- Parameter `my_param` → `TAP_PARAM_MY_PARAM`
- Parameter names are uppercased and prefixed with `TAP_PARAM_`

### Finding Config Location

If you're unsure where tap stores its files:
```bash
# The XDG library determines paths based on your OS
# You can check by looking for existing files:

# macOS
ls ~/Library/Application\ Support/tap/

# Linux
ls ~/.config/tap/
ls ~/.local/share/tap/
```

---

## Project Structure

```
tap/
├── cmd/tap/           # Main entry point
├── internal/
│   ├── cli/           # Cobra commands
│   ├── config/        # Configuration management
│   ├── core/          # Scanner, parser, executor
│   ├── templates/     # Script templates
│   └── tui/           # Bubble Tea TUI
├── specs/             # Feature specifications
└── testdata/          # Test fixtures
```

---

## License

MIT
