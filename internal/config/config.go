// Package config provides configuration management for tap.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Default values for configuration.
var (
	DefaultExtensions = []string{".sh", ".bash", ".py"}
	DefaultIgnoreDirs = []string{".git", "node_modules", "__pycache__", ".venv", "vendor", ".cache"}
	DefaultMaxDepth   = 10
)

// Config represents tap's configuration.
type Config struct {
	ScanDirs        []string  `yaml:"scan_dirs"`
	Extensions      []string  `yaml:"extensions"`
	IgnoreDirs      []string  `yaml:"ignore_dirs"`
	MaxDepth        int       `yaml:"max_depth"`
	TUI             TUIConfig `yaml:"tui"`
	DefaultShell    string    `yaml:"default_shell"`
	Editor          string    `yaml:"editor"`
	AutoGenMetadata *bool     `yaml:"auto_gen_metadata,omitempty"` // nil = true (default), false = disabled
}

// TUIConfig holds TUI-specific settings.
type TUIConfig struct {
	ShowPaths bool   `yaml:"show_paths"`
	Theme     string `yaml:"theme"`
}

// Registry holds explicitly registered scripts.
type Registry struct {
	Scripts []RegisteredScript `json:"scripts"`
}

// RegisteredScript represents a script that has been explicitly registered.
type RegisteredScript struct {
	Path    string    `json:"path"`
	Alias   string    `json:"alias,omitempty"`
	AddedAt time.Time `json:"added_at"`
}

// Cache holds parsed metadata for faster startup.
type Cache struct {
	Version   int                     `json:"version"`
	Scripts   map[string]CachedScript `json:"scripts"`
	UpdatedAt time.Time               `json:"updated_at"`
}

// CachedScript represents cached script metadata.
type CachedScript struct {
	ModTime time.Time `json:"mod_time"`
	Name    string    `json:"name"`
}

// Manager provides configuration management operations.
type Manager interface {
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

// DefaultManager implements the Manager interface.
type DefaultManager struct {
	configPath   string
	registryPath string
	cachePath    string
}

// NewManager creates a new configuration manager with XDG paths.
func NewManager() (*DefaultManager, error) {
	// Check for environment override
	configPath := os.Getenv("TAP_CONFIG")
	if configPath == "" {
		var err error
		configPath, err = xdg.ConfigFile("tap/config.yaml")
		if err != nil {
			return nil, fmt.Errorf("determining config path: %w", err)
		}
	}

	registryPath, err := xdg.DataFile("tap/registry.json")
	if err != nil {
		return nil, fmt.Errorf("determining registry path: %w", err)
	}

	cachePath, err := xdg.CacheFile("tap/metadata.json")
	if err != nil {
		return nil, fmt.Errorf("determining cache path: %w", err)
	}

	return &DefaultManager{
		configPath:   configPath,
		registryPath: registryPath,
		cachePath:    cachePath,
	}, nil
}

// NewManagerWithPaths creates a manager with explicit paths (useful for testing).
func NewManagerWithPaths(configPath, registryPath, cachePath string) *DefaultManager {
	return &DefaultManager{
		configPath:   configPath,
		registryPath: registryPath,
		cachePath:    cachePath,
	}
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		ScanDirs:        []string{},
		Extensions:      DefaultExtensions,
		IgnoreDirs:      DefaultIgnoreDirs,
		MaxDepth:        DefaultMaxDepth,
		TUI:             TUIConfig{ShowPaths: false, Theme: "default"},
		DefaultShell:    "bash",
		Editor:          "",
		AutoGenMetadata: nil, // nil means true (default)
	}
}

// GetAutoGenMetadata returns the effective value of AutoGenMetadata.
// Returns true if not explicitly set to false.
func (c *Config) GetAutoGenMetadata() bool {
	if c.AutoGenMetadata == nil {
		return true // default to true
	}
	return *c.AutoGenMetadata
}

// Path returns the config file path.
func (m *DefaultManager) Path() string {
	return m.configPath
}

// Load reads configuration from disk or returns defaults.
func (m *DefaultManager) Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			if err := m.ensureDir(m.configPath); err != nil {
				return nil, err
			}
			if err := m.createDefaultConfig(); err != nil {
				return nil, err
			}
			// Merge environment variables and return defaults
			m.mergeEnvVars(cfg)
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Expand ~ in paths
	for i, dir := range cfg.ScanDirs {
		cfg.ScanDirs[i] = expandPath(dir)
	}

	// Merge environment variables
	m.mergeEnvVars(cfg)

	return cfg, nil
}

// mergeEnvVars merges environment variables into the config.
func (m *DefaultManager) mergeEnvVars(cfg *Config) {
	if envDirs := os.Getenv("TAP_SCAN_DIRS"); envDirs != "" {
		for _, dir := range strings.Split(envDirs, ":") {
			dir = strings.TrimSpace(dir)
			if dir != "" {
				cfg.ScanDirs = append(cfg.ScanDirs, expandPath(dir))
			}
		}
	}
}

// Save writes configuration to disk.
func (m *DefaultManager) Save(cfg *Config) error {
	if err := m.ensureDir(m.configPath); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// AddScanDir adds a directory to the scan_dirs list.
func (m *DefaultManager) AddScanDir(dir string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	// Expand and normalize the path
	expandedDir := expandPath(dir)
	absDir, err := filepath.Abs(expandedDir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Verify directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory not found: %s", dir)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Check for duplicates
	for _, existing := range cfg.ScanDirs {
		if expandPath(existing) == absDir {
			return fmt.Errorf("directory already configured: %s", dir)
		}
	}

	// Store the original path (with ~ if applicable) for readability
	cfg.ScanDirs = append(cfg.ScanDirs, dir)
	return m.Save(cfg)
}

// RemoveScanDir removes a directory from the scan_dirs list.
func (m *DefaultManager) RemoveScanDir(dir string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	expandedDir := expandPath(dir)
	absDir, _ := filepath.Abs(expandedDir)

	found := false
	newDirs := make([]string, 0, len(cfg.ScanDirs))
	for _, existing := range cfg.ScanDirs {
		existingExpanded := expandPath(existing)
		existingAbs, _ := filepath.Abs(existingExpanded)
		if existingAbs == absDir || existing == dir {
			found = true
			continue
		}
		newDirs = append(newDirs, existing)
	}

	if !found {
		return fmt.Errorf("directory not in config: %s", dir)
	}

	cfg.ScanDirs = newDirs
	return m.Save(cfg)
}

// GetRegistry returns the script registry.
func (m *DefaultManager) GetRegistry() (*Registry, error) {
	data, err := os.ReadFile(m.registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Scripts: []RegisteredScript{}}, nil
		}
		return nil, fmt.Errorf("reading registry: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("invalid registry: %w", err)
	}

	return &registry, nil
}

// RegisterScript adds a script to the registry.
func (m *DefaultManager) RegisterScript(path string, alias string) error {
	// Expand and validate path
	absPath, err := filepath.Abs(expandPath(path))
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script not found: %s", path)
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory: %s", path)
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

// UnregisterScript removes a script from the registry by path or alias.
func (m *DefaultManager) UnregisterScript(pathOrAlias string) error {
	registry, err := m.GetRegistry()
	if err != nil {
		return err
	}

	absPath, _ := filepath.Abs(expandPath(pathOrAlias))

	found := false
	newScripts := make([]RegisteredScript, 0, len(registry.Scripts))
	for _, s := range registry.Scripts {
		if s.Path == absPath || s.Alias == pathOrAlias {
			found = true
			continue
		}
		newScripts = append(newScripts, s)
	}

	if !found {
		return fmt.Errorf("script not found: %s", pathOrAlias)
	}

	registry.Scripts = newScripts
	return m.saveRegistry(registry)
}

// saveRegistry persists the registry to disk.
func (m *DefaultManager) saveRegistry(registry *Registry) error {
	if err := m.ensureDir(m.registryPath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	return os.WriteFile(m.registryPath, data, 0644)
}

// GetCache returns the metadata cache.
func (m *DefaultManager) GetCache() (*Cache, error) {
	data, err := os.ReadFile(m.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Cache{Version: 1, Scripts: make(map[string]CachedScript)}, nil
		}
		return nil, fmt.Errorf("reading cache: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("invalid cache: %w", err)
	}

	return &cache, nil
}

// SaveCache persists the metadata cache.
func (m *DefaultManager) SaveCache(cache *Cache) error {
	if err := m.ensureDir(m.cachePath); err != nil {
		return err
	}

	cache.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	return os.WriteFile(m.cachePath, data, 0644)
}

// ensureDir creates the parent directory for a file path if it doesn't exist.
func (m *DefaultManager) ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755)
}

// createDefaultConfig creates a default config file with helpful comments.
func (m *DefaultManager) createDefaultConfig() error {
	defaultYAML := `# tap configuration
# Documentation: https://github.com/shaiungar/tap

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
  - .cache

# Maximum scan depth (0 = unlimited)
max_depth: 10

# TUI settings
tui:
  show_paths: false
  theme: default

# Default shell for new scripts (tap new)
default_shell: bash

# Editor for editing scripts (falls back to $EDITOR)
editor: ""
`
	return os.WriteFile(m.configPath, []byte(defaultYAML), 0644)
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}
