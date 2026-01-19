# Scaffolding Spec

> Script generation with `tap new`

## Overview

The `tap new` command creates new scripts with proper metadata boilerplate, helping users get started quickly without memorizing the metadata format.

## User Flow

### Interactive Mode

```
$ tap new

  Create new script

  Name: backup-db█
  
  Description: Backup database to S3█

  Category: 
  ┌──────────────────────────────┐
  │ > maintenance                │
  │   data                       │
  │   deployment                 │
  │   + Create new category      │
  └──────────────────────────────┘

  Parameters:
  ┌──────────────────────────────┐
  │ > Add parameter              │
  │   Done (no parameters)       │
  └──────────────────────────────┘

  Save to: ~/scripts/maintenance/backup-db.sh█

  [Create]  [Cancel]
```

### Flag Mode

```bash
tap new backup-db \
  --description "Backup database to S3" \
  --category maintenance \
  --param "bucket:string:S3 bucket name:required" \
  --param "compress:bool:Enable compression:false" \
  --output ~/scripts/maintenance/backup-db.sh
```

## Output

Generated script with full metadata:

```bash
#!/bin/bash
# ---
# name: backup-db
# description: Backup database to S3
# category: maintenance
# parameters:
#   - name: bucket
#     type: string
#     required: true
#     description: S3 bucket name
#   - name: compress
#     type: bool
#     default: false
#     description: Enable compression
# ---

set -euo pipefail

# Access parameters via environment variables:
#   $TAP_PARAM_BUCKET
#   $TAP_PARAM_COMPRESS

main() {
    echo "Running backup-db"
    
    # Your implementation here
}

main "$@"
```

## Command Definition

```go
var newCmd = &cobra.Command{
    Use:   "new [name]",
    Short: "Create a new script",
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        mode := determineMode(cmd)
        
        if mode == ModeInteractive {
            return runNewInteractive(args)
        }
        
        return runNewHeadless(cmd, args)
    },
}

func init() {
    newCmd.Flags().StringP("description", "d", "", "Script description")
    newCmd.Flags().StringP("category", "c", "", "Script category")
    newCmd.Flags().StringP("output", "o", "", "Output path")
    newCmd.Flags().StringSliceP("param", "p", nil, "Parameter (name:type:description:default)")
    newCmd.Flags().String("shell", "bash", "Shell type (bash, python)")
    rootCmd.AddCommand(newCmd)
}
```

## Data Structures

### NewScriptConfig

```go
type NewScriptConfig struct {
    Name        string
    Description string
    Category    string
    Parameters  []ParameterConfig
    Shell       string      // bash, python
    OutputPath  string
}

type ParameterConfig struct {
    Name        string
    Type        string
    Description string
    Required    bool
    Default     string
    Choices     []string
}
```

## Interactive Form

Using `huh` for the form:

```go
func runNewInteractive(args []string) error {
    cfg := NewScriptConfig{
        Shell: "bash",
    }
    
    // Pre-fill name if provided
    if len(args) > 0 {
        cfg.Name = args[0]
    }
    
    // Load existing categories for selection
    categories := loadExistingCategories()
    
    // Build form
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Key("name").
                Title("Name").
                Description("Script identifier (used in 'tap run <n>')").
                Value(&cfg.Name).
                Validate(validateScriptName),
                
            huh.NewInput().
                Key("description").
                Title("Description").
                Description("One-line description").
                Value(&cfg.Description).
                Validate(validateRequired),
                
            huh.NewSelect[string]().
                Key("category").
                Title("Category").
                Options(buildCategoryOptions(categories)...).
                Value(&cfg.Category),
        ),
    )
    
    if err := form.Run(); err != nil {
        return err
    }
    
    // Handle "new category" selection
    if cfg.Category == "__new__" {
        cfg.Category = promptNewCategory()
    }
    
    // Parameter wizard
    if shouldAddParameters() {
        cfg.Parameters = runParameterWizard()
    }
    
    // Determine output path
    if cfg.OutputPath == "" {
        cfg.OutputPath = suggestOutputPath(cfg)
    }
    cfg.OutputPath = confirmOutputPath(cfg.OutputPath)
    
    // Generate and save
    return generateScript(cfg)
}

func buildCategoryOptions(existing []string) []huh.Option[string] {
    opts := make([]huh.Option[string], 0, len(existing)+2)
    
    opts = append(opts, huh.NewOption("uncategorized", "uncategorized"))
    
    for _, cat := range existing {
        if cat != "uncategorized" {
            opts = append(opts, huh.NewOption(cat, cat))
        }
    }
    
    opts = append(opts, huh.NewOption("+ Create new category", "__new__"))
    
    return opts
}
```

### Parameter Wizard

```go
func runParameterWizard() []ParameterConfig {
    var params []ParameterConfig
    
    for {
        action := promptParameterAction(len(params))
        
        switch action {
        case "add":
            param := promptParameter()
            params = append(params, param)
        case "done":
            return params
        case "remove":
            params = promptRemoveParameter(params)
        }
    }
}

func promptParameter() ParameterConfig {
    var param ParameterConfig
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Key("name").
                Title("Parameter name").
                Value(&param.Name).
                Validate(validateParamName),
                
            huh.NewSelect[string]().
                Key("type").
                Title("Type").
                Options(
                    huh.NewOption("string", "string"),
                    huh.NewOption("int", "int"),
                    huh.NewOption("bool", "bool"),
                    huh.NewOption("path", "path"),
                ).
                Value(&param.Type),
                
            huh.NewInput().
                Key("description").
                Title("Description").
                Value(&param.Description),
                
            huh.NewConfirm().
                Key("required").
                Title("Required?").
                Value(&param.Required),
                
            huh.NewInput().
                Key("default").
                Title("Default value (optional)").
                Value(&param.Default),
        ),
    )
    
    form.Run()
    return param
}
```

## Headless Mode

```go
func runNewHeadless(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("name required in headless mode")
    }
    
    cfg := NewScriptConfig{
        Name: args[0],
    }
    
    // Get flags
    cfg.Description, _ = cmd.Flags().GetString("description")
    cfg.Category, _ = cmd.Flags().GetString("category")
    cfg.OutputPath, _ = cmd.Flags().GetString("output")
    cfg.Shell, _ = cmd.Flags().GetString("shell")
    
    // Parse parameter flags
    paramStrs, _ := cmd.Flags().GetStringSlice("param")
    for _, p := range paramStrs {
        param, err := parseParamFlag(p)
        if err != nil {
            return fmt.Errorf("invalid param %q: %w", p, err)
        }
        cfg.Parameters = append(cfg.Parameters, param)
    }
    
    // Validate
    if cfg.Description == "" {
        return fmt.Errorf("--description required")
    }
    
    // Default output path
    if cfg.OutputPath == "" {
        cfg.OutputPath = suggestOutputPath(cfg)
    }
    
    return generateScript(cfg)
}

// Format: name:type:description:default
// Examples:
//   "env:string:Environment name:staging"
//   "count:int:Number of items"
//   "verbose:bool:Enable verbose mode:false"
func parseParamFlag(s string) (ParameterConfig, error) {
    parts := strings.SplitN(s, ":", 4)
    if len(parts) < 2 {
        return ParameterConfig{}, fmt.Errorf("expected name:type[:description[:default]]")
    }
    
    param := ParameterConfig{
        Name: parts[0],
        Type: parts[1],
    }
    
    if len(parts) >= 3 {
        param.Description = parts[2]
    }
    
    if len(parts) >= 4 {
        param.Default = parts[3]
    }
    
    // "required" as default means required=true
    if param.Default == "required" {
        param.Required = true
        param.Default = ""
    }
    
    return param, nil
}
```

## Template System

### Template Structure

```go
//go:embed templates/*.tmpl
var templateFS embed.FS

type TemplateData struct {
    Name        string
    Description string
    Category    string
    Parameters  []ParameterConfig
    ParamEnvVars []string  // Pre-computed env var names
}
```

### Bash Template

```
#!/bin/bash
# ---
# name: {{.Name}}
# description: {{.Description}}
{{- if .Category}}
# category: {{.Category}}
{{- end}}
{{- if .Parameters}}
# parameters:
{{- range .Parameters}}
#   - name: {{.Name}}
#     type: {{.Type}}
{{- if .Required}}
#     required: true
{{- end}}
{{- if .Default}}
#     default: {{.Default}}
{{- end}}
{{- if .Description}}
#     description: {{.Description}}
{{- end}}
{{- end}}
{{- end}}
# ---

set -euo pipefail

# Access parameters via environment variables:
{{- range .ParamEnvVars}}
#   ${{.}}
{{- end}}

main() {
    echo "Running {{.Name}}"
    
    # Your implementation here
}

main "$@"
```

### Python Template

```
#!/usr/bin/env python3
"""
---
name: {{.Name}}
description: {{.Description}}
{{- if .Category}}
category: {{.Category}}
{{- end}}
{{- if .Parameters}}
parameters:
{{- range .Parameters}}
  - name: {{.Name}}
    type: {{.Type}}
{{- if .Required}}
    required: true
{{- end}}
{{- if .Default}}
    default: {{.Default}}
{{- end}}
{{- if .Description}}
    description: {{.Description}}
{{- end}}
{{- end}}
{{- end}}
---
"""

import os
import sys

# Access parameters via environment variables:
{{- range .ParamEnvVars}}
# {{.}} = os.environ.get("{{.}}")
{{- end}}

def main():
    print(f"Running {{.Name}}")
    
    # Your implementation here

if __name__ == "__main__":
    main()
```

## Script Generation

```go
func generateScript(cfg NewScriptConfig) error {
    // Prepare template data
    data := TemplateData{
        Name:        cfg.Name,
        Description: cfg.Description,
        Category:    cfg.Category,
        Parameters:  cfg.Parameters,
    }
    
    // Pre-compute env var names
    for _, p := range cfg.Parameters {
        data.ParamEnvVars = append(data.ParamEnvVars, 
            fmt.Sprintf("TAP_PARAM_%s", strings.ToUpper(p.Name)))
    }
    
    // Load and execute template
    tmplName := cfg.Shell + ".tmpl"
    tmpl, err := template.ParseFS(templateFS, "templates/"+tmplName)
    if err != nil {
        return fmt.Errorf("template not found for shell %s", cfg.Shell)
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return fmt.Errorf("template error: %w", err)
    }
    
    // Ensure directory exists
    dir := filepath.Dir(cfg.OutputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("cannot create directory: %w", err)
    }
    
    // Check for existing file
    if _, err := os.Stat(cfg.OutputPath); err == nil {
        if !confirmOverwrite(cfg.OutputPath) {
            return fmt.Errorf("cancelled")
        }
    }
    
    // Write file
    if err := os.WriteFile(cfg.OutputPath, buf.Bytes(), 0755); err != nil {
        return err
    }
    
    fmt.Printf("✓ Created: %s\n", cfg.OutputPath)
    fmt.Printf("\nEdit your script, then run it with:\n")
    fmt.Printf("  tap run %s\n", cfg.Name)
    
    return nil
}
```

## Output Path Suggestions

```go
func suggestOutputPath(cfg NewScriptConfig) string {
    // Get first scan directory
    config, _ := config.NewManager().Load()
    
    var baseDir string
    if len(config.ScanDirs) > 0 {
        baseDir = config.ScanDirs[0]
    } else {
        homeDir, _ := os.UserHomeDir()
        baseDir = filepath.Join(homeDir, "scripts")
    }
    
    // Add category subdirectory
    if cfg.Category != "" && cfg.Category != "uncategorized" {
        baseDir = filepath.Join(baseDir, cfg.Category)
    }
    
    // Determine extension
    ext := ".sh"
    if cfg.Shell == "python" {
        ext = ".py"
    }
    
    return filepath.Join(baseDir, cfg.Name+ext)
}
```

## Validation

```go
func validateScriptName(s string) error {
    if s == "" {
        return fmt.Errorf("name required")
    }
    
    // Must be valid identifier
    if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(s) {
        return fmt.Errorf("name must start with letter, contain only letters/numbers/hyphens/underscores")
    }
    
    // Check for conflicts with existing scripts
    scripts, _ := loadAllScripts()
    for _, script := range scripts {
        if script.Name == s {
            return fmt.Errorf("script %q already exists", s)
        }
    }
    
    return nil
}

func validateParamName(s string) error {
    if s == "" {
        return fmt.Errorf("parameter name required")
    }
    
    if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(s) {
        return fmt.Errorf("must be valid identifier")
    }
    
    return nil
}
```

## Testing

### Unit Tests

```go
func TestParseParamFlag(t *testing.T)
func TestParseParamFlag_Required(t *testing.T)
func TestSuggestOutputPath(t *testing.T)
func TestValidateScriptName(t *testing.T)
func TestValidateParamName(t *testing.T)
func TestTemplateGeneration_Bash(t *testing.T)
func TestTemplateGeneration_Python(t *testing.T)
func TestTemplateGeneration_WithParams(t *testing.T)
```

### Integration Tests

```go
func TestNew_Interactive(t *testing.T)
func TestNew_Headless(t *testing.T)
func TestNew_ExistingFile(t *testing.T)
func TestNew_CreatesDirectory(t *testing.T)
```

### Generated Script Tests

Verify generated scripts:
1. Have valid metadata that parses correctly
2. Are executable
3. Run without errors
