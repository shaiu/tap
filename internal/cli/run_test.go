package cli

import (
	"testing"

	"github.com/shaiungar/tap/internal/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInlineParams(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: map[string]string{},
		},
		{
			name:     "single param",
			args:     []string{"env=production"},
			expected: map[string]string{"env": "production"},
		},
		{
			name:     "multiple params",
			args:     []string{"env=production", "version=v2.1"},
			expected: map[string]string{"env": "production", "version": "v2.1"},
		},
		{
			name:     "param with equals in value",
			args:     []string{"query=a=b"},
			expected: map[string]string{"query": "a=b"},
		},
		{
			name:     "empty value",
			args:     []string{"flag="},
			expected: map[string]string{"flag": ""},
		},
		{
			name:     "ignore non-param args",
			args:     []string{"notaparam", "valid=yes"},
			expected: map[string]string{"valid": "yes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInlineParams(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseParamFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		expected map[string]string
	}{
		{
			name:     "empty flags",
			flags:    []string{},
			expected: map[string]string{},
		},
		{
			name:     "single flag",
			flags:    []string{"env=staging"},
			expected: map[string]string{"env": "staging"},
		},
		{
			name:     "multiple flags",
			flags:    []string{"env=staging", "debug=true"},
			expected: map[string]string{"env": "staging", "debug": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseParamFlags(tt.flags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeParams(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name:     "empty maps",
			base:     map[string]string{},
			override: map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "override takes precedence",
			base:     map[string]string{"env": "staging"},
			override: map[string]string{"env": "production"},
			expected: map[string]string{"env": "production"},
		},
		{
			name:     "combines unique keys",
			base:     map[string]string{"env": "staging"},
			override: map[string]string{"version": "v2"},
			expected: map[string]string{"env": "staging", "version": "v2"},
		},
		{
			name:     "partial overlap",
			base:     map[string]string{"a": "1", "b": "2"},
			override: map[string]string{"b": "3", "c": "4"},
			expected: map[string]string{"a": "1", "b": "3", "c": "4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeParams(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindScript(t *testing.T) {
	categories := []core.Category{
		{
			Name: "deploy",
			Scripts: []core.Script{
				{Name: "app", Category: "deploy", Path: "/scripts/deploy/app.sh"},
				{Name: "db", Category: "deploy", Path: "/scripts/deploy/db.sh"},
			},
		},
		{
			Name: "utils",
			Scripts: []core.Script{
				{Name: "cleanup", Category: "utils", Path: "/scripts/utils/cleanup.sh"},
			},
		},
		{
			Name: "data",
			Scripts: []core.Script{
				{Name: "app", Category: "data", Path: "/scripts/data/app.sh"}, // Same name as deploy/app
			},
		},
	}

	tests := []struct {
		name        string
		scriptName  string
		expectError bool
		expectName  string
		expectCat   string
	}{
		{
			name:       "find by name",
			scriptName: "cleanup",
			expectName: "cleanup",
			expectCat:  "utils",
		},
		{
			name:       "find by category/name",
			scriptName: "deploy/app",
			expectName: "app",
			expectCat:  "deploy",
		},
		{
			name:       "find other category/name",
			scriptName: "data/app",
			expectName: "app",
			expectCat:  "data",
		},
		{
			name:        "ambiguous name error",
			scriptName:  "app",
			expectError: true,
		},
		{
			name:        "not found error",
			scriptName:  "nonexistent",
			expectError: true,
		},
		{
			name:        "wrong category",
			scriptName:  "utils/nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := findScript(categories, tt.scriptName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, script)
			} else {
				require.NoError(t, err)
				require.NotNil(t, script)
				assert.Equal(t, tt.expectName, script.Name)
				assert.Equal(t, tt.expectCat, script.Category)
			}
		})
	}
}

func TestNeedsParamInput(t *testing.T) {
	tests := []struct {
		name     string
		script   core.Script
		provided map[string]string
		expected bool
	}{
		{
			name:     "no params",
			script:   core.Script{},
			provided: map[string]string{},
			expected: false,
		},
		{
			name: "optional param missing",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "opt", Required: false}},
			},
			provided: map[string]string{},
			expected: false,
		},
		{
			name: "required param with default",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true, Default: "dev"}},
			},
			provided: map[string]string{},
			expected: false,
		},
		{
			name: "required param missing",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true}},
			},
			provided: map[string]string{},
			expected: true,
		},
		{
			name: "required param provided",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true}},
			},
			provided: map[string]string{"env": "prod"},
			expected: false,
		},
		{
			name: "some required missing",
			script: core.Script{
				Parameters: []core.Parameter{
					{Name: "env", Required: true},
					{Name: "version", Required: true},
				},
			},
			provided: map[string]string{"env": "prod"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsParamInput(tt.script, tt.provided)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name        string
		script      core.Script
		params      map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no params required",
			script:      core.Script{},
			params:      map[string]string{},
			expectError: false,
		},
		{
			name: "required param provided",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true, Type: "string"}},
			},
			params:      map[string]string{"env": "prod"},
			expectError: false,
		},
		{
			name: "required param missing",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true}},
			},
			params:      map[string]string{},
			expectError: true,
			errorMsg:    "missing required parameter: env",
		},
		{
			name: "required with default not required",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Required: true, Default: "dev"}},
			},
			params:      map[string]string{},
			expectError: false,
		},
		{
			name: "invalid int type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "count", Type: "int"}},
			},
			params:      map[string]string{"count": "abc"},
			expectError: true,
			errorMsg:    "expected integer",
		},
		{
			name: "valid int type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "count", Type: "int"}},
			},
			params:      map[string]string{"count": "42"},
			expectError: false,
		},
		{
			name: "invalid float type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "rate", Type: "float"}},
			},
			params:      map[string]string{"rate": "not-a-number"},
			expectError: true,
			errorMsg:    "expected number",
		},
		{
			name: "valid float type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "rate", Type: "float"}},
			},
			params:      map[string]string{"rate": "3.14"},
			expectError: false,
		},
		{
			name: "invalid bool type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "debug", Type: "bool"}},
			},
			params:      map[string]string{"debug": "yes"},
			expectError: true,
			errorMsg:    "expected boolean",
		},
		{
			name: "valid bool type",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "debug", Type: "bool"}},
			},
			params:      map[string]string{"debug": "true"},
			expectError: false,
		},
		{
			name: "valid choice",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Choices: []any{"dev", "staging", "prod"}}},
			},
			params:      map[string]string{"env": "staging"},
			expectError: false,
		},
		{
			name: "invalid choice",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Choices: []any{"dev", "staging", "prod"}}},
			},
			params:      map[string]string{"env": "invalid"},
			expectError: true,
			errorMsg:    "must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParams(tt.script, tt.params)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		script   core.Script
		params   map[string]string
		expected map[string]string
	}{
		{
			name:     "no defaults",
			script:   core.Script{},
			params:   map[string]string{"a": "1"},
			expected: map[string]string{"a": "1"},
		},
		{
			name: "apply default for missing",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Default: "dev"}},
			},
			params:   map[string]string{},
			expected: map[string]string{"env": "dev"},
		},
		{
			name: "don't override provided",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "env", Default: "dev"}},
			},
			params:   map[string]string{"env": "prod"},
			expected: map[string]string{"env": "prod"},
		},
		{
			name: "mixed defaults and provided",
			script: core.Script{
				Parameters: []core.Parameter{
					{Name: "env", Default: "dev"},
					{Name: "version", Default: "v1"},
				},
			},
			params:   map[string]string{"env": "prod"},
			expected: map[string]string{"env": "prod", "version": "v1"},
		},
		{
			name: "int default",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "count", Type: "int", Default: 5}},
			},
			params:   map[string]string{},
			expected: map[string]string{"count": "5"},
		},
		{
			name: "bool default",
			script: core.Script{
				Parameters: []core.Parameter{{Name: "debug", Type: "bool", Default: true}},
			},
			params:   map[string]string{},
			expected: map[string]string{"debug": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyDefaults(tt.script, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunCmd_Registered(t *testing.T) {
	cmd := RootCmd()

	var runCommand *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Use == "run <script> [flags] [param=value...]" {
			runCommand = sub
			break
		}
	}

	require.NotNil(t, runCommand, "run command should exist")
	assert.Equal(t, "Run a script", runCommand.Short)
}

func TestRunCmd_HasCorrectFlags(t *testing.T) {
	cmd := RootCmd()

	var runCommand *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Use == "run <script> [flags] [param=value...]" {
			runCommand = sub
			break
		}
	}

	require.NotNil(t, runCommand)

	paramFlag := runCommand.Flags().Lookup("param")
	assert.NotNil(t, paramFlag)
	assert.Equal(t, "p", paramFlag.Shorthand)
	assert.Equal(t, "stringSlice", paramFlag.Value.Type())
}

func TestScriptNotFoundError_WithAvailable(t *testing.T) {
	categories := []core.Category{
		{
			Name: "test",
			Scripts: []core.Script{
				{Name: "script1"},
				{Name: "script2"},
			},
		},
	}

	err := scriptNotFoundError(categories, "missing")
	assert.Contains(t, err.Error(), "script not found: missing")
	assert.Contains(t, err.Error(), "Available scripts")
	assert.Contains(t, err.Error(), "script1")
}

func TestScriptNotFoundError_Empty(t *testing.T) {
	categories := []core.Category{}

	err := scriptNotFoundError(categories, "missing")
	assert.Contains(t, err.Error(), "script not found: missing")
	assert.Contains(t, err.Error(), "No scripts available")
	assert.Contains(t, err.Error(), "tap config add-dir")
}
