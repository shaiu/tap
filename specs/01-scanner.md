# Scanner Spec

> Script discovery and metadata parsing

## Overview

The scanner is responsible for finding scripts in configured directories and extracting their metadata. It's the foundation that populates tap's script registry.

## Responsibilities

1. Recursively scan configured directories for script files
2. Filter by supported extensions (.sh, .bash, .py)
3. Extract and parse YAML front matter metadata
4. Organize scripts by category
5. Cache parsed metadata for performance

## Metadata Schema

### Full Schema

```yaml
# Required fields
name: string              # Unique identifier, used in `tap run <name>`
description: string       # One-line description shown in list

# Optional fields
category: string          # Grouping for menu organization (default: "uncategorized")
author: string            # Script author
version: string           # Script version
tags: string[]            # Additional tags for searching

# Parameters (optional)
parameters:
  - name: string          # Parameter name (used as --param name=value)
    type: string          # One of: string, int, float, bool, path
    required: bool        # Whether parameter must be provided (default: false)
    default: any          # Default value if not provided
    choices: any[]        # Valid values (enables select input in TUI)
    description: string   # Help text for this parameter
    short: string         # Single-letter short flag (e.g., "e" for -e)

# Examples (optional, for help display)
examples:
  - command: string       # Example invocation
    description: string   # What this example does
```

### Minimal Valid Metadata

```bash
#!/bin/bash
# ---
# name: my-script
# description: Does something useful
# ---
```

### Complete Example

```bash
#!/bin/bash
# ---
# name: deploy
# description: Deploy application to specified environment
# category: deployment
# author: platform-team
# version: 2.1.0
# tags: [kubernetes, production]
# parameters:
#   - name: environment
#     type: string
#     required: true
#     choices: [staging, production]
#     short: e
#     description: Target deployment environment
#   - name: version
#     type: string
#     default: latest
#     short: v
#     description: Version tag to deploy
#   - name: dry_run
#     type: bool
#     default: false
#     short: d
#     description: Show what would be done without executing
#   - name: replicas
#     type: int
#     default: 3
#     description: Number of replicas to deploy
# examples:
#   - command: deploy -e production -v v2.1.0
#     description: Deploy version 2.1.0 to production
#   - command: deploy -e staging --dry_run
#     description: Preview staging deployment
# ---

set -euo pipefail
# Script implementation follows...
```

### Python Format

```python
#!/usr/bin/env python3
"""
---
name: process-data
description: Transform and validate data files
category: data
parameters:
  - name: input_file
    type: path
    required: true
    description: Input data file
  - name: format
    type: string
    default: json
    choices: [json, csv, parquet]
---
"""

import sys
# Implementation...
```

## Data Structures

### Script

```go
type Script struct {
    // Identity
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Category    string   `yaml:"category"`
    
    // Metadata
    Author      string   `yaml:"author,omitempty"`
    Version     string   `yaml:"version,omitempty"`
    Tags        []string `yaml:"tags,omitempty"`
    
    // Parameters
    Parameters  []Parameter `yaml:"parameters,omitempty"`
    
    // Examples
    Examples    []Example `yaml:"examples,omitempty"`
    
    // Runtime (not from YAML)
    Path        string   `yaml:"-"` // Absolute path to script file
    Shell       string   `yaml:"-"` // Detected shell (bash, python, etc.)
    Source      string   `yaml:"-"` // "scanned" or "registered"
}
```

### Parameter

```go
type Parameter struct {
    Name        string   `yaml:"name"`
    Type        string   `yaml:"type"`        // string, int, float, bool, path
    Required    bool     `yaml:"required"`
    Default     any      `yaml:"default,omitempty"`
    Choices     []any    `yaml:"choices,omitempty"`
    Description string   `yaml:"description,omitempty"`
    Short       string   `yaml:"short,omitempty"`
}

// ParamType constants
const (
    ParamTypeString = "string"
    ParamTypeInt    = "int"
    ParamTypeFloat  = "float"
    ParamTypeBool   = "bool"
    ParamTypePath   = "path"
)
```

### Example

```go
type Example struct {
    Command     string `yaml:"command"`
    Description string `yaml:"description"`
}
```

### Category

```go
type Category struct {
    Name    string
    Scripts []Script
}
```

## Scanner Interface

```go
type Scanner interface {
    // Scan discovers scripts from all configured sources
    Scan(ctx context.Context) ([]Script, error)
    
    // ScanDirectory scans a single directory
    ScanDirectory(ctx context.Context, dir string) ([]Script, error)
    
    // ParseScript extracts metadata from a single file
    ParseScript(path string) (*Script, error)
}

type ScannerConfig struct {
    // Directories to scan
    Directories []string
    
    // File extensions to consider
    Extensions []string  // Default: [".sh", ".bash", ".py"]
    
    // Directories to skip
    IgnoreDirs []string  // Default: [".git", "node_modules", "__pycache__", ".venv", "vendor"]
    
    // Maximum directory depth
    MaxDepth int         // Default: 10, 0 = unlimited
    
    // Explicitly registered scripts (always included)
    RegisteredScripts []string
}
```

## Implementation Details

### Directory Scanning

Use `filepath.WalkDir` (Go 1.16+) for efficient traversal:

```go
func (s *scanner) ScanDirectory(ctx context.Context, root string) ([]Script, error) {
    var scripts []Script
    root = expandPath(root) // Handle ~ expansion
    
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
        script, err := s.ParseScript(path)
        if err != nil || script == nil {
            return nil // Skip scripts without valid metadata
        }
        
        script.Source = "scanned"
        scripts = append(scripts, *script)
        return nil
    })
    
    return scripts, err
}
```

### Metadata Extraction

Parse YAML front matter from comment blocks:

```go
func (s *scanner) ParseScript(filepath string) (*Script, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // Detect comment prefix from extension
    commentPrefix := getCommentPrefix(filepath)
    
    scanner := bufio.NewScanner(file)
    var yamlContent strings.Builder
    var inYamlBlock bool
    lineNum := 0
    
    for scanner.Scan() {
        lineNum++
        line := scanner.Text()
        
        // Skip shebang
        if lineNum == 1 && strings.HasPrefix(line, "#!") {
            continue
        }
        
        // Handle Python docstrings
        if commentPrefix == "\"\"\"" {
            if strings.HasPrefix(strings.TrimSpace(line), "\"\"\"") {
                if inYamlBlock {
                    break
                }
                continue
            }
        }
        
        // Strip comment prefix
        trimmed := strings.TrimPrefix(line, commentPrefix)
        trimmed = strings.TrimSpace(trimmed)
        
        // Detect YAML delimiters
        if trimmed == "---" {
            if inYamlBlock {
                break // End of YAML block
            }
            inYamlBlock = true
            continue
        }
        
        if inYamlBlock {
            // Remove comment prefix and one space
            yamlLine := strings.TrimPrefix(line, commentPrefix)
            if strings.HasPrefix(yamlLine, " ") {
                yamlLine = yamlLine[1:]
            }
            yamlContent.WriteString(yamlLine + "\n")
        }
        
        // Stop if we've gone past reasonable header length without finding metadata
        if !inYamlBlock && lineNum > 5 {
            return nil, nil // No metadata found
        }
    }
    
    if yamlContent.Len() == 0 {
        return nil, nil // No metadata found
    }
    
    var script Script
    if err := yaml.Unmarshal([]byte(yamlContent.String()), &script); err != nil {
        return nil, fmt.Errorf("invalid YAML in %s: %w", filepath, err)
    }
    
    // Validate required fields
    if script.Name == "" || script.Description == "" {
        return nil, nil // Missing required fields
    }
    
    // Set defaults
    if script.Category == "" {
        script.Category = "uncategorized"
    }
    
    // Set runtime fields
    script.Path = filepath
    script.Shell = detectShell(filepath)
    
    return &script, nil
}

func getCommentPrefix(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".py":
        return "#" // We handle docstrings separately
    default:
        return "#"
    }
}

func detectShell(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".py":
        return "python"
    case ".sh", ".bash":
        // Could also read shebang for more accuracy
        return "bash"
    default:
        return "sh"
    }
}
```

### Category Organization

Group scripts by category after scanning:

```go
func OrganizeByCategory(scripts []Script) []Category {
    categoryMap := make(map[string][]Script)
    
    for _, script := range scripts {
        categoryMap[script.Category] = append(categoryMap[script.Category], script)
    }
    
    // Sort categories alphabetically, but put "uncategorized" last
    var categories []Category
    var names []string
    for name := range categoryMap {
        if name != "uncategorized" {
            names = append(names, name)
        }
    }
    sort.Strings(names)
    
    for _, name := range names {
        categories = append(categories, Category{
            Name:    name,
            Scripts: categoryMap[name],
        })
    }
    
    // Add uncategorized at the end if it exists
    if scripts, ok := categoryMap["uncategorized"]; ok {
        categories = append(categories, Category{
            Name:    "uncategorized",
            Scripts: scripts,
        })
    }
    
    return categories
}
```

### Caching

Cache parsed metadata to avoid rescanning unchanged files:

```go
type Cache struct {
    Entries map[string]CacheEntry `json:"entries"`
    path    string
}

type CacheEntry struct {
    ModTime   time.Time `json:"mod_time"`
    Size      int64     `json:"size"`
    Script    *Script   `json:"script"`
}

func (c *Cache) Get(path string) (*Script, bool) {
    entry, ok := c.Entries[path]
    if !ok {
        return nil, false
    }
    
    // Check if file has changed
    info, err := os.Stat(path)
    if err != nil {
        delete(c.Entries, path)
        return nil, false
    }
    
    if info.ModTime() != entry.ModTime || info.Size() != entry.Size {
        delete(c.Entries, path)
        return nil, false
    }
    
    return entry.Script, true
}

func (c *Cache) Set(path string, script *Script) {
    info, err := os.Stat(path)
    if err != nil {
        return
    }
    
    c.Entries[path] = CacheEntry{
        ModTime: info.ModTime(),
        Size:    info.Size(),
        Script:  script,
    }
}
```

## Validation Rules

1. **name** — Required, must be unique across all scripts, valid identifier (alphanumeric + hyphen/underscore)
2. **description** — Required, non-empty
3. **category** — Optional, defaults to "uncategorized"
4. **parameters[].name** — Required if parameters defined, valid identifier
5. **parameters[].type** — Must be one of: string, int, float, bool, path
6. **parameters[].choices** — If defined, default must be in choices
7. **parameters[].short** — Must be single character if defined

## Error Handling

| Error | Behavior |
|-------|----------|
| Directory doesn't exist | Log warning, skip |
| Permission denied | Log warning, skip |
| Invalid YAML | Log warning, skip script |
| Missing required fields | Skip script silently |
| Duplicate script names | Keep first found, log warning |
| File read error | Log warning, skip |

## Testing

### Unit Tests

```go
func TestParseScript_ValidBash(t *testing.T)
func TestParseScript_ValidPython(t *testing.T)
func TestParseScript_NoMetadata(t *testing.T)
func TestParseScript_InvalidYAML(t *testing.T)
func TestParseScript_MissingRequiredFields(t *testing.T)
func TestScanDirectory_Basic(t *testing.T)
func TestScanDirectory_SkipsIgnoredDirs(t *testing.T)
func TestScanDirectory_RespectsMaxDepth(t *testing.T)
func TestOrganizeByCategory(t *testing.T)
func TestCache_HitAndMiss(t *testing.T)
```

### Test Fixtures

Create a `testdata/` directory with sample scripts for testing.
