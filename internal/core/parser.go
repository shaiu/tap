package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseScript extracts metadata from a script file's YAML front matter.
// It returns nil, nil if the file has no valid metadata (not an error).
// It returns nil, error if the file has malformed YAML.
func ParseScript(path string) (*Script, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	yamlContent, err := extractYAML(file, path)
	if err != nil {
		return nil, err
	}
	if yamlContent == "" {
		return nil, nil // No metadata found
	}

	var script Script
	if err := yaml.Unmarshal([]byte(yamlContent), &script); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}

	// Validate required fields
	if script.Name == "" || script.Description == "" {
		return nil, nil // Missing required fields, skip silently
	}

	// Set defaults
	if script.Category == "" {
		script.Category = "uncategorized"
	}

	// Validate parameters
	if err := validateParameters(script.Parameters); err != nil {
		return nil, fmt.Errorf("invalid parameter in %s: %w", path, err)
	}

	// Set runtime fields
	script.Path = absPath
	script.Shell = detectShell(path)

	return &script, nil
}

// extractYAML reads the YAML front matter from a script file.
// Returns empty string if no metadata found.
func extractYAML(file *os.File, path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".py" {
		return extractYAMLFromPython(file)
	}
	return extractYAMLFromBash(file)
}

// extractYAMLFromBash extracts YAML from bash/shell scripts using # comments.
// Format:
//
//	#!/bin/bash
//	# ---
//	# name: foo
//	# description: bar
//	# ---
func extractYAMLFromBash(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	var yamlContent strings.Builder
	var inYamlBlock bool
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip shebang on first line
		if lineNum == 1 && strings.HasPrefix(line, "#!") {
			continue
		}

		// Check for YAML delimiter
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Remove the # prefix and check for ---
			content := strings.TrimPrefix(trimmed, "#")
			content = strings.TrimSpace(content)

			if content == "---" {
				if inYamlBlock {
					break // End of YAML block
				}
				inYamlBlock = true
				continue
			}

			if inYamlBlock {
				// Extract the content after the # prefix
				// We need to preserve indentation for YAML parsing
				yamlLine := extractYAMLLine(line)
				yamlContent.WriteString(yamlLine + "\n")
			}
		} else if !inYamlBlock && lineNum > 5 {
			// Stop if we've gone past reasonable header length without finding metadata
			return "", nil
		} else if inYamlBlock && !strings.HasPrefix(trimmed, "#") {
			// Non-comment line while in YAML block means end of metadata
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return yamlContent.String(), nil
}

// extractYAMLLine extracts the YAML content from a comment line.
// Input: "# name: foo" -> "name: foo"
// Input: "#   nested: value" -> "  nested: value"
func extractYAMLLine(line string) string {
	// Find the # character
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}

	// Get everything after the #
	rest := line[idx+1:]

	// Remove exactly one space after # if present (for standard formatting)
	if strings.HasPrefix(rest, " ") {
		return rest[1:]
	}
	return rest
}

// extractYAMLFromPython extracts YAML from Python scripts using docstrings.
// Format:
//
//	#!/usr/bin/env python3
//	"""
//	---
//	name: foo
//	description: bar
//	---
//	"""
func extractYAMLFromPython(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	var yamlContent strings.Builder
	var inDocstring bool
	var inYamlBlock bool
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip shebang on first line
		if lineNum == 1 && strings.HasPrefix(line, "#!") {
			continue
		}

		// Look for docstring start
		if !inDocstring {
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				inDocstring = true
				// Check if --- is on the same line after the docstring opener
				afterQuotes := strings.TrimPrefix(trimmed, `"""`)
				afterQuotes = strings.TrimPrefix(afterQuotes, `'''`)
				afterQuotes = strings.TrimSpace(afterQuotes)
				if afterQuotes == "---" {
					inYamlBlock = true
				}
				continue
			}
			// Stop if we've gone past reasonable header length without finding metadata
			if lineNum > 5 {
				return "", nil
			}
			continue
		}

		// We're inside the docstring
		// Check for docstring end
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
			break // End of docstring
		}

		// Check for YAML delimiters
		if trimmed == "---" {
			if inYamlBlock {
				break // End of YAML block
			}
			inYamlBlock = true
			continue
		}

		if inYamlBlock {
			yamlContent.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	return yamlContent.String(), nil
}

// detectShell determines the shell/interpreter type from the file extension.
func detectShell(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py":
		return "python"
	case ".sh", ".bash":
		return "bash"
	default:
		return "sh"
	}
}

// GenerateMetadata creates a Script with auto-generated metadata from the filepath.
// The root parameter is the scan directory root, used to derive the category.
func GenerateMetadata(path string, root string) *Script {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Extract name from filename (without extension)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	name = sanitizeName(name)

	// Derive category from parent directory relative to root
	category := deriveCategory(path, root)

	return &Script{
		Name:        name,
		Description: "(no description)",
		Category:    category,
		Path:        absPath,
		Shell:       detectShell(path),
		AutoGen:     true,
	}
}

// sanitizeName converts a filename to a valid script name.
// Converts to lowercase and replaces underscores and spaces with hyphens.
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

// deriveCategory extracts the category from the path relative to the scan root.
// Uses the first directory level, or "uncategorized" if at root level.
func deriveCategory(path string, root string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "uncategorized"
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "uncategorized"
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "uncategorized"
	}

	// Get the directory part of the relative path
	dir := filepath.Dir(relPath)
	if dir == "." {
		return "uncategorized" // Script is at root level
	}

	// Split the path and take the first directory
	parts := strings.Split(dir, string(filepath.Separator))
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return "uncategorized"
}

// validateParameters validates all parameters in a script.
// Returns an error if any parameter is invalid.
func validateParameters(params []Parameter) error {
	seen := make(map[string]bool)
	seenShorts := make(map[string]bool)

	for i, p := range params {
		// Validate parameter name is non-empty
		if p.Name == "" {
			return fmt.Errorf("parameter %d: name is required", i+1)
		}

		// Validate parameter name is a valid identifier
		if !isValidIdentifier(p.Name) {
			return fmt.Errorf("parameter %q: invalid name (must be alphanumeric with hyphens/underscores)", p.Name)
		}

		// Check for duplicate names
		if seen[p.Name] {
			return fmt.Errorf("parameter %q: duplicate name", p.Name)
		}
		seen[p.Name] = true

		// Validate type (default to string if not specified)
		if p.Type != "" && !IsValidParamType(p.Type) {
			return fmt.Errorf("parameter %q: invalid type %q (must be one of: string, int, float, bool, path)", p.Name, p.Type)
		}

		// Validate short flag is a single character
		if p.Short != "" {
			if len(p.Short) != 1 {
				return fmt.Errorf("parameter %q: short flag must be a single character, got %q", p.Name, p.Short)
			}
			if seenShorts[p.Short] {
				return fmt.Errorf("parameter %q: duplicate short flag %q", p.Name, p.Short)
			}
			seenShorts[p.Short] = true
		}

		// Validate default is in choices if choices are defined
		if len(p.Choices) > 0 && p.Default != nil {
			if !containsChoice(p.Choices, p.Default) {
				return fmt.Errorf("parameter %q: default value %v is not in choices %v", p.Name, p.Default, p.Choices)
			}
		}
	}

	return nil
}

// isValidIdentifier checks if a string is a valid identifier (alphanumeric with hyphens/underscores).
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && r >= '0' && r <= '9' {
			return false // Cannot start with a digit
		}
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// containsChoice checks if a value is in the choices slice.
func containsChoice(choices []any, value any) bool {
	for _, c := range choices {
		if choicesEqual(c, value) {
			return true
		}
	}
	return false
}

// choicesEqual compares two values for equality, handling type coercion for common cases.
func choicesEqual(a, b any) bool {
	// Direct comparison
	if a == b {
		return true
	}

	// Handle numeric type coercion (YAML may parse as int or float)
	switch av := a.(type) {
	case int:
		switch bv := b.(type) {
		case int:
			return av == bv
		case float64:
			return float64(av) == bv
		}
	case float64:
		switch bv := b.(type) {
		case int:
			return av == float64(bv)
		case float64:
			return av == bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}

	return false
}
