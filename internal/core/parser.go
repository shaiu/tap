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
