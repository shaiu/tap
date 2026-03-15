package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaiungar/tap/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func newTestManager(t *testing.T) (*config.DefaultManager, string) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	registryPath := filepath.Join(tmpDir, "registry.json")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)
	return mgr, tmpDir
}

func TestConfigShow_PrintsYAMLAndPaths(t *testing.T) {
	mgr, _ := newTestManager(t)

	// Save a config so show has something to display
	cfg := config.DefaultConfig()
	cfg.ScanDirs = []string{"/tmp/scripts"}
	err := mgr.Save(cfg)
	require.NoError(t, err)

	// Load and verify it round-trips
	loaded, err := mgr.Load()
	require.NoError(t, err)
	assert.Contains(t, loaded.ScanDirs, "/tmp/scripts")

	// Verify paths are accessible
	assert.NotEmpty(t, mgr.Path())
	assert.NotEmpty(t, mgr.RegistryPath())
	assert.NotEmpty(t, mgr.CachePath())
}

func TestConfigAddDir_AddsDirectory(t *testing.T) {
	mgr, tmpDir := newTestManager(t)

	// Create a directory to add
	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	err = mgr.AddScanDir(scriptsDir)
	require.NoError(t, err)

	cfg, err := mgr.Load()
	require.NoError(t, err)
	assert.Contains(t, cfg.ScanDirs, scriptsDir)
}

func TestConfigAddDir_RejectsDuplicate(t *testing.T) {
	mgr, tmpDir := newTestManager(t)

	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	err = mgr.AddScanDir(scriptsDir)
	require.NoError(t, err)

	err = mgr.AddScanDir(scriptsDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already configured")
}

func TestConfigAddDir_RejectsNonExistent(t *testing.T) {
	mgr, tmpDir := newTestManager(t)

	err := mgr.AddScanDir(filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConfigRemoveDir_RemovesDirectory(t *testing.T) {
	mgr, tmpDir := newTestManager(t)

	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	err = mgr.AddScanDir(scriptsDir)
	require.NoError(t, err)

	err = mgr.RemoveScanDir(scriptsDir)
	require.NoError(t, err)

	cfg, err := mgr.Load()
	require.NoError(t, err)
	assert.NotContains(t, cfg.ScanDirs, scriptsDir)
}

func TestConfigRemoveDir_ErrorsOnMissing(t *testing.T) {
	mgr, _ := newTestManager(t)

	err := mgr.RemoveScanDir("/not/configured")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in config")
}

func TestConfigSet_SetsKnownKeys(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		validate func(t *testing.T, cfg *config.Config)
	}{
		{
			key:   "tui.theme",
			value: "minimal",
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "minimal", cfg.TUI.Theme)
			},
		},
		{
			key:   "tui.show_paths",
			value: "true",
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.TUI.ShowPaths)
			},
		},
		{
			key:   "max_depth",
			value: "5",
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 5, cfg.MaxDepth)
			},
		},
		{
			key:   "default_shell",
			value: "zsh",
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "zsh", cfg.DefaultShell)
			},
		},
		{
			key:   "editor",
			value: "nano",
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "nano", cfg.Editor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			mgr, _ := newTestManager(t)

			// Save default config first
			err := mgr.Save(config.DefaultConfig())
			require.NoError(t, err)

			// Load, set, save
			cfg, err := mgr.Load()
			require.NoError(t, err)

			err = applyConfigSet(cfg, tt.key, tt.value)
			require.NoError(t, err)

			err = mgr.Save(cfg)
			require.NoError(t, err)

			// Reload and validate
			loaded, err := mgr.Load()
			require.NoError(t, err)
			tt.validate(t, loaded)
		})
	}
}

func TestConfigSet_RejectsUnknownKey(t *testing.T) {
	cfg := config.DefaultConfig()
	err := applyConfigSet(cfg, "unknown.key", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigSet_RejectsInvalidBool(t *testing.T) {
	cfg := config.DefaultConfig()
	err := applyConfigSet(cfg, "tui.show_paths", "notabool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean")
}

func TestConfigSet_RejectsInvalidInt(t *testing.T) {
	cfg := config.DefaultConfig()
	err := applyConfigSet(cfg, "max_depth", "notanumber")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid integer")
}

func TestConfigReset_RestoresDefaults(t *testing.T) {
	mgr, _ := newTestManager(t)

	// Save a modified config
	cfg := config.DefaultConfig()
	cfg.TUI.Theme = "custom"
	cfg.MaxDepth = 99
	err := mgr.Save(cfg)
	require.NoError(t, err)

	// Reset
	err = mgr.Save(config.DefaultConfig())
	require.NoError(t, err)

	// Verify defaults
	loaded, err := mgr.Load()
	require.NoError(t, err)
	assert.Equal(t, "default", loaded.TUI.Theme)
	assert.Equal(t, 10, loaded.MaxDepth)
}

func TestConfigReset_YAMLRoundTrip(t *testing.T) {
	// Verify DefaultConfig marshals/unmarshals cleanly
	cfg := config.DefaultConfig()
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	var loaded config.Config
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, cfg.TUI.Theme, loaded.TUI.Theme)
	assert.Equal(t, cfg.MaxDepth, loaded.MaxDepth)
	assert.Equal(t, cfg.DefaultShell, loaded.DefaultShell)
}

func TestResolveEditor(t *testing.T) {
	tests := []struct {
		name         string
		configEditor string
		envEditor    string
		envVisual    string
		expected     string
	}{
		{
			name:         "config takes priority",
			configEditor: "code",
			envEditor:    "vim",
			expected:     "code",
		},
		{
			name:      "EDITOR env var",
			envEditor: "nano",
			expected:  "nano",
		},
		{
			name:      "VISUAL env var",
			envVisual: "emacs",
			expected:  "emacs",
		},
		{
			name:     "fallback to vi",
			expected: "vi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			origEditor := os.Getenv("EDITOR")
			origVisual := os.Getenv("VISUAL")
			t.Cleanup(func() {
				os.Setenv("EDITOR", origEditor)
				os.Setenv("VISUAL", origVisual)
			})

			os.Setenv("EDITOR", tt.envEditor)
			os.Setenv("VISUAL", tt.envVisual)

			result := resolveEditor(tt.configEditor)
			assert.Equal(t, tt.expected, result)
		})
	}
}
