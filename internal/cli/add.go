package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaiungar/tap/internal/config"
	"github.com/shaiungar/tap/internal/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register an external script",
	Long: `Register an external script or directory of scripts.

For a single script, tap validates it has proper metadata.
For a directory with --recursive, all valid scripts are registered.

Examples:
  tap add ~/scripts/my-tool.sh
  tap add ~/scripts/my-tool.sh --alias tool
  tap add ~/company-scripts/ --recursive`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringP("alias", "a", "", "Alias for the script")
	addCmd.Flags().BoolP("recursive", "r", false, "Add all scripts in directory")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	path := args[0]
	alias, _ := cmd.Flags().GetString("alias")
	recursive, _ := cmd.Flags().GetBool("recursive")

	// Expand ~ in path
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path not found: %s", path)
		}
		return err
	}

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if info.IsDir() {
		if !recursive {
			return fmt.Errorf("%s is a directory. Use --recursive to add all scripts", path)
		}
		if alias != "" {
			return fmt.Errorf("--alias cannot be used with --recursive")
		}
		return addDirectory(mgr, absPath)
	}

	return addSingleScript(mgr, absPath, alias)
}

// addSingleScript registers a single script file.
func addSingleScript(mgr *config.DefaultManager, absPath, alias string) error {
	// Verify it has valid metadata
	script, err := core.ParseScript(absPath)
	if err != nil {
		return fmt.Errorf("invalid script: %w", err)
	}
	if script == nil {
		return fmt.Errorf("no valid metadata found in %s (requires name and description)", absPath)
	}

	if err := mgr.RegisterScript(absPath, alias); err != nil {
		return err
	}

	fmt.Printf("Registered: %s", script.Name)
	if alias != "" {
		fmt.Printf(" (alias: %s)", alias)
	}
	fmt.Println()

	return nil
}

// addDirectory recursively adds all valid scripts in a directory.
func addDirectory(mgr *config.DefaultManager, dir string) error {
	// Load config to get extensions and ignore dirs
	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	scanner := core.NewScanner(core.ScannerConfig{
		Directories:     []string{dir},
		Extensions:      cfg.Extensions,
		IgnoreDirs:      cfg.IgnoreDirs,
		MaxDepth:        cfg.MaxDepth,
		AutoGenMetadata: false, // Only add scripts with explicit metadata
	})

	scripts, err := scanner.ScanDirectory(context.Background(), dir)
	if err != nil {
		return fmt.Errorf("scanning directory: %w", err)
	}

	if len(scripts) == 0 {
		return fmt.Errorf("no scripts with valid metadata found in %s", dir)
	}

	// Get existing registry to check for already registered scripts
	registry, err := mgr.GetRegistry()
	if err != nil {
		registry = &config.Registry{}
	}

	// Build map of already registered paths for quick lookup
	registeredPaths := make(map[string]bool)
	for _, s := range registry.Scripts {
		registeredPaths[s.Path] = true
	}

	var added, skipped int
	for _, script := range scripts {
		// Skip if already registered
		if registeredPaths[script.Path] {
			skipped++
			continue
		}

		if err := mgr.RegisterScript(script.Path, ""); err != nil {
			// Skip duplicates and continue
			skipped++
			continue
		}

		fmt.Printf("Registered: %s (%s)\n", script.Name, script.Path)
		added++
	}

	if added == 0 && skipped > 0 {
		fmt.Printf("No new scripts added (%d already registered)\n", skipped)
	} else if skipped > 0 {
		fmt.Printf("\nAdded %d script(s), %d skipped (already registered)\n", added, skipped)
	} else {
		fmt.Printf("\nAdded %d script(s)\n", added)
	}

	return nil
}
