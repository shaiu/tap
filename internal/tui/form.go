package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
)

// FormSubmittedMsg is sent when form is completed.
type FormSubmittedMsg struct {
	Script     core.Script
	Parameters map[string]string
}

// FormCancelledMsg is sent when user cancels.
type FormCancelledMsg struct{}

// FormModel handles parameter input for scripts.
type FormModel struct {
	// Script being configured
	script core.Script

	// Form from huh library
	form *huh.Form

	// Collected values - stored as pointers for huh bindings
	values map[string]any

	// State
	submitted bool
	cancelled bool

	// Dimensions
	width  int
	height int
}

// NewFormModel creates a new FormModel for the given script.
func NewFormModel(script core.Script, width, height int) FormModel {
	return NewFormModelWithValues(script, nil, width, height)
}

// NewFormModelWithValues creates a new FormModel with prefilled values.
func NewFormModelWithValues(script core.Script, prefilled map[string]string, width, height int) FormModel {
	values := make(map[string]any)
	if prefilled == nil {
		prefilled = make(map[string]string)
	}

	// Build form fields
	var fields []huh.Field
	for _, param := range script.Parameters {
		field := buildField(param, values, prefilled)
		fields = append(fields, field)
	}

	// Handle empty fields case - huh panics with empty groups
	var form *huh.Form
	if len(fields) == 0 {
		// Create a minimal form with a note field
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("No parameters").Description("Press Enter to run"),
			),
		).WithWidth(width - 4)
	} else {
		form = huh.NewForm(
			huh.NewGroup(fields...),
		).WithWidth(width - 4)
	}

	return FormModel{
		script: script,
		form:   form,
		values: values,
		width:  width,
		height: height,
	}
}

// buildField creates the appropriate huh field for a parameter.
func buildField(param core.Parameter, values map[string]any, prefilled map[string]string) huh.Field {
	title := formatTitle(param)

	// Check for prefilled value
	prefilledValue, hasPrefilled := prefilled[param.Name]

	switch {
	case len(param.Choices) > 0:
		// Select field for choices
		options := make([]huh.Option[string], len(param.Choices))
		for i, choice := range param.Choices {
			str := fmt.Sprintf("%v", choice)
			options[i] = huh.NewOption(str, str)
		}

		var value string
		if hasPrefilled {
			value = prefilledValue
		} else if param.Default != nil {
			value = fmt.Sprintf("%v", param.Default)
		}
		values[param.Name] = &value

		return huh.NewSelect[string]().
			Key(param.Name).
			Title(title).
			Description(param.Description).
			Options(options...).
			Value(values[param.Name].(*string))

	case param.Type == core.ParamTypeBool:
		// Confirm field for booleans
		value := false
		if hasPrefilled {
			value, _ = strconv.ParseBool(prefilledValue)
		} else if param.Default != nil {
			if b, ok := param.Default.(bool); ok {
				value = b
			}
		}
		values[param.Name] = &value

		return huh.NewConfirm().
			Key(param.Name).
			Title(title).
			Description(param.Description).
			Value(values[param.Name].(*bool))

	case param.Type == core.ParamTypeInt:
		// Text input with int validation
		value := ""
		if hasPrefilled {
			value = prefilledValue
		} else if param.Default != nil {
			value = fmt.Sprintf("%v", param.Default)
		}
		values[param.Name] = &value

		input := huh.NewInput().
			Key(param.Name).
			Title(title).
			Description(param.Description).
			Placeholder(value).
			Value(values[param.Name].(*string)).
			Validate(validateInt)

		return input

	case param.Type == core.ParamTypeFloat:
		// Text input with float validation
		value := ""
		if hasPrefilled {
			value = prefilledValue
		} else if param.Default != nil {
			value = fmt.Sprintf("%g", param.Default)
		}
		values[param.Name] = &value

		input := huh.NewInput().
			Key(param.Name).
			Title(title).
			Description(param.Description).
			Placeholder(value).
			Value(values[param.Name].(*string)).
			Validate(validateFloat)

		return input

	default:
		// Text input for string/path
		value := ""
		if hasPrefilled {
			value = prefilledValue
		} else if param.Default != nil {
			value = fmt.Sprintf("%v", param.Default)
		}
		values[param.Name] = &value

		input := huh.NewInput().
			Key(param.Name).
			Title(title).
			Description(param.Description).
			Placeholder(value).
			Value(values[param.Name].(*string))

		if param.Required {
			input = input.Validate(validateRequired)
		}

		return input
	}
}

// formatTitle formats the parameter title with required indicator.
func formatTitle(param core.Parameter) string {
	title := param.Name
	if param.Required {
		title += " *"
	}
	return title
}

// Validation functions

func validateRequired(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	return nil
}

func validateInt(s string) error {
	if s == "" {
		return nil // Allow empty for optional
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be an integer")
	}
	return nil
}

func validateFloat(s string) error {
	if s == "" {
		return nil
	}
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	return nil
}

// Init implements tea.Model.
func (m FormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.form = m.form.WithWidth(msg.Width - 4)
		return m, nil

	case FormSubmittedMsg:
		// When running standalone (not embedded in AppModel), quit after submission
		return m, tea.Quit

	case FormCancelledMsg:
		// When running standalone, quit after cancellation
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.cancelled = true
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form completed
	if m.form.State == huh.StateCompleted {
		m.submitted = true
		return m, func() tea.Msg {
			return FormSubmittedMsg{
				Script:     m.script,
				Parameters: m.CollectValues(),
			}
		}
	}

	return m, cmd
}

// CollectValues gathers all form values as strings.
func (m FormModel) CollectValues() map[string]string {
	result := make(map[string]string)

	for _, param := range m.script.Parameters {
		val := m.values[param.Name]

		switch v := val.(type) {
		case *string:
			if *v != "" {
				result[param.Name] = *v
			} else if param.Default != nil {
				result[param.Name] = fmt.Sprintf("%v", param.Default)
			}
		case *bool:
			result[param.Name] = strconv.FormatBool(*v)
		default:
			result[param.Name] = fmt.Sprintf("%v", val)
		}
	}

	return result
}

// View implements tea.Model.
func (m FormModel) View() string {
	var s strings.Builder

	// Header with script info
	s.WriteString(Styles.Header.Render(m.script.Name))
	s.WriteString("\n")
	s.WriteString(Styles.Dimmed.Render(m.script.Description))
	s.WriteString("\n\n")

	// Form
	s.WriteString(m.form.View())

	// Footer
	s.WriteString("\n")
	s.WriteString(Styles.Footer.Render("tab next  shift+tab prev  enter run  esc cancel"))

	return s.String()
}

// Script returns the script being configured.
func (m FormModel) Script() core.Script {
	return m.script
}

// Submitted returns true if the form was submitted.
func (m FormModel) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if the form was cancelled.
func (m FormModel) Cancelled() bool {
	return m.cancelled
}

// CanSkip returns true if all parameters have defaults or are optional.
func (m FormModel) CanSkip() bool {
	for _, param := range m.script.Parameters {
		if param.Required && param.Default == nil {
			return false
		}
	}
	return true
}

// DefaultValues returns all default values as a map.
func (m FormModel) DefaultValues() map[string]string {
	result := make(map[string]string)
	for _, param := range m.script.Parameters {
		if param.Default != nil {
			result[param.Name] = fmt.Sprintf("%v", param.Default)
		} else if param.Type == core.ParamTypeBool {
			result[param.Name] = "false"
		}
	}
	return result
}
