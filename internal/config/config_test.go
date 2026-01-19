package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Empty(t, cfg.ScanDirs)
	assert.Equal(t, DefaultExtensions, cfg.Extensions)
	assert.Equal(t, DefaultIgnoreDirs, cfg.IgnoreDirs)
	assert.Equal(t, DefaultMaxDepth, cfg.MaxDepth)
	assert.Equal(t, "bash", cfg.DefaultShell)
	assert.Equal(t, "default", cfg.TUI.Theme)
	assert.False(t, cfg.TUI.ShowPaths)
}

func TestLoad_NoFile_CreatesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	cfg, err := m.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify default config file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify defaults are returned
	assert.Empty(t, cfg.ScanDirs)
	assert.Equal(t, DefaultExtensions, cfg.Extensions)
}

func TestLoad_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Create a custom config
	configYAML := `
scan_dirs:
  - /tmp/scripts
extensions:
  - .sh
max_depth: 5
tui:
  show_paths: true
  theme: minimal
default_shell: zsh
`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	cfg, err := m.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{"/tmp/scripts"}, cfg.ScanDirs)
	assert.Equal(t, []string{".sh"}, cfg.Extensions)
	assert.Equal(t, 5, cfg.MaxDepth)
	assert.True(t, cfg.TUI.ShowPaths)
	assert.Equal(t, "minimal", cfg.TUI.Theme)
	assert.Equal(t, "zsh", cfg.DefaultShell)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Create invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0644)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	_, err = m.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	cfg := &Config{
		ScanDirs:     []string{"/test/path"},
		Extensions:   []string{".sh"},
		IgnoreDirs:   []string{".git"},
		MaxDepth:     5,
		TUI:          TUIConfig{ShowPaths: true, Theme: "custom"},
		DefaultShell: "zsh",
		Editor:       "vim",
	}

	err := m.Save(cfg)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load it back and verify
	loaded, err := m.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.ScanDirs, loaded.ScanDirs)
	assert.Equal(t, cfg.Extensions, loaded.Extensions)
	assert.Equal(t, cfg.MaxDepth, loaded.MaxDepth)
	assert.Equal(t, cfg.TUI.Theme, loaded.TUI.Theme)
}

func TestAddScanDir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Create a directory to add
	scriptDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptDir, 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	// Load to create default config
	_, err = m.Load()
	require.NoError(t, err)

	// Add the directory
	err = m.AddScanDir(scriptDir)
	require.NoError(t, err)

	// Verify it was added
	cfg, err := m.Load()
	require.NoError(t, err)
	assert.Contains(t, cfg.ScanDirs, scriptDir)
}

func TestAddScanDir_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	_, _ = m.Load()

	err := m.AddScanDir("/nonexistent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory not found")
}

func TestAddScanDir_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	scriptDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptDir, 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	_, _ = m.Load()

	err = m.AddScanDir(scriptDir)
	require.NoError(t, err)

	// Try to add again
	err = m.AddScanDir(scriptDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already configured")
}

func TestRemoveScanDir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	scriptDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptDir, 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	_, _ = m.Load()

	// Add then remove
	err = m.AddScanDir(scriptDir)
	require.NoError(t, err)

	err = m.RemoveScanDir(scriptDir)
	require.NoError(t, err)

	// Verify it was removed
	cfg, err := m.Load()
	require.NoError(t, err)
	assert.NotContains(t, cfg.ScanDirs, scriptDir)
}

func TestRemoveScanDir_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	_, _ = m.Load()

	err := m.RemoveScanDir("/not/configured")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in config")
}

func TestRegisterScript(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Create a script file
	scriptPath := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hello"), 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err = m.RegisterScript(scriptPath, "test-alias")
	require.NoError(t, err)

	// Verify it was registered
	registry, err := m.GetRegistry()
	require.NoError(t, err)
	require.Len(t, registry.Scripts, 1)
	assert.Equal(t, scriptPath, registry.Scripts[0].Path)
	assert.Equal(t, "test-alias", registry.Scripts[0].Alias)
}

func TestRegisterScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err := m.RegisterScript("/nonexistent/script.sh", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script not found")
}

func TestRegisterScript_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	scriptPath := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hello"), 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err = m.RegisterScript(scriptPath, "")
	require.NoError(t, err)

	err = m.RegisterScript(scriptPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterScript_DuplicateAlias(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	script1 := filepath.Join(tmpDir, "test1.sh")
	script2 := filepath.Join(tmpDir, "test2.sh")
	err := os.WriteFile(script1, []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(script2, []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err = m.RegisterScript(script1, "myalias")
	require.NoError(t, err)

	err = m.RegisterScript(script2, "myalias")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alias already in use")
}

func TestUnregisterScript(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	scriptPath := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err = m.RegisterScript(scriptPath, "myalias")
	require.NoError(t, err)

	// Unregister by path
	err = m.UnregisterScript(scriptPath)
	require.NoError(t, err)

	registry, err := m.GetRegistry()
	require.NoError(t, err)
	assert.Empty(t, registry.Scripts)
}

func TestUnregisterScript_ByAlias(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	scriptPath := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(scriptPath, []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err = m.RegisterScript(scriptPath, "myalias")
	require.NoError(t, err)

	// Unregister by alias
	err = m.UnregisterScript("myalias")
	require.NoError(t, err)

	registry, err := m.GetRegistry()
	require.NoError(t, err)
	assert.Empty(t, registry.Scripts)
}

func TestUnregisterScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	err := m.UnregisterScript("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script not found")
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/scripts", filepath.Join(home, "scripts")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~something", "~something"}, // Not a home expansion
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Set environment variable
	envDir := filepath.Join(tmpDir, "env-scripts")
	err := os.MkdirAll(envDir, 0755)
	require.NoError(t, err)

	t.Setenv("TAP_SCAN_DIRS", envDir)

	m := NewManagerWithPaths(configPath, registryPath, cachePath)
	cfg, err := m.Load()
	require.NoError(t, err)

	// Verify env var dirs are included
	assert.Contains(t, cfg.ScanDirs, envDir)
}

func TestCache(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")

	m := NewManagerWithPaths(configPath, registryPath, cachePath)

	// Get empty cache
	cache, err := m.GetCache()
	require.NoError(t, err)
	assert.Equal(t, 1, cache.Version)
	assert.Empty(t, cache.Scripts)

	// Save cache
	cache.Scripts["test.sh"] = CachedScript{
		ModTime: time.Now(),
		Name:    "test",
	}
	err = m.SaveCache(cache)
	require.NoError(t, err)

	// Load it back
	loaded, err := m.GetCache()
	require.NoError(t, err)
	assert.Contains(t, loaded.Scripts, "test.sh")
	assert.Equal(t, "test", loaded.Scripts["test.sh"].Name)
}

func TestPath(t *testing.T) {
	m := NewManagerWithPaths("/custom/config.yaml", "/custom/registry.json", "/custom/cache.json")
	assert.Equal(t, "/custom/config.yaml", m.Path())
}
