package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateMetadata_BasicFilename(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "my_script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hello"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script := GenerateMetadata(scriptPath, tmpDir)

	if script == nil {
		t.Fatal("expected script, got nil")
	}
	if script.Name != "my-script" {
		t.Errorf("Name = %q, want %q", script.Name, "my-script")
	}
	if script.Description != "(no description)" {
		t.Errorf("Description = %q, want %q", script.Description, "(no description)")
	}
	if script.Category != "uncategorized" {
		t.Errorf("Category = %q, want %q", script.Category, "uncategorized")
	}
	if script.Shell != "bash" {
		t.Errorf("Shell = %q, want %q", script.Shell, "bash")
	}
	if !script.AutoGen {
		t.Error("AutoGen should be true")
	}
}

func TestGenerateMetadata_WithCategory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "docker")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	scriptPath := filepath.Join(subDir, "wait.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho waiting"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script := GenerateMetadata(scriptPath, tmpDir)

	if script == nil {
		t.Fatal("expected script, got nil")
	}
	if script.Name != "wait" {
		t.Errorf("Name = %q, want %q", script.Name, "wait")
	}
	if script.Category != "docker" {
		t.Errorf("Category = %q, want %q", script.Category, "docker")
	}
}

func TestGenerateMetadata_DeepNesting(t *testing.T) {
	tmpDir := t.TempDir()
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	scriptPath := filepath.Join(deepDir, "deep.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho deep"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script := GenerateMetadata(scriptPath, tmpDir)

	if script == nil {
		t.Fatal("expected script, got nil")
	}
	// Should use first directory level only
	if script.Category != "level1" {
		t.Errorf("Category = %q, want %q", script.Category, "level1")
	}
}

func TestGenerateMetadata_PythonScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "data_processor.py")
	if err := os.WriteFile(scriptPath, []byte("#!/usr/bin/env python3\nprint('hello')"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script := GenerateMetadata(scriptPath, tmpDir)

	if script == nil {
		t.Fatal("expected script, got nil")
	}
	if script.Name != "data-processor" {
		t.Errorf("Name = %q, want %q", script.Name, "data-processor")
	}
	if script.Shell != "python" {
		t.Errorf("Shell = %q, want %q", script.Shell, "python")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my_script", "my-script"},
		{"My Script", "my-script"},
		{"UPPERCASE", "uppercase"},
		{"already-good", "already-good"},
		{"multi_word_name", "multi-word-name"},
		{"Mixed_Case Name", "mixed-case-name"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeriveCategory(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		relPath  string
		want     string
	}{
		{"root level", "script.sh", "uncategorized"},
		{"single level", "docker/script.sh", "docker"},
		{"deep nesting", "a/b/c/script.sh", "a"},
		{"two levels", "k8s/pods/script.sh", "k8s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tmpDir, tt.relPath)
			// Create parent directory
			if dir := filepath.Dir(fullPath); dir != tmpDir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
			}
			// Create file
			if err := os.WriteFile(fullPath, []byte("#!/bin/bash"), 0755); err != nil {
				t.Fatalf("failed to create file: %v", err)
			}

			got := deriveCategory(fullPath, tmpDir)
			if got != tt.want {
				t.Errorf("deriveCategory(%q, %q) = %q, want %q", fullPath, tmpDir, got, tt.want)
			}
		})
	}
}

func TestScanDirectory_AutoGeneratesMetadata(t *testing.T) {
	scanner := NewScanner(ScannerConfig{
		AutoGenMetadata: true,
	})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should now find no_meta.sh as an auto-generated script
	var foundNoMeta *Script
	for i := range scripts {
		if scripts[i].Name == "no-meta" {
			foundNoMeta = &scripts[i]
			break
		}
	}

	if foundNoMeta == nil {
		t.Error("expected to find auto-generated script 'no-meta'")
		return
	}

	if !foundNoMeta.AutoGen {
		t.Error("no-meta script should have AutoGen=true")
	}
	if foundNoMeta.Description != "(no description)" {
		t.Errorf("Description = %q, want %q", foundNoMeta.Description, "(no description)")
	}
}

func TestScanDirectory_DisabledAutoGen(t *testing.T) {
	scanner := NewScanner(ScannerConfig{
		AutoGenMetadata: false,
	})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT find no_meta.sh when auto-gen is disabled
	for _, s := range scripts {
		if s.Name == "no-meta" {
			t.Error("should not find 'no-meta' script when AutoGenMetadata is disabled")
		}
	}
}

func TestScanDirectory_ExplicitMetadataTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script with explicit metadata
	scriptPath := filepath.Join(tmpDir, "my_script.sh")
	content := `#!/bin/bash
# ---
# name: explicit-name
# description: Explicit description
# category: custom
# ---
echo "hello"
`
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{
		AutoGenMetadata: true,
	})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}

	script := scripts[0]
	// Should use explicit metadata, not auto-generated
	if script.Name != "explicit-name" {
		t.Errorf("Name = %q, want %q", script.Name, "explicit-name")
	}
	if script.Description != "Explicit description" {
		t.Errorf("Description = %q, want %q", script.Description, "Explicit description")
	}
	if script.Category != "custom" {
		t.Errorf("Category = %q, want %q", script.Category, "custom")
	}
	if script.AutoGen {
		t.Error("AutoGen should be false for scripts with explicit metadata")
	}
}

func TestScanDirectory_AutoGenCategoryFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script in a subdirectory without metadata
	subDir := filepath.Join(tmpDir, "docker")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	scriptPath := filepath.Join(subDir, "wait_for_db.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho waiting"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{
		AutoGenMetadata: true,
	})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}

	script := scripts[0]
	if script.Name != "wait-for-db" {
		t.Errorf("Name = %q, want %q", script.Name, "wait-for-db")
	}
	if script.Category != "docker" {
		t.Errorf("Category = %q, want %q", script.Category, "docker")
	}
	if !script.AutoGen {
		t.Error("AutoGen should be true")
	}
}
