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
