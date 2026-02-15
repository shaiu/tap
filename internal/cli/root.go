// Package cli provides the command-line interface for tap.
package cli

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/shaiungar/tap/internal/config"
	"github.com/shaiungar/tap/internal/core"
	"github.com/shaiungar/tap/internal/tui"
	"github.com/spf13/cobra"
)

// Build information (set via ldflags).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// ExecutionMode determines how tap runs.
type ExecutionMode int

const (
	// ModeInteractive launches the TUI.
	ModeInteractive ExecutionMode = iota
	// ModeHeadless runs without TUI (for scripting/CI).
	ModeHeadless
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "tap",
	Short: "Script runner with interactive TUI",
	Long: `tap is a terminal-based script runner that gives developers
quick access to their shell and Python scripts through an
interactive TUI menu or direct CLI invocation.`,
	SilenceUsage:  true,
	SilenceErrors: true,
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
	rootCmd.PersistentFlags().String("config", "", "Config file path")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// determineMode returns the execution mode based on flags, env vars, and TTY state.
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

// runTUI launches the interactive TUI.
func runTUI() error {
	// Load configuration
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Get registered scripts
	registry, err := mgr.GetRegistry()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	var registeredPaths []string
	for _, s := range registry.Scripts {
		registeredPaths = append(registeredPaths, s.Path)
	}

	// Scan for scripts
	scanner := core.NewScanner(core.ScannerConfig{
		Directories:       cfg.ScanDirs,
		Extensions:        cfg.Extensions,
		IgnoreDirs:        cfg.IgnoreDirs,
		MaxDepth:          cfg.MaxDepth,
		RegisteredScripts: registeredPaths,
		AutoGenMetadata:   cfg.GetAutoGenMetadata(),
	})

	scripts, err := scanner.Scan(context.Background())
	if err != nil {
		return fmt.Errorf("scanning scripts: %w", err)
	}

	// Organize into categories
	categories := core.OrganizeByCategory(scripts)

	// Create and run the TUI
	model := tui.NewAppModel(categories)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	// Check if a script was selected
	appModel, ok := finalModel.(tui.AppModel)
	if !ok {
		return nil
	}

	selectedScript := appModel.SelectedScript()
	if selectedScript == nil {
		return nil // User quit without selecting
	}

	// Execute the selected script
	executor := core.NewExecutor()

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script:     *selectedScript,
		Parameters: appModel.SelectedParams(),
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	})
	if err != nil {
		return err
	}

	// Exit with the script's exit code
	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}

// RootCmd returns the root command for testing purposes.
func RootCmd() *cobra.Command {
	return rootCmd
}
