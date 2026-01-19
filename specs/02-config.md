# Config Spec

> Configuration and state management

## Overview

The config system manages tap's settings, registered scripts, and cached data. It follows XDG Base Directory specifications for cross-platform compatibility.

## File Locations

Following XDG conventions via `github.com/adrg/xdg`:

| Purpose | Path | Content |
|---------|------|---------|
| Configuration | `~/.config/tap/config.yaml` | User settings |
| Data | `~/.local/share/tap/registry.json` | Registered scripts |
| Cache | `~/.cache/tap/metadata.json` | Parsed metadata cache |

On macOS, these resolve to:
- Config: `~/Library/Application Support/tap/config.yaml`
- Data: `~/Library/Application Support/tap/registry.json`
- Cache: `~/Library/Caches/tap/metadata.json`

## Configuration Schema

### config.yaml

```yaml
# Directories to scan for scripts
scan_dirs:
  - ~/scripts
  - ~/.local/bin
  - ~/projects/devops/scripts

# File extensions to consider (default shown)
extensions:
  - .sh
  - .bash
  - .py

# Directories to skip during scanning (default shown)
ignore_dirs:
  - .git
  - node_modules
  - __pycache__
  - .venv
  - vendor
  - .cache

# Maximum directory depth for scanning (0 = unlimited)
max_depth: 10

# TUI settings
tui:
  # Show script path in list
  show_paths: false
  # Theme: "default", "minimal", "colorful"
  theme: default

# Default shell for new scripts
default_shell: bash

# Editor for `tap edit` (falls back to $EDITOR)
editor: ""
```

### Default Configuration

When no config file exists, tap uses these defaults:

```go
var DefaultConfig = Config{
    ScanDirs:   []string{},  // Empty by default, user must configure
    Extensions: []string{".sh", ".bash", ".py"},
    IgnoreDirs: []string{".git", "node_modules", "__pycache__", ".venv", "vendor", ".cache"},
    MaxDepth:   10,
    TUI: TUIConfig{
        ShowPaths: false,
        Theme:     "default",
    },
    DefaultShell: "bash",
    Editor:       "",
}
```

## Data Structures

### Config

```go
type Config struct {
    ScanDirs     []string  `yaml:"scan_dirs"`
    Extensions   []string  `yaml:"extensions"`
    IgnoreDirs   []string  `yaml:"ignore_dirs"`
    MaxDepth     int       `yaml:"max_depth"`
    TUI          TUIConfig `yaml:"tui"`
    DefaultShell string    `yaml:"default_shell"`
    Editor       string    `yaml:"editor"`
}

type TUIConfig struct {
    ShowPaths bool   `yaml:"show_paths"`
    Theme     string `yaml:"theme"`
}
```

### Registry (Explicitly Registered Scripts)

```go
type Registry struct {
    Scripts []RegisteredScript `json:"scripts"`
}

type RegisteredScript struct {
    Path      string    `json:"path"`      // Absolute path
    Alias     string    `json:"alias"`     // Optional alias (overrides name)
    AddedAt   time.Time `json:"added_at"`
}
```

## Config Interface

```go
type ConfigManager interface {
    // Load reads configuration from disk (or returns defaults)
    Load() (*Config, error)
    
    // Save writes configuration to disk
    Save(cfg *Config) error
    
    // Path returns the config file path
    Path() string
    
    // AddScanDir adds a directory to scan_dirs
    AddScanDir(dir string) error
    
    // RemoveScanDir removes a directory from scan_dirs
    RemoveScanDir(dir string) error
    
    // GetRegistry returns registered scripts
    GetRegistry() (*Registry, error)
    
    // RegisterScript adds a script to the registry
    RegisterScript(path string, alias string) error
    
    // UnregisterScript removes a script from the registry
    UnregisterScript(pathOrAlias string) error
    
    // GetCache returns the metadata cache
    GetCache() (*Cache, error)
    
    // SaveCache persists the metadata cache
    SaveCache(cache *Cache) error
}
```

## Implementation

### Initialization

On first run, tap should:

1. Create config directory if it doesn't exist
2. Create a default config file with helpful comments
3. Prompt user to add their first scan directory (optional)

```go
func (m *configManager) ensureConfigDir() error {
    configDir := filepath.Dir(m.configPath)
    return os.MkdirAll(configDir, 0755)
}

func (m *configManager) createDefaultConfig() error {
    defaultYAML := `# tap configuration
# Documentation: https://github.com/user/tap

# Directories to scan for scripts
# Add your script directories here, e.g.:
#   - ~/scripts
#   - ~/projects/devops
scan_dirs: []

# File extensions to look for
extensions:
  - .sh
  - .bash
  - .py

# Directories to skip
ignore_dirs:
  - .git
  - node_modules
  - __pycache__
  - .venv
  - vendor

# Maximum scan depth (0 = unlimited)
max_depth: 10

# TUI settings
tui:
  show_paths: false
  theme: default

# Default shell for new scripts (tap new)
default_shell: bash
`
    return os.WriteFile(m.configPath, []byte(defaultYAML), 0644)
}
```

### Loading Configuration

```go
func (m *configManager) Load() (*Config, error) {
    // Start with defaults
    cfg := DefaultConfig
    
    data, err := os.ReadFile(m.configPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Create default config
            if err := m.createDefaultConfig(); err != nil {
                return nil, err
            }
            return &cfg, nil
        }
        return nil, err
    }
    
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    // Expand ~ in paths
    for i, dir := range cfg.ScanDirs {
        cfg.ScanDirs[i] = expandPath(dir)
    }
    
    return &cfg, nil
}
```

### Registry Management

```go
func (m *configManager) RegisterScript(path string, alias string) error {
    // Expand and validate path
    absPath, err := filepath.Abs(expandPath(path))
    if err != nil {
        return err
    }
    
    if _, err := os.Stat(absPath); err != nil {
        return fmt.Errorf("script not found: %s", path)
    }
    
    registry, err := m.GetRegistry()
    if err != nil {
        registry = &Registry{}
    }
    
    // Check for duplicates
    for _, s := range registry.Scripts {
        if s.Path == absPath {
            return fmt.Errorf("script already registered: %s", path)
        }
        if alias != "" && s.Alias == alias {
            return fmt.Errorf("alias already in use: %s", alias)
        }
    }
    
    registry.Scripts = append(registry.Scripts, RegisteredScript{
        Path:    absPath,
        Alias:   alias,
        AddedAt: time.Now(),
    })
    
    return m.saveRegistry(registry)
}
```

### Path Expansion

```go
func expandPath(path string) string {
    if strings.HasPrefix(path, "~/") {
        home, err := os.UserHomeDir()
        if err == nil {
            return filepath.Join(home, path[2:])
        }
    }
    return path
}
```

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `TAP_CONFIG` | Override config file path |
| `TAP_SCAN_DIRS` | Additional scan directories (colon-separated) |
| `TAP_NO_COLOR` | Disable color output |
| `EDITOR` | Default editor for `tap edit` |

```go
func (m *configManager) Load() (*Config, error) {
    cfg := DefaultConfig
    
    // Load from file...
    
    // Merge environment variables
    if envDirs := os.Getenv("TAP_SCAN_DIRS"); envDirs != "" {
        for _, dir := range strings.Split(envDirs, ":") {
            if dir != "" {
                cfg.ScanDirs = append(cfg.ScanDirs, expandPath(dir))
            }
        }
    }
    
    return &cfg, nil
}
```

## First-Run Experience

When tap detects no configuration (no scan_dirs configured):

```
Welcome to tap! 🎯

No script directories configured yet.

Would you like to add a directory to scan for scripts?
Enter path (or press Enter to skip): ~/scripts

✓ Added ~/scripts to scan directories

You can add more directories with:
  tap config add-dir <path>

Run `tap` to browse your scripts!
```

## Commands

### tap config

```bash
# Show current configuration
tap config show

# Edit config in $EDITOR
tap config edit

# Add scan directory
tap config add-dir ~/scripts

# Remove scan directory  
tap config remove-dir ~/old-scripts

# Set a value
tap config set tui.theme minimal
tap config set default_shell python

# Reset to defaults
tap config reset
```

## Testing

### Unit Tests

```go
func TestLoad_Default(t *testing.T)
func TestLoad_WithFile(t *testing.T)
func TestLoad_InvalidYAML(t *testing.T)
func TestSave(t *testing.T)
func TestAddScanDir(t *testing.T)
func TestRemoveScanDir(t *testing.T)
func TestRegisterScript(t *testing.T)
func TestUnregisterScript(t *testing.T)
func TestExpandPath(t *testing.T)
func TestEnvironmentOverrides(t *testing.T)
```

## Migration

If configuration format changes in future versions:

```go
type ConfigVersion struct {
    Version int `yaml:"version"`
}

func (m *configManager) migrate(data []byte) ([]byte, error) {
    var v ConfigVersion
    yaml.Unmarshal(data, &v)
    
    switch v.Version {
    case 0, 1:
        // Current version, no migration needed
        return data, nil
    default:
        return nil, fmt.Errorf("unknown config version: %d", v.Version)
    }
}
```
