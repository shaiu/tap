package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/shaiungar/tap/internal/config"
	"github.com/shaiungar/tap/internal/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available scripts",
	Long: `List all discovered scripts from configured scan directories.

By default, scripts are grouped by category. Use --flat to list without grouping,
or --json to output in JSON format for scripting.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringP("category", "c", "", "Filter by category")
	listCmd.Flags().Bool("flat", false, "List without category grouping")
	listCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	categoryFilter, _ := cmd.Flags().GetString("category")
	flat, _ := cmd.Flags().GetBool("flat")
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Load scripts using shared helper
	categories, err := loadScripts()
	if err != nil {
		return err
	}

	// Filter by category if specified
	if categoryFilter != "" {
		categories = filterByCategory(categories, categoryFilter)
	}

	// Output based on format
	if jsonOut {
		return outputJSON(categories)
	}

	if flat {
		return outputFlat(categories)
	}

	return outputGrouped(categories)
}

// loadScripts loads all scripts from configured sources and organizes them by category.
func loadScripts() ([]core.Category, error) {
	mgr, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	cfg, err := mgr.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	registry, err := mgr.GetRegistry()
	if err != nil {
		return nil, fmt.Errorf("loading registry: %w", err)
	}

	var registeredPaths []string
	for _, s := range registry.Scripts {
		registeredPaths = append(registeredPaths, s.Path)
	}

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
		return nil, fmt.Errorf("scanning scripts: %w", err)
	}

	return core.OrganizeByCategory(scripts), nil
}

// filterByCategory filters categories to only include the specified category.
func filterByCategory(categories []core.Category, name string) []core.Category {
	for _, cat := range categories {
		if cat.Name == name {
			return []core.Category{cat}
		}
	}
	return nil
}

// outputGrouped prints scripts grouped by category.
func outputGrouped(categories []core.Category) error {
	if len(categories) == 0 {
		fmt.Println("No scripts found.")
		fmt.Println("\nRun 'tap config add-dir <path>' to add a directory to scan.")
		return nil
	}

	for i, cat := range categories {
		fmt.Printf("%s:\n", cat.Name)
		for _, script := range cat.Scripts {
			if script.Description != "" {
				fmt.Printf("  %-20s %s\n", script.Name, script.Description)
			} else {
				fmt.Printf("  %s\n", script.Name)
			}
		}
		if i < len(categories)-1 {
			fmt.Println()
		}
	}
	return nil
}

// outputFlat prints scripts without category grouping.
func outputFlat(categories []core.Category) error {
	if len(categories) == 0 {
		fmt.Println("No scripts found.")
		return nil
	}

	for _, cat := range categories {
		for _, script := range cat.Scripts {
			if script.Description != "" {
				fmt.Printf("%-20s %s\n", script.Name, script.Description)
			} else {
				fmt.Println(script.Name)
			}
		}
	}
	return nil
}

// outputJSON prints scripts as JSON.
func outputJSON(categories []core.Category) error {
	// Flatten for JSON output
	var scripts []core.Script
	for _, cat := range categories {
		scripts = append(scripts, cat.Scripts...)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(scripts)
}
