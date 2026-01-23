package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaiungar/tap/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveScript_ByPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
# ---
# name: test-script
# description: A test script
# category: testing
# ---
echo "Hello"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager with temp paths
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Register the script first
	err = mgr.RegisterScript(scriptPath, "")
	require.NoError(t, err)

	// Verify it's registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)

	// Remove by path
	err = mgr.UnregisterScript(scriptPath)
	require.NoError(t, err)

	// Verify it's removed
	registry, err = mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 0)
}

func TestRemoveScript_ByAlias(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
# ---
# name: test-script
# description: A test script
# category: testing
# ---
echo "Hello"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager with temp paths
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Register the script with an alias
	err = mgr.RegisterScript(scriptPath, "myalias")
	require.NoError(t, err)

	// Verify it's registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
	assert.Equal(t, "myalias", registry.Scripts[0].Alias)

	// Remove by alias
	err = mgr.UnregisterScript("myalias")
	require.NoError(t, err)

	// Verify it's removed
	registry, err = mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 0)
}

func TestRemoveScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config manager with temp paths
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Try to remove non-existent script
	err := mgr.UnregisterScript("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveScript_OneOfMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two valid scripts
	script1Path := filepath.Join(tmpDir, "script1.sh")
	script1Content := `#!/bin/bash
# ---
# name: script1
# description: First script
# category: testing
# ---
echo "Script 1"
`
	err := os.WriteFile(script1Path, []byte(script1Content), 0755)
	require.NoError(t, err)

	script2Path := filepath.Join(tmpDir, "script2.sh")
	script2Content := `#!/bin/bash
# ---
# name: script2
# description: Second script
# category: testing
# ---
echo "Script 2"
`
	err = os.WriteFile(script2Path, []byte(script2Content), 0755)
	require.NoError(t, err)

	// Create config manager with temp paths
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Register both scripts
	err = mgr.RegisterScript(script1Path, "alias1")
	require.NoError(t, err)
	err = mgr.RegisterScript(script2Path, "alias2")
	require.NoError(t, err)

	// Verify both are registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 2)

	// Remove first script by alias
	err = mgr.UnregisterScript("alias1")
	require.NoError(t, err)

	// Verify only one remains
	registry, err = mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
	assert.Equal(t, script2Path, registry.Scripts[0].Path)
	assert.Equal(t, "alias2", registry.Scripts[0].Alias)
}
