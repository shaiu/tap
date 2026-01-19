# CLI Spec

> Cobra commands and headless mode

## Overview

The CLI layer provides the entry points to tap, handling both interactive (TUI) and headless modes. Built with Cobra for consistent flag parsing and help generation.

## Command Structure

```
tap                       # Interactive TUI (default)
tap run <script>          # Run script (headless or interactive params)
tap list                  # List all scripts
tap new                   # Create new script
tap add <path>            # Register external script
tap remove <script>       # Unregister script
tap config                # Configuration management
tap version               # Show version
tap help                  # Show help
```

## Mode Detection

Tap automatically switches between interactive and headless modes:

```go
func determineMode(cmd *cobra.Command) ExecutionMode {
    // 1. Explicit flags take priority
    if headless, _ := cmd.Flags().GetBool("headless"); headless {
        return ModeHeadless
    }
    if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
        return ModeInteractive
    }
    
    // 2. Environment variables
    if os.Getenv("CI") != "" || os.Getenv("TAP_HEADLESS") != "" {
        return ModeHeadless
    }
    
    // 3. TTY detection
    if isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd()) {
        return ModeInteractive
    }
    
    return ModeHeadless
}
```

## Commands

### tap (root)

Launch interactive TUI.

```bash
tap                  # Full TUI experience
tap --headless       # Error: no script specified
```

```go
var rootCmd = &cobra.Command{
    Use:   "tap",
    Short: "Script runner with interactive TUI",
    Long: `tap is a terminal-based script runner that gives developers
quick access to their shell and Python scripts through an
interactive TUI menu or direct CLI invocation.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        mode := determineMode(cmd)
        
        if mode == ModeHeadless {
            return fmt.Errorf("no script specified. Use: tap run <script>")
        }
        
        return runTUI()
    },
}

func init() {
    rootCmd.PersistentFlags().Bool("headless", false, "Run in headless mode")
    rootCmd.PersistentFlags().Bool("interactive", false, "Force interactive mode")
}
```

### tap run

Run a script directly.

```bash
# Basic usage
tap run deploy
tap run deployment/deploy     # With category prefix

# With parameters
tap run deploy --param env=production --param version=v2.1
tap run deploy -p env=staging -p dry_run=true

# Short parameter syntax
tap run deploy env=production version=v2.1

# Force modes
tap run deploy --headless
tap run deploy --interactive   # Show param form even with all defaults
```

```go
var runCmd = &cobra.Command{
    Use:   "run <script> [flags] [param=value...]",
    Short: "Run a script",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        scriptName := args[0]
        
        // Parse inline params (key=value)
        inlineParams := parseInlineParams(args[1:])
        
        // Merge with --param flags
        flagParams, _ := cmd.Flags().GetStringSlice("param")
        params := mergeParams(inlineParams, parseParamFlags(flagParams))
        
        // Find script
        script, err := findScript(scriptName)
        if err != nil {
            return err
        }
        
        mode := determineMode(cmd)
        
        // Check if we need interactive param input
        if mode == ModeInteractive && needsParamInput(script, params) {
            return runParamForm(script, params)
        }
        
        // Validate params
        if err := validateParams(script, params); err != nil {
            return err
        }
        
        // Execute
        return executeAndExit(script, params)
    },
}

func init() {
    runCmd.Flags().StringSliceP("param", "p", nil, "Parameter in key=value format")
    rootCmd.AddCommand(runCmd)
}

func parseInlineParams(args []string) map[string]string {
    params := make(map[string]string)
    for _, arg := range args {
        if parts := strings.SplitN(arg, "=", 2); len(parts) == 2 {
            params[parts[0]] = parts[1]
        }
    }
    return params
}

func parseParamFlags(flags []string) map[string]string {
    params := make(map[string]string)
    for _, flag := range flags {
        if parts := strings.SplitN(flag, "=", 2); len(parts) == 2 {
            params[parts[0]] = parts[1]
        }
    }
    return params
}

func needsParamInput(script core.Script, provided map[string]string) bool {
    for _, param := range script.Parameters {
        if param.Required && param.Default == nil {
            if _, ok := provided[param.Name]; !ok {
                return true
            }
        }
    }
    return false
}
```

### tap list

List discovered scripts.

```bash
tap list                      # All scripts, grouped by category
tap list --category deploy    # Filter by category
tap list --flat               # No category grouping
tap list --json               # JSON output for scripting
```

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all available scripts",
    RunE: func(cmd *cobra.Command, args []string) error {
        category, _ := cmd.Flags().GetString("category")
        flat, _ := cmd.Flags().GetBool("flat")
        jsonOut, _ := cmd.Flags().GetBool("json")
        
        scripts, err := loadScripts()
        if err != nil {
            return err
        }
        
        // Filter by category
        if category != "" {
            scripts = filterByCategory(scripts, category)
        }
        
        // Output format
        if jsonOut {
            return outputJSON(scripts)
        }
        
        if flat {
            return outputFlat(scripts)
        }
        
        return outputGrouped(scripts)
    },
}

func init() {
    listCmd.Flags().StringP("category", "c", "", "Filter by category")
    listCmd.Flags().Bool("flat", false, "List without category grouping")
    listCmd.Flags().Bool("json", false, "Output as JSON")
    rootCmd.AddCommand(listCmd)
}

func outputGrouped(categories []core.Category) error {
    for _, cat := range categories {
        fmt.Printf("%s:\n", cat.Name)
        for _, script := range cat.Scripts {
            fmt.Printf("  %-20s %s\n", script.Name, script.Description)
        }
        fmt.Println()
    }
    return nil
}

func outputFlat(categories []core.Category) error {
    for _, cat := range categories {
        for _, script := range cat.Scripts {
            fmt.Printf("%-20s %s\n", script.Name, script.Description)
        }
    }
    return nil
}

func outputJSON(categories []core.Category) error {
    // Flatten for JSON
    var scripts []core.Script
    for _, cat := range categories {
        scripts = append(scripts, cat.Scripts...)
    }
    
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(scripts)
}
```

### tap new

Create a new script with scaffolding. See [07-scaffolding.md](07-scaffolding.md) for details.

```bash
tap new                       # Interactive prompts
tap new deploy                # With name
tap new deploy --category deployment --description "Deploy app"
```

### tap add

Register an external script.

```bash
tap add ~/scripts/my-tool.sh
tap add ~/scripts/my-tool.sh --alias tool
tap add ~/company-scripts/ --recursive    # Add all scripts in directory
```

```go
var addCmd = &cobra.Command{
    Use:   "add <path>",
    Short: "Register an external script",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        path := args[0]
        alias, _ := cmd.Flags().GetString("alias")
        recursive, _ := cmd.Flags().GetBool("recursive")
        
        info, err := os.Stat(path)
        if err != nil {
            return fmt.Errorf("path not found: %s", path)
        }
        
        cfg := config.NewManager()
        
        if info.IsDir() {
            if recursive {
                return addDirectory(cfg, path)
            }
            return fmt.Errorf("%s is a directory. Use --recursive to add all scripts", path)
        }
        
        // Verify it has valid metadata
        script, err := scanner.ParseScript(path)
        if err != nil || script == nil {
            return fmt.Errorf("no valid metadata found in %s", path)
        }
        
        if err := cfg.RegisterScript(path, alias); err != nil {
            return err
        }
        
        fmt.Printf("✓ Registered: %s", script.Name)
        if alias != "" {
            fmt.Printf(" (alias: %s)", alias)
        }
        fmt.Println()
        
        return nil
    },
}

func init() {
    addCmd.Flags().StringP("alias", "a", "", "Alias for the script")
    addCmd.Flags().BoolP("recursive", "r", false, "Add all scripts in directory")
    rootCmd.AddCommand(addCmd)
}
```

### tap remove

Unregister a script.

```bash
tap remove my-tool
```

```go
var removeCmd = &cobra.Command{
    Use:   "remove <script>",
    Short: "Unregister a script",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        
        cfg := config.NewManager()
        if err := cfg.UnregisterScript(name); err != nil {
            return err
        }
        
        fmt.Printf("✓ Removed: %s\n", name)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(removeCmd)
}
```

### tap config

Configuration management.

```bash
tap config show               # Display current config
tap config edit               # Open in $EDITOR
tap config add-dir <path>     # Add scan directory
tap config remove-dir <path>  # Remove scan directory
```

```go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
    Use:   "show",
    Short: "Show current configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.NewManager().Load()
        if err != nil {
            return err
        }
        
        data, _ := yaml.Marshal(cfg)
        fmt.Println(string(data))
        return nil
    },
}

var configAddDirCmd = &cobra.Command{
    Use:   "add-dir <path>",
    Short: "Add a directory to scan",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        mgr := config.NewManager()
        if err := mgr.AddScanDir(args[0]); err != nil {
            return err
        }
        fmt.Printf("✓ Added: %s\n", args[0])
        return nil
    },
}

func init() {
    configCmd.AddCommand(configShowCmd)
    configCmd.AddCommand(configAddDirCmd)
    // ... more subcommands
    rootCmd.AddCommand(configCmd)
}
```

### tap version

```go
var (
    Version   = "dev"
    Commit    = "none"
    BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("tap %s\n", Version)
        fmt.Printf("  commit: %s\n", Commit)
        fmt.Printf("  built:  %s\n", BuildDate)
    },
}
```

## Output Formatting

### Normal Output

Human-readable for terminals:

```
$ tap list
deployment:
  deploy               Deploy application to environment
  rollback             Rollback to previous version

data:
  export               Export data to file
  import               Import data from backup
```

### JSON Output

Machine-readable for scripting:

```json
[
  {
    "name": "deploy",
    "description": "Deploy application to environment",
    "category": "deployment",
    "path": "/home/user/scripts/deploy.sh",
    "parameters": [...]
  }
]
```

### Error Output

```
$ tap run nonexistent
Error: script not found: nonexistent

Available scripts:
  deploy, rollback, export, import

Run 'tap list' to see all scripts.
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage |
| 126 | Permission denied |
| 127 | Script not found |
| 128+N | Script killed by signal N |

## Global Flags

```go
func init() {
    rootCmd.PersistentFlags().Bool("headless", false, "Disable interactive mode")
    rootCmd.PersistentFlags().Bool("interactive", false, "Force interactive mode")
    rootCmd.PersistentFlags().String("config", "", "Config file path")
    rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
```

## Shell Completions (Phase 4)

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish]",
    Short: "Generate shell completion script",
    Args:  cobra.ExactValidArgs(1),
    ValidArgs: []string{"bash", "zsh", "fish"},
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            return rootCmd.GenZshCompletion(os.Stdout)
        case "fish":
            return rootCmd.GenFishCompletion(os.Stdout, true)
        }
        return nil
    },
}

// Dynamic script name completion
func scriptNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    scripts, _ := loadScripts()
    var names []string
    for _, cat := range scripts {
        for _, s := range cat.Scripts {
            if strings.HasPrefix(s.Name, toComplete) {
                names = append(names, s.Name+"\t"+s.Description)
            }
        }
    }
    return names, cobra.ShellCompDirectiveNoFileComp
}
```

## Testing

### Unit Tests

```go
func TestDetermineMode_Headless(t *testing.T)
func TestDetermineMode_Interactive(t *testing.T)
func TestDetermineMode_TTY(t *testing.T)
func TestDetermineMode_CI(t *testing.T)
func TestParseInlineParams(t *testing.T)
func TestParseParamFlags(t *testing.T)
func TestMergeParams(t *testing.T)
func TestNeedsParamInput(t *testing.T)
```

### Integration Tests

```go
func TestCLI_Run_Basic(t *testing.T)
func TestCLI_Run_WithParams(t *testing.T)
func TestCLI_List(t *testing.T)
func TestCLI_List_JSON(t *testing.T)
func TestCLI_Add(t *testing.T)
func TestCLI_Config(t *testing.T)
```

### Command-Line Tests

Using shell scripts to test full CLI behavior:

```bash
#!/bin/bash
# test/cli_test.sh

# Test list
output=$(tap list --json)
if ! echo "$output" | jq -e '.[0].name' > /dev/null; then
    echo "FAIL: list --json"
    exit 1
fi

# Test run
tap run echo-test --param message=hello
if [ $? -ne 0 ]; then
    echo "FAIL: run"
    exit 1
fi

echo "All CLI tests passed"
```
