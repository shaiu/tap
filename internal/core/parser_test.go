package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseScript_ValidBash(t *testing.T) {
	script, err := ParseScript("testdata/valid_bash.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	// Check basic fields
	if script.Name != "deploy" {
		t.Errorf("Name = %q, want %q", script.Name, "deploy")
	}
	if script.Description != "Deploy application to specified environment" {
		t.Errorf("Description = %q, want %q", script.Description, "Deploy application to specified environment")
	}
	if script.Category != "deployment" {
		t.Errorf("Category = %q, want %q", script.Category, "deployment")
	}
	if script.Author != "platform-team" {
		t.Errorf("Author = %q, want %q", script.Author, "platform-team")
	}
	if script.Version != "2.1.0" {
		t.Errorf("Version = %q, want %q", script.Version, "2.1.0")
	}

	// Check tags
	expectedTags := []string{"kubernetes", "production"}
	if len(script.Tags) != len(expectedTags) {
		t.Fatalf("Tags length = %d, want %d", len(script.Tags), len(expectedTags))
	}
	for i, tag := range expectedTags {
		if script.Tags[i] != tag {
			t.Errorf("Tags[%d] = %q, want %q", i, script.Tags[i], tag)
		}
	}

	// Check parameters
	if len(script.Parameters) != 4 {
		t.Fatalf("Parameters length = %d, want 4", len(script.Parameters))
	}

	// Check first parameter
	p := script.Parameters[0]
	if p.Name != "environment" {
		t.Errorf("Parameters[0].Name = %q, want %q", p.Name, "environment")
	}
	if p.Type != "string" {
		t.Errorf("Parameters[0].Type = %q, want %q", p.Type, "string")
	}
	if !p.Required {
		t.Error("Parameters[0].Required = false, want true")
	}
	if p.Short != "e" {
		t.Errorf("Parameters[0].Short = %q, want %q", p.Short, "e")
	}

	// Check examples
	if len(script.Examples) != 2 {
		t.Fatalf("Examples length = %d, want 2", len(script.Examples))
	}
	if script.Examples[0].Command != "deploy -e production -v v2.1.0" {
		t.Errorf("Examples[0].Command = %q, want %q", script.Examples[0].Command, "deploy -e production -v v2.1.0")
	}

	// Check runtime fields
	if script.Shell != "bash" {
		t.Errorf("Shell = %q, want %q", script.Shell, "bash")
	}
	if script.Path == "" {
		t.Error("Path should not be empty")
	}
}

func TestParseScript_ValidPython(t *testing.T) {
	script, err := ParseScript("testdata/valid_python.py")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	// Check basic fields
	if script.Name != "process-data" {
		t.Errorf("Name = %q, want %q", script.Name, "process-data")
	}
	if script.Description != "Transform and validate data files" {
		t.Errorf("Description = %q, want %q", script.Description, "Transform and validate data files")
	}
	if script.Category != "data" {
		t.Errorf("Category = %q, want %q", script.Category, "data")
	}
	if script.Author != "data-team" {
		t.Errorf("Author = %q, want %q", script.Author, "data-team")
	}
	if script.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", script.Version, "1.0.0")
	}

	// Check parameters
	if len(script.Parameters) != 2 {
		t.Fatalf("Parameters length = %d, want 2", len(script.Parameters))
	}

	// Check first parameter
	p := script.Parameters[0]
	if p.Name != "input_file" {
		t.Errorf("Parameters[0].Name = %q, want %q", p.Name, "input_file")
	}
	if p.Type != "path" {
		t.Errorf("Parameters[0].Type = %q, want %q", p.Type, "path")
	}
	if !p.Required {
		t.Error("Parameters[0].Required = false, want true")
	}

	// Check runtime fields
	if script.Shell != "python" {
		t.Errorf("Shell = %q, want %q", script.Shell, "python")
	}
}

func TestParseScript_Minimal(t *testing.T) {
	script, err := ParseScript("testdata/minimal.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	if script.Name != "hello" {
		t.Errorf("Name = %q, want %q", script.Name, "hello")
	}
	if script.Description != "Say hello world" {
		t.Errorf("Description = %q, want %q", script.Description, "Say hello world")
	}
	// Category should default to "uncategorized"
	if script.Category != "uncategorized" {
		t.Errorf("Category = %q, want %q", script.Category, "uncategorized")
	}
	// Optional fields should be empty
	if script.Author != "" {
		t.Errorf("Author = %q, want empty", script.Author)
	}
	if script.Version != "" {
		t.Errorf("Version = %q, want empty", script.Version)
	}
	if len(script.Tags) != 0 {
		t.Errorf("Tags length = %d, want 0", len(script.Tags))
	}
	if len(script.Parameters) != 0 {
		t.Errorf("Parameters length = %d, want 0", len(script.Parameters))
	}
	if len(script.Examples) != 0 {
		t.Errorf("Examples length = %d, want 0", len(script.Examples))
	}
}

func TestParseScript_NoMetadata(t *testing.T) {
	script, err := ParseScript("testdata/no_metadata.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script != nil {
		t.Errorf("expected nil script for file without metadata, got %+v", script)
	}
}

func TestParseScript_InvalidYAML(t *testing.T) {
	script, err := ParseScript("testdata/invalid_yaml.sh")
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script for invalid YAML, got %+v", script)
	}
}

func TestParseScript_MissingRequiredFields(t *testing.T) {
	// Create a temporary file with missing description
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "missing_desc.sh")
	content := `#!/bin/bash
# ---
# name: incomplete
# ---
echo "Missing description"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return nil for missing required fields (not an error)
	if script != nil {
		t.Errorf("expected nil script for missing description, got %+v", script)
	}
}

func TestParseScript_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "missing_name.sh")
	content := `#!/bin/bash
# ---
# description: Has description but no name
# ---
echo "Missing name"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script != nil {
		t.Errorf("expected nil script for missing name, got %+v", script)
	}
}

func TestParseScript_FileNotFound(t *testing.T) {
	script, err := ParseScript("testdata/nonexistent.sh")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script for nonexistent file, got %+v", script)
	}
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"script.sh", "bash"},
		{"script.bash", "bash"},
		{"script.py", "python"},
		{"script.SH", "bash"},    // case insensitive
		{"script.PY", "python"},  // case insensitive
		{"script.unknown", "sh"}, // default to sh
		{"script", "sh"},         // no extension
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := detectShell(tt.path)
			if got != tt.want {
				t.Errorf("detectShell(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestExtractYAMLLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"# name: foo", "name: foo"},
		{"#name: foo", "name: foo"},           // no space after #
		{"#   nested: value", "  nested: value"}, // preserve indentation after single space
		{"#", ""},                              // just a hash
		{"", ""},                               // empty line
		{"  # indented: comment", "indented: comment"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractYAMLLine(tt.input)
			if got != tt.want {
				t.Errorf("extractYAMLLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseScript_PythonSingleQuoteDocstring(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "single_quote.py")
	content := `#!/usr/bin/env python3
'''
---
name: single-quote
description: Uses single quote docstring
---
'''

print("Hello")
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	if script.Name != "single-quote" {
		t.Errorf("Name = %q, want %q", script.Name, "single-quote")
	}
	if script.Description != "Uses single quote docstring" {
		t.Errorf("Description = %q, want %q", script.Description, "Uses single quote docstring")
	}
}

func TestParseScript_PathIsAbsolute(t *testing.T) {
	script, err := ParseScript("testdata/minimal.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	if !filepath.IsAbs(script.Path) {
		t.Errorf("Path should be absolute, got: %q", script.Path)
	}
}

func TestParseScript_ParameterValidation_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid_type.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: foo
#     type: invalid
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for invalid parameter type, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "invalid type") {
		t.Errorf("error message should mention 'invalid type', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid_name.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: "foo bar"
#     type: string
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for invalid parameter name, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "invalid name") {
		t.Errorf("error message should mention 'invalid name', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_DuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "duplicate_name.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: foo
#     type: string
#   - name: foo
#     type: int
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for duplicate parameter name, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "duplicate name") {
		t.Errorf("error message should mention 'duplicate name', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_InvalidShortFlag(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid_short.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: foo
#     type: string
#     short: abc
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for invalid short flag, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "short flag must be a single character") {
		t.Errorf("error message should mention 'short flag must be a single character', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_DuplicateShortFlag(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "duplicate_short.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: foo
#     type: string
#     short: f
#   - name: bar
#     type: string
#     short: f
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for duplicate short flag, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "duplicate short flag") {
		t.Errorf("error message should mention 'duplicate short flag', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_DefaultNotInChoices(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "default_not_in_choices.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: env
#     type: string
#     choices: [staging, production]
#     default: development
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for default not in choices, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
	if err != nil && !contains(err.Error(), "not in choices") {
		t.Errorf("error message should mention 'not in choices', got: %v", err)
	}
}

func TestParseScript_ParameterValidation_ValidChoicesWithDefault(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid_choices.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: env
#     type: string
#     choices: [staging, production]
#     default: staging
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}
	if script.Parameters[0].Default != "staging" {
		t.Errorf("Default = %v, want %v", script.Parameters[0].Default, "staging")
	}
}

func TestParseScript_ParameterValidation_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty_name.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: ""
#     type: string
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for empty parameter name, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
}

func TestParseScript_ParameterValidation_NameStartsWithDigit(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "digit_name.sh")
	content := `#!/bin/bash
# ---
# name: test
# description: Test script
# parameters:
#   - name: 1foo
#     type: string
# ---
echo "test"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err == nil {
		t.Error("expected error for parameter name starting with digit, got nil")
	}
	if script != nil {
		t.Errorf("expected nil script, got %+v", script)
	}
}

func TestValidateParameters(t *testing.T) {
	tests := []struct {
		name    string
		params  []Parameter
		wantErr bool
	}{
		{
			name:    "empty params",
			params:  nil,
			wantErr: false,
		},
		{
			name: "valid params",
			params: []Parameter{
				{Name: "foo", Type: "string"},
				{Name: "bar", Type: "int", Short: "b"},
			},
			wantErr: false,
		},
		{
			name: "valid with underscore and hyphen",
			params: []Parameter{
				{Name: "foo_bar", Type: "string"},
				{Name: "baz-qux", Type: "int"},
			},
			wantErr: false,
		},
		{
			name: "empty type defaults ok",
			params: []Parameter{
				{Name: "foo"}, // empty type is valid (defaults to string)
			},
			wantErr: false,
		},
		{
			name: "numeric choices with default",
			params: []Parameter{
				{Name: "replicas", Type: "int", Choices: []any{1, 2, 3}, Default: 2},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameters(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"foo", true},
		{"Foo", true},
		{"foo_bar", true},
		{"foo-bar", true},
		{"foo123", true},
		{"_foo", true},
		{"FOO_BAR", true},
		{"", false},
		{"123foo", false},
		{"foo bar", false},
		{"foo.bar", false},
		{"foo@bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestChoicesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{"string equal", "foo", "foo", true},
		{"string not equal", "foo", "bar", false},
		{"int equal", 42, 42, true},
		{"int not equal", 42, 43, false},
		{"float equal", 3.14, 3.14, true},
		{"int and float equal", 42, 42.0, true},
		{"float and int equal", 42.0, 42, true},
		{"bool equal", true, true, true},
		{"bool not equal", true, false, false},
		{"string and int not equal", "42", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := choicesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("choicesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestParseScript_InteractiveTrue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "interactive.sh")
	content := `#!/bin/bash
# ---
# name: interactive-script
# description: Script with its own prompts
# category: tools
# interactive: true
# parameters:
#   - name: env
#     type: string
#     required: true
# ---
echo "Hello"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	script, err := ParseScript(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	if !script.Interactive {
		t.Error("Interactive = false, want true")
	}
	if script.Name != "interactive-script" {
		t.Errorf("Name = %q, want %q", script.Name, "interactive-script")
	}
	// Should still have parameters parsed
	if len(script.Parameters) != 1 {
		t.Fatalf("Parameters length = %d, want 1", len(script.Parameters))
	}
}

func TestParseScript_InteractiveDefaultFalse(t *testing.T) {
	// Scripts without interactive field should default to false
	script, err := ParseScript("testdata/valid_bash.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script == nil {
		t.Fatal("expected script, got nil")
	}

	if script.Interactive {
		t.Error("Interactive = true, want false (default)")
	}
}

// helper function for string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
