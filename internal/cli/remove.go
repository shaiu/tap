package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaiungar/tap/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <script>",
	Short: "Unregister a script",
	Long: `Unregister a script from tap's registry.

Scripts can be removed by:
  - Path (absolute or relative)
  - Alias (if one was set during registration)

Examples:
  tap remove ~/scripts/my-tool.sh
  tap remove my-tool                  # Remove by alias`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	nameOrPath := args[0]

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Try to resolve as path first
	pathToRemove := nameOrPath
	if len(nameOrPath) > 1 && nameOrPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			pathToRemove = filepath.Join(home, nameOrPath[2:])
		}
	}
	absPath, _ := filepath.Abs(pathToRemove)

	// Try to unregister - this will match by path or alias
	err = mgr.UnregisterScript(absPath)
	if err != nil {
		// If not found by path, try the original name (might be an alias)
		if absPath != nameOrPath {
			err = mgr.UnregisterScript(nameOrPath)
		}
	}

	if err != nil {
		return err
	}

	fmt.Printf("Removed: %s\n", nameOrPath)
	return nil
}
