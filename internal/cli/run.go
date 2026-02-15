package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
	"github.com/shaiungar/tap/internal/tui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <script> [flags] [param=value...]",
	Short: "Run a script",
	Long: `Run a script directly by name.

Scripts can be specified by name alone or with category prefix (category/name).
Parameters can be passed via --param flags or inline as key=value arguments.

Examples:
  tap run deploy
  tap run deployment/deploy
  tap run deploy --param env=production --param version=v2.1
  tap run deploy -p env=staging -p dry_run=true
  tap run deploy env=production version=v2.1`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringSliceP("param", "p", nil, "Parameter in key=value format")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	scriptName := args[0]

	// Parse inline params (key=value)
	inlineParams := parseInlineParams(args[1:])

	// Parse --param flags
	flagParams, _ := cmd.Flags().GetStringSlice("param")
	parsedFlagParams := parseParamFlags(flagParams)

	// Merge params (inline params take precedence over flag params)
	params := mergeParams(parsedFlagParams, inlineParams)

	// Load scripts and find the requested one
	categories, err := loadScripts()
	if err != nil {
		return err
	}

	script, err := findScript(categories, scriptName)
	if err != nil {
		return err
	}

	mode := determineMode(cmd)

	// Check if we need interactive param input (skip for interactive scripts)
	if mode == ModeInteractive && !script.Interactive && needsParamInput(*script, params) {
		return runParamForm(*script, params)
	}

	// Validate params (skip for interactive scripts that handle their own input)
	if !script.Interactive {
		if err := validateParams(*script, params); err != nil {
			return err
		}
	}

	// Apply defaults for missing optional params
	params = applyDefaults(*script, params)

	// Execute
	return executeAndExit(*script, params)
}

// parseInlineParams extracts key=value pairs from args.
func parseInlineParams(args []string) map[string]string {
	params := make(map[string]string)
	for _, arg := range args {
		if parts := strings.SplitN(arg, "=", 2); len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}
	return params
}

// parseParamFlags extracts key=value pairs from --param flags.
func parseParamFlags(flags []string) map[string]string {
	params := make(map[string]string)
	for _, flag := range flags {
		if parts := strings.SplitN(flag, "=", 2); len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}
	return params
}

// mergeParams merges two param maps, with later taking precedence.
func mergeParams(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// findScript finds a script by name or category/name.
func findScript(categories []core.Category, name string) (*core.Script, error) {
	// Check if name includes category prefix
	var categoryFilter, scriptName string
	if parts := strings.SplitN(name, "/", 2); len(parts) == 2 {
		categoryFilter = parts[0]
		scriptName = parts[1]
	} else {
		scriptName = name
	}

	// Search for the script
	var matches []core.Script
	for _, cat := range categories {
		if categoryFilter != "" && cat.Name != categoryFilter {
			continue
		}
		for _, s := range cat.Scripts {
			if s.Name == scriptName {
				matches = append(matches, s)
			}
		}
	}

	if len(matches) == 0 {
		return nil, scriptNotFoundError(categories, name)
	}

	if len(matches) > 1 {
		// Multiple scripts with same name in different categories
		var names []string
		for _, s := range matches {
			names = append(names, fmt.Sprintf("%s/%s", s.Category, s.Name))
		}
		return nil, fmt.Errorf("ambiguous script name %q. Use category prefix:\n  %s",
			name, strings.Join(names, "\n  "))
	}

	return &matches[0], nil
}

// scriptNotFoundError builds a helpful error message.
func scriptNotFoundError(categories []core.Category, name string) error {
	var available []string
	for _, cat := range categories {
		for _, s := range cat.Scripts {
			available = append(available, s.Name)
		}
	}

	if len(available) == 0 {
		return fmt.Errorf("script not found: %s\n\nNo scripts available. Run 'tap config add-dir <path>' to add a directory to scan.", name)
	}

	// Limit to first 10
	if len(available) > 10 {
		available = append(available[:10], "...")
	}

	return fmt.Errorf("script not found: %s\n\nAvailable scripts:\n  %s\n\nRun 'tap list' to see all scripts.",
		name, strings.Join(available, ", "))
}

// needsParamInput returns true if required params are missing.
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

// validateParams checks that all required params are provided and types are valid.
func validateParams(script core.Script, params map[string]string) error {
	for _, param := range script.Parameters {
		value, provided := params[param.Name]

		// Check required params
		if param.Required && param.Default == nil && !provided {
			return fmt.Errorf("missing required parameter: %s", param.Name)
		}

		// Validate type if value provided
		if provided {
			if err := validateParamType(param, value); err != nil {
				return fmt.Errorf("invalid parameter %q: %w", param.Name, err)
			}

			// Validate choices if specified
			if len(param.Choices) > 0 {
				if err := validateChoice(param, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateParamType checks if the value matches the expected type.
func validateParamType(param core.Parameter, value string) error {
	switch param.Type {
	case core.ParamTypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("expected integer")
		}
	case core.ParamTypeFloat:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("expected number")
		}
	case core.ParamTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("expected boolean (true/false)")
		}
	}
	return nil
}

// validateChoice checks if the value is in the allowed choices.
func validateChoice(param core.Parameter, value string) error {
	for _, choice := range param.Choices {
		if fmt.Sprintf("%v", choice) == value {
			return nil
		}
	}
	var choiceStrs []string
	for _, c := range param.Choices {
		choiceStrs = append(choiceStrs, fmt.Sprintf("%v", c))
	}
	return fmt.Errorf("invalid value for %q: must be one of [%s]",
		param.Name, strings.Join(choiceStrs, ", "))
}

// applyDefaults fills in default values for missing optional params.
func applyDefaults(script core.Script, params map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		result[k] = v
	}
	for _, param := range script.Parameters {
		if _, ok := result[param.Name]; !ok && param.Default != nil {
			result[param.Name] = fmt.Sprintf("%v", param.Default)
		}
	}
	return result
}

// runParamForm shows the interactive parameter form.
func runParamForm(script core.Script, prefilledParams map[string]string) error {
	// Create form model with prefilled values
	model := tui.NewFormModelWithValues(script, prefilledParams, 80, 24)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running form: %w", err)
	}

	// Check if form was submitted
	formModel, ok := finalModel.(tui.FormModel)
	if !ok || formModel.Cancelled() {
		return nil // User cancelled
	}

	if !formModel.Submitted() {
		return nil // User quit
	}

	// Get collected params and merge with prefilled
	collectedParams := formModel.CollectValues()
	params := mergeParams(prefilledParams, collectedParams)

	// Apply defaults
	params = applyDefaults(script, params)

	// Execute
	return executeAndExit(script, params)
}

// executeAndExit runs the script and exits with its exit code.
func executeAndExit(script core.Script, params map[string]string) error {
	executor := core.NewExecutor()

	result, err := executor.Execute(context.Background(), core.ExecutionRequest{
		Script:     script,
		Parameters: params,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	})
	if err != nil {
		return err
	}

	if result.Error != nil {
		return result.Error
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}
