package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParamFlag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ParameterConfig
		wantErr bool
	}{
		{
			name:  "name and type only",
			input: "count:int",
			want: ParameterConfig{
				Name: "count",
				Type: "int",
			},
		},
		{
			name:  "with description",
			input: "env:string:Environment name",
			want: ParameterConfig{
				Name:        "env",
				Type:        "string",
				Description: "Environment name",
			},
		},
		{
			name:  "with default",
			input: "verbose:bool:Enable verbose mode:false",
			want: ParameterConfig{
				Name:        "verbose",
				Type:        "bool",
				Description: "Enable verbose mode",
				Default:     "false",
			},
		},
		{
			name:  "required flag",
			input: "bucket:string:S3 bucket name:required",
			want: ParameterConfig{
				Name:        "bucket",
				Type:        "string",
				Description: "S3 bucket name",
				Required:    true,
				Default:     "",
			},
		},
		{
			name:  "path type",
			input: "output:path:Output directory",
			want: ParameterConfig{
				Name:        "output",
				Type:        "path",
				Description: "Output directory",
			},
		},
		{
			name:  "float type",
			input: "threshold:float:Threshold value:0.5",
			want: ParameterConfig{
				Name:        "threshold",
				Type:        "float",
				Description: "Threshold value",
				Default:     "0.5",
			},
		},
		{
			name:    "missing type",
			input:   "count",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "count:number",
			wantErr: true,
		},
		{
			name:    "invalid name - starts with number",
			input:   "1count:int",
			wantErr: true,
		},
		{
			name:    "invalid name - contains hyphen",
			input:   "my-count:int",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParamFlag(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateScriptName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "deploy", false},
		{"valid with hyphen", "backup-db", false},
		{"valid with underscore", "backup_db", false},
		{"valid with numbers", "deploy2", false},
		{"valid mixed", "backup-db_v2", false},
		{"empty", "", true},
		{"starts with number", "2deploy", true},
		{"starts with hyphen", "-deploy", true},
		{"contains space", "my script", true},
		{"contains dot", "my.script", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScriptName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParamName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "count", false},
		{"valid with underscore", "max_count", false},
		{"valid with numbers", "count2", false},
		{"empty", "", true},
		{"starts with number", "2count", true},
		{"contains hyphen", "max-count", true},
		{"contains space", "max count", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParamName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCategoryName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "deployment", false},
		{"valid with hyphen", "dev-ops", false},
		{"valid with underscore", "dev_ops", false},
		{"valid with numbers", "deploy2", false},
		{"empty", "", true},
		{"starts with number", "2deploy", true},
		{"contains space", "dev ops", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCategoryName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSuggestOutputPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		cfg      NewScriptConfig
		contains string // partial path check since full path depends on config
	}{
		{
			name: "bash script no category",
			cfg: NewScriptConfig{
				Name:     "deploy",
				Shell:    "bash",
				Category: "",
			},
			contains: "deploy.sh",
		},
		{
			name: "python script no category",
			cfg: NewScriptConfig{
				Name:     "deploy",
				Shell:    "python",
				Category: "",
			},
			contains: "deploy.py",
		},
		{
			name: "bash script with category",
			cfg: NewScriptConfig{
				Name:     "backup",
				Shell:    "bash",
				Category: "maintenance",
			},
			contains: filepath.Join("maintenance", "backup.sh"),
		},
		{
			name: "uncategorized should not add subdir",
			cfg: NewScriptConfig{
				Name:     "test",
				Shell:    "bash",
				Category: "uncategorized",
			},
			contains: "test.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestOutputPath(tt.cfg)
			assert.Contains(t, got, tt.contains)
			// Should be under home directory
			assert.True(t, strings.HasPrefix(got, homeDir) || strings.HasPrefix(got, "/"))
		})
	}
}

func TestGenerateScript_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.sh")

	cfg := NewScriptConfig{
		Name:        "test-script",
		Description: "A test script",
		Category:    "testing",
		Shell:       "bash",
		OutputPath:  outputPath,
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	// Verify file exists and is executable
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

	// Verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "#!/bin/bash")
	assert.Contains(t, contentStr, "name: test-script")
	assert.Contains(t, contentStr, "description: A test script")
	assert.Contains(t, contentStr, "category: testing")
	assert.Contains(t, contentStr, "set -euo pipefail")
}

func TestGenerateScript_Python(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.py")

	cfg := NewScriptConfig{
		Name:        "test-script",
		Description: "A test script",
		Category:    "testing",
		Shell:       "python",
		OutputPath:  outputPath,
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	// Verify file exists and is executable
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

	// Verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "#!/usr/bin/env python3")
	assert.Contains(t, contentStr, "name: test-script")
	assert.Contains(t, contentStr, "description: A test script")
	assert.Contains(t, contentStr, "category: testing")
	assert.Contains(t, contentStr, "def main():")
}

func TestGenerateScript_WithParams(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "params.sh")

	cfg := NewScriptConfig{
		Name:        "param-script",
		Description: "Script with parameters",
		Shell:       "bash",
		OutputPath:  outputPath,
		Parameters: []ParameterConfig{
			{
				Name:        "env",
				Type:        "string",
				Description: "Environment name",
				Required:    true,
			},
			{
				Name:        "count",
				Type:        "int",
				Description: "Number of items",
				Default:     "10",
			},
		},
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Check parameters are in metadata
	assert.Contains(t, contentStr, "- name: env")
	assert.Contains(t, contentStr, "type: string")
	assert.Contains(t, contentStr, "required: true")
	assert.Contains(t, contentStr, "- name: count")
	assert.Contains(t, contentStr, "type: int")
	assert.Contains(t, contentStr, "default: 10")

	// Check env var fallback patterns
	assert.Contains(t, contentStr, "${TAP_PARAM_ENV:-}")
	assert.Contains(t, contentStr, "${TAP_PARAM_COUNT:-10}")
}

func TestGenerateScript_ExistingFile_Headless(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "existing.sh")

	// Create existing file
	err := os.WriteFile(outputPath, []byte("existing"), 0644)
	require.NoError(t, err)

	cfg := NewScriptConfig{
		Name:        "existing",
		Description: "Test",
		Shell:       "bash",
		OutputPath:  outputPath,
	}

	// Should fail in headless mode
	err = generateScript(cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file already exists")
}

func TestGenerateScript_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "test.sh")

	cfg := NewScriptConfig{
		Name:        "nested-script",
		Description: "Test nested creation",
		Shell:       "bash",
		OutputPath:  nestedPath,
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(nestedPath)
	assert.NoError(t, err)
}

func TestRunNewHeadless_MissingName(t *testing.T) {
	// Test that headless mode requires a name
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "Test", "")
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("shell", "bash", "")
	cmd.Flags().StringSlice("param", nil, "")

	err := runNewHeadless(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name required")
}

func TestRunNewHeadless_MissingDescription(t *testing.T) {
	// Test that headless mode requires a description
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("shell", "bash", "")
	cmd.Flags().StringSlice("param", nil, "")

	err := runNewHeadless(cmd, []string{"test-script"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--description required")
}

func TestRunNewHeadless_InvalidShell(t *testing.T) {
	// Test that headless mode validates shell type
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "Test", "")
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("shell", "ruby", "")
	cmd.Flags().StringSlice("param", nil, "")

	err := runNewHeadless(cmd, []string{"test-script"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid shell")
}

func TestRunNewHeadless_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "success.sh")

	cmd := &cobra.Command{}
	cmd.Flags().String("description", "A successful test", "")
	cmd.Flags().String("category", "testing", "")
	cmd.Flags().String("output", outputPath, "")
	cmd.Flags().String("shell", "bash", "")
	cmd.Flags().StringSlice("param", nil, "")

	err := runNewHeadless(cmd, []string{"success-script"})
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: success-script")
	assert.Contains(t, string(content), "description: A successful test")
	assert.Contains(t, string(content), "category: testing")
}

func TestRunNewHeadless_WithParams(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "with-params.sh")

	cmd := &cobra.Command{}
	cmd.Flags().String("description", "Script with params", "")
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("output", outputPath, "")
	cmd.Flags().String("shell", "bash", "")
	cmd.Flags().StringSlice("param", []string{
		"name:string:Your name:required",
		"count:int:Number of items:5",
	}, "")

	err := runNewHeadless(cmd, []string{"param-script"})
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "- name: name")
	assert.Contains(t, contentStr, "required: true")
	assert.Contains(t, contentStr, "- name: count")
	assert.Contains(t, contentStr, "default: 5")
}

func TestGenerateScript_BashWithFallbackPattern(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "fallback.sh")

	cfg := NewScriptConfig{
		Name:        "fallback-test",
		Description: "Test fallback patterns",
		Shell:       "bash",
		OutputPath:  outputPath,
		Parameters: []ParameterConfig{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:    "count",
				Type:    "int",
				Default: "10",
			},
			{
				Name:    "verbose",
				Type:    "bool",
				Default: "false",
			},
		},
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Check fallback patterns with ${VAR:-default}
	assert.Contains(t, contentStr, `NAME="${TAP_PARAM_NAME:-}"`, "Required param without default should have empty fallback")
	assert.Contains(t, contentStr, `COUNT="${TAP_PARAM_COUNT:-10}"`, "Optional param should have default in fallback")
	assert.Contains(t, contentStr, `VERBOSE="${TAP_PARAM_VERBOSE:-false}"`, "Bool param should have default in fallback")

	// Check validation for required params without defaults
	assert.Contains(t, contentStr, `if [[ -z "$NAME" ]]; then`)
	assert.Contains(t, contentStr, `echo "Error: name is required"`)

	// Count should NOT have validation (has default)
	assert.NotContains(t, contentStr, `if [[ -z "$COUNT" ]]; then`)
}

func TestGenerateScript_PythonWithFallbackPattern(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "fallback.py")

	cfg := NewScriptConfig{
		Name:        "fallback-test",
		Description: "Test Python fallback patterns",
		Shell:       "python",
		OutputPath:  outputPath,
		Parameters: []ParameterConfig{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:    "count",
				Type:    "int",
				Default: "10",
			},
			{
				Name:    "debug",
				Type:    "bool",
				Default: "false",
			},
			{
				Name:    "threshold",
				Type:    "float",
				Default: "0.5",
			},
		},
	}

	err := generateScript(cfg, true)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Check typed fallback patterns
	assert.Contains(t, contentStr, `name = os.environ.get("TAP_PARAM_NAME", "")`, "String param should use os.environ.get")
	assert.Contains(t, contentStr, `count = int(os.environ.get("TAP_PARAM_COUNT", "10"))`, "Int param should wrap with int()")
	assert.Contains(t, contentStr, `debug = os.environ.get("TAP_PARAM_DEBUG", "false").lower() in ("true", "1", "yes")`, "Bool param should use .lower() check")
	assert.Contains(t, contentStr, `threshold = float(os.environ.get("TAP_PARAM_THRESHOLD", "0.5"))`, "Float param should wrap with float()")

	// Check validation for required params without defaults
	assert.Contains(t, contentStr, `if not name:`)
	assert.Contains(t, contentStr, `print("Error: name is required", file=sys.stderr)`)

	// Count should NOT have validation (has default)
	assert.NotContains(t, contentStr, `if not count:`)
}

func TestBuildCategoryOptions(t *testing.T) {
	existing := []string{"deployment", "maintenance"}
	opts := buildCategoryOptions(existing)

	assert.Len(t, opts, 4) // uncategorized + 2 existing + new category

	// Check order: uncategorized first, then existing, then new
	assert.Equal(t, "uncategorized", opts[0].Value)
	assert.Equal(t, "deployment", opts[1].Value)
	assert.Equal(t, "maintenance", opts[2].Value)
	assert.Equal(t, "__new__", opts[3].Value)
}

func TestBuildCategoryOptions_Empty(t *testing.T) {
	opts := buildCategoryOptions(nil)

	assert.Len(t, opts, 2) // uncategorized + new category
	assert.Equal(t, "uncategorized", opts[0].Value)
	assert.Equal(t, "__new__", opts[1].Value)
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "expand tilde path",
			input: "~/scripts/test.sh",
			want:  filepath.Join(homeDir, "scripts/test.sh"),
		},
		{
			name:  "expand tilde only",
			input: "~",
			want:  homeDir,
		},
		{
			name:  "no expansion needed",
			input: "/tmp/test.sh",
			want:  "/tmp/test.sh",
		},
		{
			name:  "relative path unchanged",
			input: "scripts/test.sh",
			want:  "scripts/test.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
