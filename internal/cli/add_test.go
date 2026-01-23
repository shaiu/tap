package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaiungar/tap/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSingleScript_Valid(t *testing.T) {
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

	// Add the script
	err = addSingleScript(mgr, scriptPath, "")
	require.NoError(t, err)

	// Verify it was registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
	assert.Equal(t, scriptPath, registry.Scripts[0].Path)
}

func TestAddSingleScript_WithAlias(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid script
	scriptPath := filepath.Join(tmpDir, "long-script-name.sh")
	scriptContent := `#!/bin/bash
# ---
# name: long-script-name
# description: A script with a long name
# category: testing
# ---
echo "Hello"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add with alias
	err = addSingleScript(mgr, scriptPath, "short")
	require.NoError(t, err)

	// Verify alias was set
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
	assert.Equal(t, "short", registry.Scripts[0].Alias)
}

func TestAddSingleScript_NoMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script without metadata
	scriptPath := filepath.Join(tmpDir, "no-meta.sh")
	scriptContent := `#!/bin/bash
echo "Hello"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Should fail - no metadata
	err = addSingleScript(mgr, scriptPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid metadata")
}

func TestAddSingleScript_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script with invalid YAML
	scriptPath := filepath.Join(tmpDir, "invalid.sh")
	scriptContent := `#!/bin/bash
# ---
# name: test
# description: [invalid yaml
# ---
echo "Hello"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Should fail - invalid YAML
	err = addSingleScript(mgr, scriptPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid script")
}

func TestAddSingleScript_AlreadyRegistered(t *testing.T) {
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

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add the script twice
	err = addSingleScript(mgr, scriptPath, "")
	require.NoError(t, err)

	err = addSingleScript(mgr, scriptPath, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestAddDirectory_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with scripts
	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	// Create two valid scripts
	script1 := filepath.Join(scriptsDir, "script1.sh")
	script1Content := `#!/bin/bash
# ---
# name: script1
# description: First script
# category: testing
# ---
echo "Script 1"
`
	err = os.WriteFile(script1, []byte(script1Content), 0755)
	require.NoError(t, err)

	script2 := filepath.Join(scriptsDir, "script2.sh")
	script2Content := `#!/bin/bash
# ---
# name: script2
# description: Second script
# category: testing
# ---
echo "Script 2"
`
	err = os.WriteFile(script2, []byte(script2Content), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add directory
	err = addDirectory(mgr, scriptsDir)
	require.NoError(t, err)

	// Verify both scripts were registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 2)
}

func TestAddDirectory_NoValidScripts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with no valid scripts
	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	// Create a script without metadata
	noMeta := filepath.Join(scriptsDir, "no-meta.sh")
	err = os.WriteFile(noMeta, []byte("#!/bin/bash\necho hi\n"), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Should fail - no valid scripts
	err = addDirectory(mgr, scriptsDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no scripts with valid metadata")
}

func TestAddDirectory_SkipsDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with scripts
	scriptsDir := filepath.Join(tmpDir, "scripts")
	err := os.MkdirAll(scriptsDir, 0755)
	require.NoError(t, err)

	// Create a valid script
	script1 := filepath.Join(scriptsDir, "script1.sh")
	script1Content := `#!/bin/bash
# ---
# name: script1
# description: First script
# category: testing
# ---
echo "Script 1"
`
	err = os.WriteFile(script1, []byte(script1Content), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add directory twice
	err = addDirectory(mgr, scriptsDir)
	require.NoError(t, err)

	// Second time should not error but should skip the duplicate
	err = addDirectory(mgr, scriptsDir)
	require.NoError(t, err)

	// Verify only one script registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
}

func TestRunAdd_PathNotFound(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("alias", "", "")
	cmd.Flags().Bool("recursive", false, "")

	err := runAdd(cmd, []string{"/nonexistent/path/to/script.sh"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path not found")
}

func TestRunAdd_DirectoryWithoutRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().String("alias", "", "")
	cmd.Flags().Bool("recursive", false, "")

	err := runAdd(cmd, []string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
	assert.Contains(t, err.Error(), "--recursive")
}

func TestRunAdd_AliasWithRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().String("alias", "myalias", "")
	cmd.Flags().Bool("recursive", true, "")

	err := runAdd(cmd, []string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--alias cannot be used with --recursive")
}

func TestAddDirectory_Nested(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "scripts", "deployment")
	err := os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Create script in nested directory
	script := filepath.Join(nestedDir, "deploy.sh")
	scriptContent := `#!/bin/bash
# ---
# name: deploy
# description: Deploy script
# category: deployment
# ---
echo "Deploying"
`
	err = os.WriteFile(script, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add parent directory
	err = addDirectory(mgr, filepath.Join(tmpDir, "scripts"))
	require.NoError(t, err)

	// Verify nested script was registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
	assert.Equal(t, script, registry.Scripts[0].Path)
}

func TestAddSingleScript_PythonScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid Python script
	scriptPath := filepath.Join(tmpDir, "test.py")
	scriptContent := `#!/usr/bin/env python3
"""
---
name: python-test
description: A Python test script
category: testing
---
"""

def main():
    print("Hello from Python")

if __name__ == "__main__":
    main()
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create config manager
	registryPath := filepath.Join(tmpDir, "registry.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache.json")
	mgr := config.NewManagerWithPaths(configPath, registryPath, cachePath)

	// Add the script
	err = addSingleScript(mgr, scriptPath, "")
	require.NoError(t, err)

	// Verify it was registered
	registry, err := mgr.GetRegistry()
	require.NoError(t, err)
	assert.Len(t, registry.Scripts, 1)
}
