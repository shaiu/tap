# TUI Params Spec

> Parameter input forms

## Overview

When a script has parameters, tap displays an interactive form to collect values before execution. This spec covers the form UI, validation, and user experience.

## User Flow

```
Script selected (has params)
           │
           ▼
┌─────────────────────────────────────────┐
│  deploy                                 │
│  Deploy application to environment      │
│                                         │
│  Environment *                          │
│  ┌─────────────────────────────────┐    │
│  │ > staging                       │    │
│  │   production                    │    │
│  └─────────────────────────────────┘    │
│                                         │
│  Version                                │
│  ┌─────────────────────────────────┐    │
│  │ latest                          │    │
│  └─────────────────────────────────┘    │
│                                         │
│  Dry run                                │
│  ○ Yes  ● No                            │
│                                         │
│  tab next  shift+tab prev  enter run    │
│  esc cancel                             │
└─────────────────────────────────────────┘
           │ enter
           ▼
      Execute script
```

## Form Elements

Based on parameter types:

| Param Type | Form Element | Notes |
|------------|--------------|-------|
| string | Text input | With placeholder from default |
| string + choices | Select dropdown | Options from choices array |
| int | Text input | With numeric validation |
| float | Text input | With decimal validation |
| bool | Toggle/Confirm | Yes/No selection |
| path | Text input | With path completion (Phase 4) |

## Data Structures

### Form Model

```go
type FormModel struct {
    // Script being configured
    script      core.Script
    
    // Form from huh library
    form        *huh.Form
    
    // Collected values
    values      map[string]any
    
    // State
    submitted   bool
    cancelled   bool
    
    // Dimensions
    width       int
    height      int
}
```

### Form Result

```go
type FormResult struct {
    Script     core.Script
    Parameters map[string]string  // All values as strings for CLI
    Cancelled  bool
}
```

## Implementation

### Form Building

Using the `huh` library for form components:

```go
func NewFormModel(script core.Script, width, height int) FormModel {
    values := make(map[string]any)
    
    // Build form groups
    var fields []huh.Field
    
    for _, param := range script.Parameters {
        field := buildField(param, values)
        fields = append(fields, field)
    }
    
    form := huh.NewForm(
        huh.NewGroup(fields...),
    ).WithWidth(width - 4)
    
    return FormModel{
        script: script,
        form:   form,
        values: values,
        width:  width,
        height: height,
    }
}

func buildField(param core.Parameter, values map[string]any) huh.Field {
    // Set initial value from default
    if param.Default != nil {
        values[param.Name] = param.Default
    }
    
    title := param.Name
    if param.Required {
        title += " *"
    }
    
    switch {
    case len(param.Choices) > 0:
        // Select field for choices
        options := make([]huh.Option[string], len(param.Choices))
        for i, choice := range param.Choices {
            str := fmt.Sprintf("%v", choice)
            options[i] = huh.NewOption(str, str)
        }
        
        var value string
        if param.Default != nil {
            value = fmt.Sprintf("%v", param.Default)
        }
        values[param.Name] = &value
        
        return huh.NewSelect[string]().
            Key(param.Name).
            Title(title).
            Description(param.Description).
            Options(options...).
            Value(values[param.Name].(*string))
            
    case param.Type == "bool":
        // Confirm field for booleans
        value := false
        if param.Default != nil {
            value = param.Default.(bool)
        }
        values[param.Name] = &value
        
        return huh.NewConfirm().
            Key(param.Name).
            Title(title).
            Description(param.Description).
            Value(values[param.Name].(*bool))
            
    case param.Type == "int":
        // Text input with int validation
        value := ""
        if param.Default != nil {
            value = fmt.Sprintf("%d", param.Default)
        }
        values[param.Name] = &value
        
        return huh.NewInput().
            Key(param.Name).
            Title(title).
            Description(param.Description).
            Placeholder(value).
            Value(values[param.Name].(*string)).
            Validate(validateInt)
            
    case param.Type == "float":
        // Text input with float validation
        value := ""
        if param.Default != nil {
            value = fmt.Sprintf("%g", param.Default)
        }
        values[param.Name] = &value
        
        return huh.NewInput().
            Key(param.Name).
            Title(title).
            Description(param.Description).
            Placeholder(value).
            Value(values[param.Name].(*string)).
            Validate(validateFloat)
            
    default:
        // Text input for string/path
        value := ""
        if param.Default != nil {
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
```

### Validation

```go
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

func validateChoices(choices []any) func(string) error {
    valid := make(map[string]bool)
    for _, c := range choices {
        valid[fmt.Sprintf("%v", c)] = true
    }
    return func(s string) error {
        if !valid[s] {
            return fmt.Errorf("invalid choice")
        }
        return nil
    }
}
```

### Update Logic

```go
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.form = m.form.WithWidth(msg.Width - 4)
        return m, nil
        
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
                Parameters: m.collectValues(),
            }
        }
    }
    
    return m, cmd
}

func (m FormModel) collectValues() map[string]string {
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
```

### View Rendering

```go
func (m FormModel) View() string {
    var s strings.Builder
    
    // Header with script info
    s.WriteString(styles.Header.Render(m.script.Name))
    s.WriteString("\n")
    s.WriteString(styles.Dimmed.Render(m.script.Description))
    s.WriteString("\n\n")
    
    // Form
    s.WriteString(m.form.View())
    
    // Footer
    s.WriteString("\n")
    s.WriteString(styles.Footer.Render("tab next  shift+tab prev  enter run  esc cancel"))
    
    return s.String()
}
```

## Messages

```go
// FormSubmittedMsg is sent when form is completed
type FormSubmittedMsg struct {
    Script     core.Script
    Parameters map[string]string
}

// FormCancelledMsg is sent when user cancels
type FormCancelledMsg struct{}
```

## Integration with Menu

The main app model handles transitions:

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ScriptSelectedMsg:
        if len(msg.Script.Parameters) > 0 {
            // Script has params - show form
            m.state = stateForm
            m.formModel = NewFormModel(msg.Script, m.width, m.height)
            return m, nil
        }
        // No params - execute directly
        return m, m.executeScript(msg.Script, nil)
        
    case FormSubmittedMsg:
        // Form completed - execute
        m.state = stateExecuting
        return m, m.executeScript(msg.Script, msg.Parameters)
        
    case FormCancelledMsg:
        // Return to script list
        m.state = stateScriptList
        return m, nil
    }
    
    // Route to active component...
}
```

## Skip Form Option

If all parameters have defaults or are optional:

```go
func (m FormModel) canSkip() bool {
    for _, param := range m.script.Parameters {
        if param.Required && param.Default == nil {
            return false
        }
    }
    return true
}
```

Show a "Run with defaults" option:

```
┌─────────────────────────────────────────┐
│  deploy                                 │
│                                         │
│  All parameters have defaults.          │
│                                         │
│  > Run with defaults                    │
│    Customize parameters                 │
│                                         │
└─────────────────────────────────────────┘
```

## Required Field Indicator

Mark required fields clearly:

```go
func formatTitle(param core.Parameter) string {
    title := param.Name
    if param.Required {
        title += " " + styles.Required.Render("*")
    }
    return title
}
```

With style:

```go
styles.Required = lipgloss.NewStyle().
    Foreground(lipgloss.Color("196")) // Red
```

## Keyboard Navigation

Using huh's built-in navigation:

| Key | Action |
|-----|--------|
| Tab | Next field |
| Shift+Tab | Previous field |
| Enter | Submit (on last field) or Next |
| ↑/↓ | Navigate select options |
| Space | Toggle bool |
| Esc | Cancel form |

## Error Display

Validation errors appear inline:

```
  Version
  ┌─────────────────────────────────────┐
  │ abc                                 │
  └─────────────────────────────────────┘
  ⚠ must be an integer
```

## Testing

### Unit Tests

```go
func TestBuildField_String(t *testing.T)
func TestBuildField_StringWithChoices(t *testing.T)
func TestBuildField_Int(t *testing.T)
func TestBuildField_Bool(t *testing.T)
func TestBuildField_Required(t *testing.T)
func TestValidateInt(t *testing.T)
func TestValidateFloat(t *testing.T)
func TestValidateRequired(t *testing.T)
func TestCollectValues(t *testing.T)
func TestFormSubmission(t *testing.T)
func TestFormCancellation(t *testing.T)
func TestCanSkip(t *testing.T)
```

### Integration Tests

```go
func TestFormFlow_CompleteForm(t *testing.T)
func TestFormFlow_CancelForm(t *testing.T)
func TestFormFlow_ValidationError(t *testing.T)
func TestFormFlow_DefaultValues(t *testing.T)
```
