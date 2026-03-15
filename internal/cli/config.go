package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shaiungar/tap/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage tap configuration",
	Long: `View and modify tap configuration.

Examples:
  tap config show
  tap config add-dir ~/scripts
  tap config remove-dir ~/scripts
  tap config set tui.theme minimal
  tap config reset --force
  tap config edit`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE:  runConfigShow,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in editor",
	RunE:  runConfigEdit,
}

var configAddDirCmd = &cobra.Command{
	Use:   "add-dir <path>",
	Short: "Add a scan directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigAddDir,
}

var configRemoveDirCmd = &cobra.Command{
	Use:   "remove-dir <path>",
	Short: "Remove a scan directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigRemoveDir,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dotted key notation.

Supported keys:
  tui.theme         TUI theme (default, minimal, etc.)
  tui.show_paths    Show script paths in TUI (true/false)
  max_depth         Maximum scan depth (integer)
  default_shell     Default shell for new scripts
  editor            Editor command for editing scripts`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE:  runConfigReset,
}

func init() {
	configResetCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configAddDirCmd)
	configCmd.AddCommand(configRemoveDirCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	fmt.Printf("# Configuration\n%s\n", string(data))
	fmt.Printf("# Paths\n")
	fmt.Printf("Config:   %s\n", mgr.Path())
	fmt.Printf("Registry: %s\n", mgr.RegistryPath())
	fmt.Printf("Cache:    %s\n", mgr.CachePath())

	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Load config to check for editor setting
	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	editor := resolveEditor(cfg.Editor)
	if editor == "" {
		return fmt.Errorf("no editor configured. Set $EDITOR or run: tap config set editor <command>")
	}

	c := exec.Command(editor, mgr.Path())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// resolveEditor returns the editor command to use, checking config, env vars, and fallback.
func resolveEditor(configEditor string) string {
	if configEditor != "" {
		return configEditor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	return "vi"
}

func runConfigAddDir(cmd *cobra.Command, args []string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := mgr.AddScanDir(args[0]); err != nil {
		return err
	}

	fmt.Printf("Added scan directory: %s\n", args[0])
	return nil
}

func runConfigRemoveDir(cmd *cobra.Command, args []string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := mgr.RemoveScanDir(args[0]); err != nil {
		return err
	}

	fmt.Printf("Removed scan directory: %s\n", args[0])
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := applyConfigSet(cfg, key, value); err != nil {
		return err
	}

	if err := mgr.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

// applyConfigSet applies a key=value setting to the config.
func applyConfigSet(cfg *config.Config, key, value string) error {
	switch key {
	case "tui.theme":
		cfg.TUI.Theme = value
	case "tui.show_paths":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s (use true/false)", value)
		}
		cfg.TUI.ShowPaths = b
	case "max_depth":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		cfg.MaxDepth = n
	case "default_shell":
		cfg.DefaultShell = value
	case "editor":
		cfg.Editor = value
	default:
		return fmt.Errorf("unknown config key: %s\n\nSupported keys: tui.theme, tui.show_paths, max_depth, default_shell, editor", key)
	}
	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Print("Reset configuration to defaults? This cannot be undone. [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	mgr, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := mgr.Save(config.DefaultConfig()); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("Configuration reset to defaults.")
	return nil
}
