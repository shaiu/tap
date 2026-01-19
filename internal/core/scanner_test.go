package core

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestScanDirectory_Basic(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find: deploy.sh, rollback.sh, process.py, hello.sh
	// Should NOT find: no_meta.sh (no metadata), hidden.sh (in .git)
	if len(scripts) < 4 {
		t.Errorf("expected at least 4 scripts, got %d", len(scripts))
	}

	// Check that expected scripts were found
	names := make(map[string]bool)
	for _, s := range scripts {
		names[s.Name] = true
	}

	expected := []string{"deploy", "rollback", "process", "hello"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected to find script %q", name)
		}
	}

	// Check that hidden script was NOT found
	if names["hidden"] {
		t.Error("should not find script in .git directory")
	}
}

func TestScanDirectory_SkipsIgnoredDirs(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no script from .git directory was found
	for _, s := range scripts {
		if s.Name == "hidden" {
			t.Error("should not find script in .git directory")
		}
	}
}

func TestScanDirectory_RespectsMaxDepth(t *testing.T) {
	// Scanner with depth limit of 2 should not find deep_script.sh
	scanner := NewScanner(ScannerConfig{
		MaxDepth: 2, // root + 1 level
	})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range scripts {
		if s.Name == "deep-script" {
			t.Error("should not find deeply nested script with MaxDepth=2")
		}
	}

	// Scanner with higher depth limit should find it
	scanner2 := NewScanner(ScannerConfig{
		MaxDepth: 10,
	})

	scripts2, err := scanner2.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, s := range scripts2 {
		if s.Name == "deep-script" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should find deeply nested script with MaxDepth=10")
	}
}

func TestScanDirectory_FiltersByExtension(t *testing.T) {
	// Create a temp directory with various file types
	tmpDir := t.TempDir()

	// Create a .sh file with metadata
	shFile := filepath.Join(tmpDir, "valid.sh")
	shContent := `#!/bin/bash
# ---
# name: valid-sh
# description: Valid bash script
# ---
echo "hello"
`
	if err := os.WriteFile(shFile, []byte(shContent), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a .txt file with metadata (should be ignored)
	txtFile := filepath.Join(tmpDir, "invalid.txt")
	txtContent := `# ---
# name: invalid-txt
# description: Should be ignored
# ---
Some text
`
	if err := os.WriteFile(txtFile, []byte(txtContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("expected 1 script, got %d", len(scripts))
	}
	if len(scripts) > 0 && scripts[0].Name != "valid-sh" {
		t.Errorf("expected script name 'valid-sh', got %q", scripts[0].Name)
	}
}

func TestScanDirectory_NonExistentDir(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "/nonexistent/directory/path")
	if err != nil {
		t.Fatalf("unexpected error for nonexistent dir: %v", err)
	}
	if scripts != nil && len(scripts) != 0 {
		t.Errorf("expected empty slice for nonexistent dir, got %d scripts", len(scripts))
	}
}

func TestScanDirectory_SetsScanSource(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range scripts {
		if s.Source != "scanned" {
			t.Errorf("expected Source='scanned', got %q for script %q", s.Source, s.Name)
		}
	}
}

func TestScan_MultipleDirectories(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create scripts in both directories
	script1 := filepath.Join(tmpDir1, "script1.sh")
	script1Content := `#!/bin/bash
# ---
# name: script1
# description: First script
# ---
`
	if err := os.WriteFile(script1, []byte(script1Content), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script2 := filepath.Join(tmpDir2, "script2.sh")
	script2Content := `#!/bin/bash
# ---
# name: script2
# description: Second script
# ---
`
	if err := os.WriteFile(script2, []byte(script2Content), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{
		Directories: []string{tmpDir1, tmpDir2},
	})
	ctx := context.Background()

	scripts, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 2 {
		t.Errorf("expected 2 scripts, got %d", len(scripts))
	}

	names := make(map[string]bool)
	for _, s := range scripts {
		names[s.Name] = true
	}

	if !names["script1"] || !names["script2"] {
		t.Error("expected to find both script1 and script2")
	}
}

func TestScan_DuplicateNames(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create scripts with same name in both directories
	script1 := filepath.Join(tmpDir1, "script.sh")
	script1Content := `#!/bin/bash
# ---
# name: duplicate
# description: First duplicate
# ---
`
	if err := os.WriteFile(script1, []byte(script1Content), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	script2 := filepath.Join(tmpDir2, "script.sh")
	script2Content := `#!/bin/bash
# ---
# name: duplicate
# description: Second duplicate
# ---
`
	if err := os.WriteFile(script2, []byte(script2Content), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{
		Directories: []string{tmpDir1, tmpDir2},
	})
	ctx := context.Background()

	scripts, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have one script (first one wins)
	if len(scripts) != 1 {
		t.Errorf("expected 1 script (duplicates deduplicated), got %d", len(scripts))
	}

	if len(scripts) > 0 && scripts[0].Name != "duplicate" {
		t.Errorf("expected script name 'duplicate', got %q", scripts[0].Name)
	}
}

func TestScan_RegisteredScripts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a registered script outside scan directories
	script := filepath.Join(tmpDir, "registered.sh")
	scriptContent := `#!/bin/bash
# ---
# name: registered
# description: Explicitly registered script
# ---
`
	if err := os.WriteFile(script, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewScanner(ScannerConfig{
		Directories:       []string{}, // No scan directories
		RegisteredScripts: []string{script},
	})
	ctx := context.Background()

	scripts, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("expected 1 script, got %d", len(scripts))
	}

	if len(scripts) > 0 {
		if scripts[0].Name != "registered" {
			t.Errorf("expected script name 'registered', got %q", scripts[0].Name)
		}
		if scripts[0].Source != "registered" {
			t.Errorf("expected Source='registered', got %q", scripts[0].Source)
		}
	}
}

func TestScanDirectory_ContextCancellation(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestScanDirectory_ContextTimeout(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give the timeout a moment to expire
	time.Sleep(1 * time.Millisecond)

	_, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded error, got %v", err)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not get home directory")
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/scripts", filepath.Join(home, "scripts")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewScanner_AppliesDefaults(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})

	// Check that defaults were applied
	if len(scanner.config.Extensions) != len(DefaultExtensions) {
		t.Errorf("expected %d default extensions, got %d", len(DefaultExtensions), len(scanner.config.Extensions))
	}

	if len(scanner.config.IgnoreDirs) != len(DefaultIgnoreDirs) {
		t.Errorf("expected %d default ignore dirs, got %d", len(DefaultIgnoreDirs), len(scanner.config.IgnoreDirs))
	}

	if scanner.config.MaxDepth != DefaultMaxDepth {
		t.Errorf("expected MaxDepth=%d, got %d", DefaultMaxDepth, scanner.config.MaxDepth)
	}
}

func TestOrganizeByCategory_Integration(t *testing.T) {
	scanner := NewScanner(ScannerConfig{})
	ctx := context.Background()

	scripts, err := scanner.ScanDirectory(ctx, "testdata/scripts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	categories := OrganizeByCategory(scripts)

	// Should have at least deployment, data, and uncategorized
	categoryNames := make([]string, len(categories))
	for i, c := range categories {
		categoryNames[i] = c.Name
	}
	sort.Strings(categoryNames[:len(categoryNames)-1]) // Sort all but possibly uncategorized at end

	hasDeployment := false
	hasData := false
	hasUncategorized := false

	for _, c := range categories {
		switch c.Name {
		case "deployment":
			hasDeployment = true
			if len(c.Scripts) != 2 {
				t.Errorf("expected 2 deployment scripts, got %d", len(c.Scripts))
			}
		case "data":
			hasData = true
			if len(c.Scripts) != 1 {
				t.Errorf("expected 1 data script, got %d", len(c.Scripts))
			}
		case "uncategorized":
			hasUncategorized = true
			// hello.sh has no category, should default to uncategorized
		}
	}

	if !hasDeployment {
		t.Error("expected 'deployment' category")
	}
	if !hasData {
		t.Error("expected 'data' category")
	}
	if !hasUncategorized {
		t.Error("expected 'uncategorized' category")
	}
}
