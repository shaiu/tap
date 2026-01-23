package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shaiungar/tap/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"valid value", "test", false},
		{"value with spaces", "  test  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequired(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", false},
		{"valid integer", "42", false},
		{"negative integer", "-10", false},
		{"zero", "0", false},
		{"not an integer", "abc", true},
		{"float", "3.14", true},
		{"mixed", "12abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "integer")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", false},
		{"valid integer", "42", false},
		{"valid float", "3.14", false},
		{"negative float", "-2.5", false},
		{"scientific notation", "1e10", false},
		{"not a number", "abc", true},
		{"mixed", "12.3abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFloat(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "number")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatTitle(t *testing.T) {
	tests := []struct {
		name     string
		param    core.Parameter
		expected string
	}{
		{
			"optional param",
			core.Parameter{Name: "version", Required: false},
			"version",
		},
		{
			"required param",
			core.Parameter{Name: "name", Required: true},
			"name *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTitle(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFormModel_StringParam(t *testing.T) {
	script := core.Script{
		Name:        "test",
		Description: "Test script",
		Parameters: []core.Parameter{
			{
				Name:        "name",
				Type:        core.ParamTypeString,
				Required:    true,
				Description: "Name to greet",
			},
		},
	}

	model := NewFormModel(script, 80, 24)

	assert.Equal(t, script.Name, model.Script().Name)
	assert.False(t, model.Submitted())
	assert.False(t, model.Cancelled())
	assert.NotNil(t, model.form)
}

func TestNewFormModel_BoolParam(t *testing.T) {
	script := core.Script{
		Name:        "test",
		Description: "Test script",
		Parameters: []core.Parameter{
			{
				Name:        "verbose",
				Type:        core.ParamTypeBool,
				Default:     true,
				Description: "Enable verbose output",
			},
		},
	}

	model := NewFormModel(script, 80, 24)

	assert.NotNil(t, model.values["verbose"])
	// Check the default value was set
	if val, ok := model.values["verbose"].(*bool); ok {
		assert.True(t, *val)
	}
}

func TestNewFormModel_IntParam(t *testing.T) {
	script := core.Script{
		Name:        "test",
		Description: "Test script",
		Parameters: []core.Parameter{
			{
				Name:        "count",
				Type:        core.ParamTypeInt,
				Default:     10,
				Description: "Number of items",
			},
		},
	}

	model := NewFormModel(script, 80, 24)

	assert.NotNil(t, model.values["count"])
}

func TestNewFormModel_ChoicesParam(t *testing.T) {
	script := core.Script{
		Name:        "deploy",
		Description: "Deploy application",
		Parameters: []core.Parameter{
			{
				Name:        "env",
				Type:        core.ParamTypeString,
				Choices:     []any{"staging", "production"},
				Default:     "staging",
				Description: "Target environment",
			},
		},
	}

	model := NewFormModel(script, 80, 24)

	assert.NotNil(t, model.values["env"])
	if val, ok := model.values["env"].(*string); ok {
		assert.Equal(t, "staging", *val)
	}
}

func TestFormModel_CanSkip(t *testing.T) {
	tests := []struct {
		name     string
		params   []core.Parameter
		expected bool
	}{
		{
			"no params",
			nil,
			true,
		},
		{
			"all optional",
			[]core.Parameter{
				{Name: "opt1", Required: false},
				{Name: "opt2", Required: false},
			},
			true,
		},
		{
			"all with defaults",
			[]core.Parameter{
				{Name: "param1", Required: true, Default: "default"},
				{Name: "param2", Required: true, Default: 42},
			},
			true,
		},
		{
			"required without default",
			[]core.Parameter{
				{Name: "required", Required: true},
			},
			false,
		},
		{
			"mixed",
			[]core.Parameter{
				{Name: "opt", Required: false},
				{Name: "req", Required: true},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := core.Script{
				Name:       "test",
				Parameters: tt.params,
			}
			model := NewFormModel(script, 80, 24)
			assert.Equal(t, tt.expected, model.CanSkip())
		})
	}
}

func TestFormModel_DefaultValues(t *testing.T) {
	script := core.Script{
		Name:        "test",
		Description: "Test script",
		Parameters: []core.Parameter{
			{Name: "str", Type: core.ParamTypeString, Default: "hello"},
			{Name: "num", Type: core.ParamTypeInt, Default: 42},
			{Name: "flag", Type: core.ParamTypeBool, Default: true},
			{Name: "no_default", Type: core.ParamTypeString},
			{Name: "bool_no_default", Type: core.ParamTypeBool},
		},
	}

	model := NewFormModel(script, 80, 24)
	defaults := model.DefaultValues()

	assert.Equal(t, "hello", defaults["str"])
	assert.Equal(t, "42", defaults["num"])
	assert.Equal(t, "true", defaults["flag"])
	assert.Equal(t, "false", defaults["bool_no_default"])
	_, hasNoDefault := defaults["no_default"]
	assert.False(t, hasNoDefault)
}

func TestFormModel_View(t *testing.T) {
	script := core.Script{
		Name:        "greet",
		Description: "Greet someone",
		Parameters: []core.Parameter{
			{Name: "name", Type: core.ParamTypeString, Required: true},
		},
	}

	model := NewFormModel(script, 80, 24)
	view := model.View()

	// Should contain script name
	assert.Contains(t, view, "greet")
	// Should contain description
	assert.Contains(t, view, "Greet someone")
	// Should contain footer
	assert.Contains(t, view, "tab")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "esc")
}

func TestFormModel_WindowResize(t *testing.T) {
	script := core.Script{
		Name:       "test",
		Parameters: []core.Parameter{{Name: "x", Type: core.ParamTypeString}},
	}

	model := NewFormModel(script, 80, 24)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := model.Update(msg)
	m := updated.(FormModel)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestFormModel_EscCancels(t *testing.T) {
	script := core.Script{
		Name:       "test",
		Parameters: []core.Parameter{{Name: "x", Type: core.ParamTypeString}},
	}

	model := NewFormModel(script, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	updated, cmd := model.Update(msg)
	m := updated.(FormModel)

	assert.True(t, m.Cancelled())
	assert.NotNil(t, cmd)

	// Execute the cmd and check the message type
	result := cmd()
	_, ok := result.(FormCancelledMsg)
	assert.True(t, ok)
}

func TestFormModel_CtrlCQuits(t *testing.T) {
	script := core.Script{
		Name:       "test",
		Parameters: []core.Parameter{{Name: "x", Type: core.ParamTypeString}},
	}

	model := NewFormModel(script, 80, 24)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	// Should return quit command
	assert.NotNil(t, cmd)
}

func TestFormModel_Init(t *testing.T) {
	script := core.Script{
		Name:       "test",
		Parameters: []core.Parameter{{Name: "x", Type: core.ParamTypeString}},
	}

	model := NewFormModel(script, 80, 24)
	cmd := model.Init()

	// Init should return a command from the huh form
	assert.NotNil(t, cmd)
}

func TestAppModel_ScriptSelectedWithParams(t *testing.T) {
	categories := []core.Category{
		{
			Name: "test",
			Scripts: []core.Script{
				{
					Name: "parameterized",
					Parameters: []core.Parameter{
						{Name: "name", Type: core.ParamTypeString, Required: true},
					},
				},
			},
		},
	}
	model := NewAppModel(categories)

	// Send ScriptSelectedMsg for script with params
	msg := ScriptSelectedMsg{Script: categories[0].Scripts[0]}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	// Should transition to form state
	assert.Equal(t, StateForm, m.State())
}

func TestAppModel_ScriptSelectedWithoutParams(t *testing.T) {
	categories := []core.Category{
		{
			Name: "test",
			Scripts: []core.Script{
				{Name: "simple"},
			},
		},
	}
	model := NewAppModel(categories)

	// Send ScriptSelectedMsg for script without params
	msg := ScriptSelectedMsg{Script: categories[0].Scripts[0]}
	updated, cmd := model.Update(msg)
	m := updated.(AppModel)

	// Should not transition to form state
	assert.NotEqual(t, StateForm, m.State())
	// Should return quit command
	assert.NotNil(t, cmd)
	// Selected script should be set
	assert.NotNil(t, m.SelectedScript())
}

func TestAppModel_FormCancelled(t *testing.T) {
	model := NewAppModel(nil)
	model.state = StateForm

	msg := FormCancelledMsg{}
	updated, _ := model.Update(msg)
	m := updated.(AppModel)

	assert.Equal(t, StateScriptList, m.State())
}

func TestAppModel_FormSubmitted(t *testing.T) {
	model := NewAppModel(nil)
	model.state = StateForm

	script := core.Script{Name: "test"}
	params := map[string]string{"name": "value"}

	msg := FormSubmittedMsg{Script: script, Parameters: params}
	updated, cmd := model.Update(msg)
	m := updated.(AppModel)

	// Should set selected script and params
	assert.NotNil(t, m.SelectedScript())
	assert.Equal(t, "test", m.SelectedScript().Name)
	assert.Equal(t, params, m.SelectedParams())
	// Should return quit command
	assert.NotNil(t, cmd)
}

func TestAppModel_FormView(t *testing.T) {
	script := core.Script{
		Name:        "test",
		Description: "Test script",
		Parameters: []core.Parameter{
			{Name: "x", Type: core.ParamTypeString},
		},
	}

	model := NewAppModel(nil)
	model.state = StateForm
	model.formModel = NewFormModel(script, 80, 24)

	view := model.View()

	// Should render form view
	assert.Contains(t, view, "test")
	assert.Contains(t, view, "Test script")
}

func TestViewState_FormString(t *testing.T) {
	assert.Equal(t, "form", StateForm.String())
}
