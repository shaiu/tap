package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/charmbracelet/huh"
	"github.com/shaiungar/tap/internal/config"
	"github.com/shaiungar/tap/internal/core"
	"github.com/shaiungar/tap/internal/templates"
	"github.com/spf13/cobra"
)

// NewScriptConfig holds configuration for creating a new script.
type NewScriptConfig struct {
	Name        string
	Description string
	Category    string
	Parameters  []ParameterConfig
	Shell       string // bash, python
	OutputPath  string
}

// ParameterConfig holds configuration for a script parameter.
type ParameterConfig struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
}

// TemplateData is passed to script templates.
type TemplateData struct {
	Name         string
	Description  string
	Category     string
	Parameters   []ParameterConfig
	ParamEnvVars []string
}

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new script",
	Long: `Create a new script with proper metadata boilerplate.

In interactive mode, a form guides you through script creation.
In headless mode, use flags to provide required values.

Examples:
  tap new                                    # Interactive mode
  tap new backup-db                          # Interactive with name pre-filled
  tap new backup-db -d "Backup database" -c maintenance
  tap new deploy --shell python -o ~/scripts/deploy.py
  tap new greet -d "Greet user" -p "name:string:Name to greet:required"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	newCmd.Flags().StringP("description", "d", "", "Script description")
	newCmd.Flags().StringP("category", "c", "", "Script category")
	newCmd.Flags().StringP("output", "o", "", "Output path")
	newCmd.Flags().StringSliceP("param", "p", nil, "Parameter (name:type:description:default)")
	newCmd.Flags().String("shell", "bash", "Shell type (bash, python)")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	mode := determineMode(cmd)

	if mode == ModeInteractive {
		return runNewInteractive(args)
	}

	return runNewHeadless(cmd, args)
}

// runNewInteractive runs the interactive script creation wizard.
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

	// Build main form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Name").
				Description("Script identifier (used in 'tap run <name>')").
				Value(&cfg.Name).
				Validate(validateScriptName),

			huh.NewInput().
				Key("description").
				Title("Description").
				Description("One-line description").
				Value(&cfg.Description).
				Validate(validateRequired("description")),

			huh.NewSelect[string]().
				Key("category").
				Title("Category").
				Options(buildCategoryOptions(categories)...).
				Value(&cfg.Category),

			huh.NewSelect[string]().
				Key("shell").
				Title("Shell").
				Options(
					huh.NewOption("Bash", "bash"),
					huh.NewOption("Python", "python"),
				).
				Value(&cfg.Shell),
		),
	)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil // User cancelled
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Handle "new category" selection
	if cfg.Category == "__new__" {
		var newCategory string
		categoryForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Key("new_category").
					Title("New category name").
					Value(&newCategory).
					Validate(validateCategoryName),
			),
		)
		if err := categoryForm.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return fmt.Errorf("form error: %w", err)
		}
		cfg.Category = newCategory
	}

	// Ask about parameters
	var addParams bool
	paramConfirm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("add_params").
				Title("Add parameters?").
				Value(&addParams),
		),
	)
	if err := paramConfirm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	if addParams {
		params, err := runParameterWizard()
		if err != nil {
			return err
		}
		cfg.Parameters = params
	}

	// Suggest and confirm output path
	suggestedPath := suggestOutputPath(cfg)
	cfg.OutputPath = suggestedPath

	var outputPath string
	pathForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("output_path").
				Title("Save to").
				Value(&outputPath).
				Placeholder(suggestedPath).
				Validate(func(s string) error {
					// Allow empty to use suggested path
					return nil
				}),
		),
	)
	if err := pathForm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	if outputPath != "" {
		cfg.OutputPath = expandPath(outputPath)
	}

	// Generate and save
	return generateScript(cfg, false)
}

// runParameterWizard runs an interactive wizard for adding parameters.
func runParameterWizard() ([]ParameterConfig, error) {
	var params []ParameterConfig

	for {
		var action string
		opts := []huh.Option[string]{
			huh.NewOption("Add parameter", "add"),
			huh.NewOption("Done", "done"),
		}
		if len(params) > 0 {
			opts = append([]huh.Option[string]{
				huh.NewOption("Add parameter", "add"),
				huh.NewOption(fmt.Sprintf("Done (%d parameters)", len(params)), "done"),
			}, huh.NewOption("Remove last", "remove"))
		}

		actionForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("action").
					Title("Parameters").
					Options(opts...).
					Value(&action),
			),
		)

		if err := actionForm.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return params, nil
			}
			return nil, fmt.Errorf("form error: %w", err)
		}

		switch action {
		case "add":
			param, err := promptParameter()
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		case "done":
			return params, nil
		case "remove":
			if len(params) > 0 {
				params = params[:len(params)-1]
			}
		}
	}
}

// promptParameter prompts for a single parameter configuration.
func promptParameter() (ParameterConfig, error) {
	var param ParameterConfig
	param.Type = "string" // default

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
					huh.NewOption("float", "float"),
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

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return param, err
		}
		return param, fmt.Errorf("form error: %w", err)
	}

	return param, nil
}

// runNewHeadless runs script creation in headless mode.
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

	// Validate description
	if cfg.Description == "" {
		return fmt.Errorf("--description required")
	}

	// Validate name
	if err := validateScriptName(cfg.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	// Validate shell
	if cfg.Shell != "bash" && cfg.Shell != "python" {
		return fmt.Errorf("invalid shell %q: must be bash or python", cfg.Shell)
	}

	// Parse parameter flags
	paramStrs, _ := cmd.Flags().GetStringSlice("param")
	for _, p := range paramStrs {
		param, err := parseParamFlag(p)
		if err != nil {
			return fmt.Errorf("invalid param %q: %w", p, err)
		}
		cfg.Parameters = append(cfg.Parameters, param)
	}

	// Default output path
	if cfg.OutputPath == "" {
		cfg.OutputPath = suggestOutputPath(cfg)
	} else {
		cfg.OutputPath = expandPath(cfg.OutputPath)
	}

	return generateScript(cfg, true)
}

// parseParamFlag parses a parameter flag in format name:type:description:default.
// "required" as default value means required=true.
func parseParamFlag(s string) (ParameterConfig, error) {
	parts := strings.SplitN(s, ":", 4)
	if len(parts) < 2 {
		return ParameterConfig{}, fmt.Errorf("expected name:type[:description[:default]]")
	}

	param := ParameterConfig{
		Name: parts[0],
		Type: parts[1],
	}

	// Validate name
	if err := validateParamName(param.Name); err != nil {
		return ParameterConfig{}, err
	}

	// Validate type
	if !core.IsValidParamType(param.Type) {
		return ParameterConfig{}, fmt.Errorf("invalid type %q: must be one of %v", param.Type, core.ValidParamTypes())
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

// loadExistingCategories returns unique categories from existing scripts.
func loadExistingCategories() []string {
	categories, err := loadScripts()
	if err != nil {
		return nil
	}

	var names []string
	for _, cat := range categories {
		if cat.Name != "uncategorized" {
			names = append(names, cat.Name)
		}
	}
	return names
}

// buildCategoryOptions builds huh options for category selection.
func buildCategoryOptions(existing []string) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(existing)+2)

	opts = append(opts, huh.NewOption("uncategorized", "uncategorized"))

	for _, cat := range existing {
		opts = append(opts, huh.NewOption(cat, cat))
	}

	opts = append(opts, huh.NewOption("+ Create new category", "__new__"))

	return opts
}

// suggestOutputPath suggests an output path based on config.
func suggestOutputPath(cfg NewScriptConfig) string {
	mgr, err := config.NewManager()
	if err != nil {
		return fallbackOutputPath(cfg)
	}

	conf, err := mgr.Load()
	if err != nil {
		return fallbackOutputPath(cfg)
	}

	var baseDir string
	if len(conf.ScanDirs) > 0 {
		baseDir = expandPath(conf.ScanDirs[0])
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

// fallbackOutputPath returns a fallback output path when config is unavailable.
func fallbackOutputPath(cfg NewScriptConfig) string {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, "scripts")

	if cfg.Category != "" && cfg.Category != "uncategorized" {
		baseDir = filepath.Join(baseDir, cfg.Category)
	}

	ext := ".sh"
	if cfg.Shell == "python" {
		ext = ".py"
	}

	return filepath.Join(baseDir, cfg.Name+ext)
}

// generateScript generates and writes the script file.
func generateScript(cfg NewScriptConfig, headless bool) error {
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
	tmplContent, err := templates.FS.ReadFile(tmplName)
	if err != nil {
		return fmt.Errorf("template not found for shell %s", cfg.Shell)
	}

	tmpl, err := template.New(cfg.Shell).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}

	// Check for existing file
	if _, err := os.Stat(cfg.OutputPath); err == nil {
		if headless {
			return fmt.Errorf("file already exists: %s (use different path or remove existing file)", cfg.OutputPath)
		}

		var overwrite bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Key("overwrite").
					Title(fmt.Sprintf("File exists: %s\nOverwrite?", cfg.OutputPath)).
					Value(&overwrite),
			),
		)
		if err := confirmForm.Run(); err != nil || !overwrite {
			return fmt.Errorf("cancelled")
		}
	}

	// Write file
	if err := os.WriteFile(cfg.OutputPath, buf.Bytes(), 0755); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Created: %s\n", cfg.OutputPath)
	fmt.Printf("\nEdit your script, then run it with:\n")
	fmt.Printf("  tap run %s\n", cfg.Name)

	return nil
}

// Validation functions

var scriptNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
var paramNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
var categoryNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

func validateScriptName(s string) error {
	if s == "" {
		return fmt.Errorf("name required")
	}

	if !scriptNameRegex.MatchString(s) {
		return fmt.Errorf("must start with letter, contain only letters/numbers/hyphens/underscores")
	}

	return nil
}

func validateRequired(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s required", field)
		}
		return nil
	}
}

func validateParamName(s string) error {
	if s == "" {
		return fmt.Errorf("parameter name required")
	}

	if !paramNameRegex.MatchString(s) {
		return fmt.Errorf("must be valid identifier (start with letter, alphanumeric and underscores)")
	}

	return nil
}

func validateCategoryName(s string) error {
	if s == "" {
		return fmt.Errorf("category name required")
	}

	if !categoryNameRegex.MatchString(s) {
		return fmt.Errorf("must start with letter, contain only letters/numbers/hyphens/underscores")
	}

	return nil
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}
