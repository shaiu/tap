package core

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DefaultExtensions are the file extensions that the scanner looks for by default.
var DefaultExtensions = []string{".sh", ".bash", ".py"}

// DefaultIgnoreDirs are directories that the scanner skips by default.
var DefaultIgnoreDirs = []string{".git", "node_modules", "__pycache__", ".venv", "venv", "vendor"}

// DefaultMaxDepth is the default maximum directory depth for scanning.
const DefaultMaxDepth = 10

// ScannerConfig configures the directory scanner behavior.
type ScannerConfig struct {
	// Directories to scan
	Directories []string

	// File extensions to consider (default: [".sh", ".bash", ".py"])
	Extensions []string

	// Directories to skip (default: [".git", "node_modules", "__pycache__", ".venv", "vendor"])
	IgnoreDirs []string

	// Maximum directory depth (default: 10, 0 = unlimited)
	MaxDepth int

	// Explicitly registered scripts (always included)
	RegisteredScripts []string

	// AutoGenMetadata enables auto-generation of metadata for scripts without YAML front matter
	// (default: true)
	AutoGenMetadata bool
}

// Scanner discovers scripts from configured sources.
type Scanner interface {
	// Scan discovers scripts from all configured sources.
	Scan(ctx context.Context) ([]Script, error)

	// ScanDirectory scans a single directory for scripts.
	ScanDirectory(ctx context.Context, dir string) ([]Script, error)
}

// DefaultScanner implements the Scanner interface.
type DefaultScanner struct {
	config ScannerConfig
}

// NewScanner creates a new Scanner with the given configuration.
func NewScanner(config ScannerConfig) *DefaultScanner {
	// Apply defaults
	if len(config.Extensions) == 0 {
		config.Extensions = DefaultExtensions
	}
	if len(config.IgnoreDirs) == 0 {
		config.IgnoreDirs = DefaultIgnoreDirs
	}
	if config.MaxDepth == 0 {
		config.MaxDepth = DefaultMaxDepth
	}
	// AutoGenMetadata defaults to true (we need a way to detect if it was explicitly set to false)
	// Since we can't distinguish between unset and false, we default to true here
	// and require explicit disabling via config

	return &DefaultScanner{config: config}
}

// Scan discovers scripts from all configured directories.
func (s *DefaultScanner) Scan(ctx context.Context) ([]Script, error) {
	var allScripts []Script
	seen := make(map[string]bool) // Track script names to detect duplicates

	// Scan all configured directories
	for _, dir := range s.config.Directories {
		scripts, err := s.ScanDirectory(ctx, dir)
		if err != nil {
			// Check for context cancellation
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// Log warning but continue with other directories
			continue
		}

		for _, script := range scripts {
			if !seen[script.Name] {
				seen[script.Name] = true
				allScripts = append(allScripts, script)
			}
			// Skip duplicates silently (first one wins)
		}
	}

	// Add explicitly registered scripts
	for _, path := range s.config.RegisteredScripts {
		script, err := ParseScript(path)
		if err != nil || script == nil {
			continue
		}
		if !seen[script.Name] {
			seen[script.Name] = true
			script.Source = "registered"
			allScripts = append(allScripts, *script)
		}
	}

	return allScripts, nil
}

// ScanDirectory scans a single directory for scripts.
func (s *DefaultScanner) ScanDirectory(ctx context.Context, root string) ([]Script, error) {
	var scripts []Script

	// Expand ~ to home directory
	root = expandPath(root)

	// Verify directory exists
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directory doesn't exist, return empty
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil // Not a directory, return empty
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			if os.IsPermission(err) {
				return nil // Skip permission errors
			}
			return err
		}

		// Skip ignored directories
		if d.IsDir() {
			if s.shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			// Check depth
			if s.exceedsMaxDepth(root, path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check extension
		if !s.hasValidExtension(path) {
			return nil
		}

		// Try to parse metadata
		script, parseErr := ParseScript(path)
		if parseErr != nil {
			return nil // Invalid YAML - skip
		}
		if script == nil {
			// No metadata found - auto-generate if enabled
			if !s.config.AutoGenMetadata {
				return nil // Skip when disabled
			}
			script = GenerateMetadata(path, root)
		}

		script.Source = "scanned"
		scripts = append(scripts, *script)
		return nil
	})

	if err != nil {
		return scripts, err
	}

	return scripts, nil
}

// shouldSkipDir returns true if the directory should be skipped.
func (s *DefaultScanner) shouldSkipDir(name string) bool {
	for _, ignored := range s.config.IgnoreDirs {
		if name == ignored {
			return true
		}
	}
	return false
}

// exceedsMaxDepth returns true if the path exceeds the maximum depth.
func (s *DefaultScanner) exceedsMaxDepth(root, path string) bool {
	if s.config.MaxDepth <= 0 {
		return false // Unlimited depth
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	// Count path separators to determine depth
	depth := strings.Count(relPath, string(filepath.Separator))
	return depth >= s.config.MaxDepth
}

// hasValidExtension returns true if the file has a valid extension.
func (s *DefaultScanner) hasValidExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, validExt := range s.config.Extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}
