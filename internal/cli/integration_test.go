package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
	"github.com/shaiungar/tap/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ScanToTUIToExecution tests the full pipeline:
// scan scripts -> create TUI model -> select script -> execute
func TestIntegration_ScanToTUIToExecution(t *testing.T) {
	// Get the e2e test directory
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Step 1: Scan scripts
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh", ".py"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(scripts), 4, "Expected at least 4 scripts")

	// Step 2: Organize into categories
	categories := core.OrganizeByCategory(scripts)
	assert.GreaterOrEqual(t, len(categories), 2, "Expected at least 2 categories")

	// Verify we have expected categories
	categoryNames := make([]string, len(categories))
	for i, c := range categories {
		categoryNames[i] = c.Name
	}
	assert.Contains(t, categoryNames, "demo")
	assert.Contains(t, categoryNames, "utils")

	// Step 3: Create TUI model
	model := tui.NewAppModel(categories)
	assert.Equal(t, tui.StateBrowsing, model.State())
	assert.Equal(t, len(categories), len(model.Categories()))

	// Step 4: Simulate navigation - find hello-world script
	var helloScript *core.Script
	for _, cat := range categories {
		for i := range cat.Scripts {
			if cat.Scripts[i].Name == "hello-world" {
				helloScript = &cat.Scripts[i]
				break
			}
		}
	}
	require.NotNil(t, helloScript, "hello-world script not found")

	// Step 5: Execute the script
	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script: *helloScript,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, stdout.String(), "Hello, World!")
}

// TestIntegration_ExitCodePassthrough verifies exit codes are passed through
func TestIntegration_ExitCodePassthrough(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Find the exit-code test script
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	var exitCodeScript *core.Script
	for i := range scripts {
		if scripts[i].Name == "test-exit-code" {
			exitCodeScript = &scripts[i]
			break
		}
	}
	require.NotNil(t, exitCodeScript, "test-exit-code script not found")

	// Execute and verify exit code
	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script: *exitCodeScript,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result.ExitCode, "Exit code should be 42")
	assert.Contains(t, stdout.String(), "exits with code 42")
}

// TestIntegration_TUINavigation tests TUI navigation through keyboard messages
func TestIntegration_TUINavigation(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Scan scripts
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)
	var m tea.Model = tui.NewAppModel(categories)

	// Verify initial state
	model := m.(tui.AppModel)
	assert.Equal(t, tui.StateBrowsing, model.State())

	// Simulate window size message
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = m.(tui.AppModel)
	assert.Equal(t, 80, model.Width())
	assert.Equal(t, 24, model.Height())

	// Simulate pressing '?' for help
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = m.(tui.AppModel)
	assert.Equal(t, tui.StateHelp, model.State())

	// Press any key to exit help
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = m.(tui.AppModel)
	assert.Equal(t, tui.StateBrowsing, model.State())

	// Test quit returns tea.Quit command
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	// cmd should be tea.Quit
	assert.NotNil(t, cmd)
}

// TestIntegration_FilterFunctionality tests the filter overlay
func TestIntegration_FilterFunctionality(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)
	var m tea.Model = tui.NewAppModel(categories)

	// Activate filter with '/'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model := m.(tui.AppModel)
	assert.Equal(t, tui.StateFilter, model.State())

	// Cancel filter with Escape
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = m.(tui.AppModel)
	assert.Equal(t, tui.StateBrowsing, model.State())
}

// TestIntegration_EnvironmentVariables verifies TAP_* env vars are set
func TestIntegration_EnvironmentVariables(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Find show-env script
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	var showEnvScript *core.Script
	for i := range scripts {
		if scripts[i].Name == "show-env" {
			showEnvScript = &scripts[i]
			break
		}
	}
	require.NotNil(t, showEnvScript, "show-env script not found")

	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script: *showEnvScript,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// Verify TAP environment variables are present
	output := stdout.String()
	assert.Contains(t, output, "show-env", "Should contain script name")
}

// TestIntegration_RunCommand_Basic tests the basic run command flow
func TestIntegration_RunCommand_Basic(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Scan scripts
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	// Test finding script by name
	script, err := findScript(categories, "hello-world")
	require.NoError(t, err)
	assert.Equal(t, "hello-world", script.Name)
	assert.Equal(t, "demo", script.Category)

	// Test finding script by category/name
	script, err = findScript(categories, "demo/hello-world")
	require.NoError(t, err)
	assert.Equal(t, "hello-world", script.Name)

	// Test executing found script
	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script: *script,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, stdout.String(), "Hello, World!")
}

// TestIntegration_RunCommand_WithParams tests running scripts with parameters
func TestIntegration_RunCommand_WithParams(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	// Scan scripts
	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	// Find the with-params script
	script, err := findScript(categories, "with-params")
	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, "with-params", script.Name)
	assert.Len(t, script.Parameters, 4, "Expected 4 parameters")

	// Test with required param provided
	params := map[string]string{"name": "World"}

	// Apply defaults
	params = applyDefaults(*script, params)

	// Validate
	err = validateParams(*script, params)
	require.NoError(t, err)

	// Execute
	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script:     *script,
		Parameters: params,
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, stdout.String(), "Hello, World!")
	assert.Contains(t, stdout.String(), "env: dev") // default value
}

// TestIntegration_RunCommand_InlineParams tests inline param=value syntax
func TestIntegration_RunCommand_InlineParams(t *testing.T) {
	// Test parseInlineParams with real args
	args := []string{"name=Alice", "count=3", "loud=true"}
	params := parseInlineParams(args)

	assert.Equal(t, "Alice", params["name"])
	assert.Equal(t, "3", params["count"])
	assert.Equal(t, "true", params["loud"])
}

// TestIntegration_RunCommand_FlagParams tests --param flag parsing
func TestIntegration_RunCommand_FlagParams(t *testing.T) {
	flags := []string{"name=Bob", "env=prod"}
	params := parseParamFlags(flags)

	assert.Equal(t, "Bob", params["name"])
	assert.Equal(t, "prod", params["env"])
}

// TestIntegration_RunCommand_ParamMerge tests that inline params override flag params
func TestIntegration_RunCommand_ParamMerge(t *testing.T) {
	flagParams := map[string]string{"name": "FlagName", "count": "5"}
	inlineParams := map[string]string{"name": "InlineName"} // overrides flag

	merged := mergeParams(flagParams, inlineParams)

	assert.Equal(t, "InlineName", merged["name"]) // inline wins
	assert.Equal(t, "5", merged["count"])         // flag preserved
}

// TestIntegration_RunCommand_MissingRequiredParam tests error for missing required param
func TestIntegration_RunCommand_MissingRequiredParam(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Don't provide required 'name' param
	params := map[string]string{}

	err = validateParams(*script, params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required parameter: name")
}

// TestIntegration_RunCommand_InvalidParamType tests error for invalid param type
func TestIntegration_RunCommand_InvalidParamType(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Provide invalid int value
	params := map[string]string{
		"name":  "Test",
		"count": "not-a-number", // invalid int
	}

	err = validateParams(*script, params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected integer")
}

// TestIntegration_RunCommand_InvalidChoice tests error for invalid choice value
func TestIntegration_RunCommand_InvalidChoice(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Provide invalid choice value
	params := map[string]string{
		"name": "Test",
		"env":  "invalid-env", // not in choices
	}

	err = validateParams(*script, params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}

// TestIntegration_RunCommand_DefaultValues tests that defaults are applied
func TestIntegration_RunCommand_DefaultValues(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Only provide required param
	params := map[string]string{"name": "Test"}
	params = applyDefaults(*script, params)

	// Verify defaults were applied
	assert.Equal(t, "Test", params["name"])
	assert.Equal(t, "1", params["count"])     // default
	assert.Equal(t, "false", params["loud"])  // default
	assert.Equal(t, "dev", params["env"])     // default
}

// TestIntegration_RunCommand_NeedsParamInput tests interactive param detection
func TestIntegration_RunCommand_NeedsParamInput(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Missing required param - needs input
	assert.True(t, needsParamInput(*script, map[string]string{}))

	// Required param provided - doesn't need input
	assert.False(t, needsParamInput(*script, map[string]string{"name": "Test"}))
}

// TestIntegration_RunCommand_ExecuteWithAllParams tests full execution with all params
func TestIntegration_RunCommand_ExecuteWithAllParams(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	script, err := findScript(categories, "with-params")
	require.NoError(t, err)

	// Provide all params explicitly
	params := map[string]string{
		"name":  "Claude",
		"count": "2",
		"loud":  "true",
		"env":   "prod",
	}

	err = validateParams(*script, params)
	require.NoError(t, err)

	executor := core.NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script:     *script,
		Parameters: params,
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	output := stdout.String()
	// Loud=true means uppercase, count=2 means two lines
	assert.Contains(t, output, "HELLO, CLAUDE! (ENV: PROD)")
	// Should appear twice (count=2)
	assert.Equal(t, 2, strings.Count(output, "HELLO, CLAUDE! (ENV: PROD)"))
}

// TestIntegration_RunCommand_ScriptNotFound tests error when script doesn't exist
func TestIntegration_RunCommand_ScriptNotFound(t *testing.T) {
	e2eDir := filepath.Join(getProjectRoot(t), "testdata", "e2e")
	if _, err := os.Stat(e2eDir); os.IsNotExist(err) {
		t.Skip("E2E testdata directory not found")
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories: []string{e2eDir},
		Extensions:  []string{".sh"},
		IgnoreDirs:  []string{".git"},
		MaxDepth:    10,
	})

	scripts, err := scanner.Scan(context.Background())
	require.NoError(t, err)

	categories := core.OrganizeByCategory(scripts)

	_, err = findScript(categories, "nonexistent-script")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "script not found")
}

// getProjectRoot returns the project root directory
func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
