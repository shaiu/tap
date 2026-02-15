package tui

import (
	"strings"
	"testing"

	"github.com/shaiungar/tap/internal/core"
)

func TestNewDetailsModel(t *testing.T) {
	m := NewDetailsModel()

	// Should start with no script
	if m.Script() != nil {
		t.Error("expected no script initially")
	}

	// Should not be focused initially
	if m.IsFocused() {
		t.Error("expected not focused initially")
	}

	// Panel title should be "Details"
	if m.PanelTitle() != "Details" {
		t.Errorf("expected panel title to be 'Details', got %s", m.PanelTitle())
	}
}

func TestDetailsModel_SetScript(t *testing.T) {
	m := NewDetailsModel()

	script := &core.Script{
		Name:        "deploy.sh",
		Description: "Deploy application to production",
		Category:    "Deploy",
		Shell:       "bash",
		Path:        "/scripts/deploy.sh",
	}

	m.SetScript(script)

	if m.Script() == nil {
		t.Error("expected script to be set")
	}
	if m.Script().Name != "deploy.sh" {
		t.Errorf("expected script name to be 'deploy.sh', got %s", m.Script().Name)
	}

	// Clear script
	m.SetScript(nil)
	if m.Script() != nil {
		t.Error("expected script to be cleared")
	}
}

func TestDetailsModel_Focus(t *testing.T) {
	m := NewDetailsModel()

	// Initially not focused
	if m.IsFocused() {
		t.Error("expected not focused initially")
	}

	// Set focused
	m.SetFocused(true)
	if !m.IsFocused() {
		t.Error("expected to be focused")
	}

	// Set unfocused
	m.SetFocused(false)
	if m.IsFocused() {
		t.Error("expected to be unfocused")
	}
}

func TestDetailsModel_SetSize(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(40, 30)

	// Size should be stored (verified through View rendering)
	// Just ensure no panic occurs
	_ = m.View()
}

func TestDetailsModel_ViewNoScript(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(50, 20)

	view := m.View()

	// Should contain placeholder text
	if !strings.Contains(view, "Select a script") {
		t.Error("expected view to contain placeholder text when no script selected")
	}

	// Should contain "Details" title
	if !strings.Contains(view, "Details") {
		t.Error("expected view to contain 'Details' title")
	}
}

func TestDetailsModel_ViewWithScript(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)

	script := &core.Script{
		Name:        "backup.sh",
		Description: "Backup database to S3",
		Category:    "Database",
		Shell:       "bash",
		Path:        "/scripts/db/backup.sh",
	}
	m.SetScript(script)

	view := m.View()

	// Should contain script name
	if !strings.Contains(view, "backup.sh") {
		t.Error("expected view to contain script name")
	}

	// Should contain description
	if !strings.Contains(view, "Backup database") {
		t.Error("expected view to contain script description")
	}

	// Should contain category
	if !strings.Contains(view, "Category") && !strings.Contains(view, "Database") {
		t.Error("expected view to contain category metadata")
	}

	// Should contain shell
	if !strings.Contains(view, "Shell") || !strings.Contains(view, "bash") {
		t.Error("expected view to contain shell metadata")
	}
}

func TestDetailsModel_ViewWithParameters(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 40)

	script := &core.Script{
		Name:        "deploy.sh",
		Description: "Deploy application",
		Category:    "Deploy",
		Shell:       "bash",
		Path:        "/scripts/deploy.sh",
		Parameters: []core.Parameter{
			{
				Name:        "environment",
				Type:        "string",
				Required:    true,
				Description: "Target environment",
			},
			{
				Name:        "skip-tests",
				Type:        "bool",
				Required:    false,
				Default:     "false",
				Description: "Skip running tests",
			},
		},
	}
	m.SetScript(script)

	view := m.View()

	// Should contain "Parameters" header
	if !strings.Contains(view, "Parameters") {
		t.Error("expected view to contain 'Parameters' header")
	}

	// Should contain parameter names
	if !strings.Contains(view, "environment") {
		t.Error("expected view to contain 'environment' parameter")
	}
	if !strings.Contains(view, "skip-tests") {
		t.Error("expected view to contain 'skip-tests' parameter")
	}

	// Should show required indicator for environment
	if !strings.Contains(view, "*") {
		t.Error("expected view to contain required indicator (*)")
	}

	// Should show type info
	if !strings.Contains(view, "string") {
		t.Error("expected view to contain parameter type 'string'")
	}
	if !strings.Contains(view, "bool") {
		t.Error("expected view to contain parameter type 'bool'")
	}
}

func TestDetailsModel_ViewWithTags(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)

	script := &core.Script{
		Name:        "deploy.sh",
		Description: "Deploy application",
		Category:    "Deploy",
		Shell:       "bash",
		Path:        "/scripts/deploy.sh",
		Tags:        []string{"production", "critical"},
	}
	m.SetScript(script)

	view := m.View()

	// Should contain tags
	if !strings.Contains(view, "Tags") {
		t.Error("expected view to contain 'Tags' label")
	}
	if !strings.Contains(view, "production") {
		t.Error("expected view to contain 'production' tag")
	}
	if !strings.Contains(view, "critical") {
		t.Error("expected view to contain 'critical' tag")
	}
}

func TestDetailsModel_ViewFocusedBorder(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)
	m.SetScript(&core.Script{
		Name:     "test.sh",
		Shell:    "bash",
		Category: "Test",
		Path:     "/test.sh",
	})

	// Unfocused view
	m.SetFocused(false)
	unfocusedView := m.View()

	// Focused view
	m.SetFocused(true)
	focusedView := m.View()

	// Views should be different (different border colors, though hard to test directly)
	// At minimum, both should render without error
	if len(unfocusedView) == 0 {
		t.Error("expected unfocused view to have content")
	}
	if len(focusedView) == 0 {
		t.Error("expected focused view to have content")
	}
}

func TestDetailsModel_WrapText(t *testing.T) {
	m := NewDetailsModel()

	tests := []struct {
		name     string
		text     string
		width    int
		wantWrap bool
	}{
		{
			name:     "short text no wrap",
			text:     "Hello world",
			width:    50,
			wantWrap: false,
		},
		{
			name:     "long text wraps",
			text:     "This is a very long description that should definitely wrap to multiple lines when given a narrow width",
			width:    30,
			wantWrap: true,
		},
		{
			name:     "empty text",
			text:     "",
			width:    50,
			wantWrap: false,
		},
		{
			name:     "single word",
			text:     "deploy",
			width:    10,
			wantWrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.wrapText(tt.text, tt.width)
			hasNewlines := strings.Contains(result, "\n")
			if tt.wantWrap && !hasNewlines && len(tt.text) > tt.width {
				t.Errorf("expected text to wrap but it didn't: %q", result)
			}
			if !tt.wantWrap && hasNewlines {
				t.Errorf("expected no wrap but got newlines: %q", result)
			}
		})
	}
}

func TestDetailsModel_RenderParameter(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)

	tests := []struct {
		name  string
		param core.Parameter
		wants []string
	}{
		{
			name: "required param",
			param: core.Parameter{
				Name:        "target",
				Type:        "string",
				Required:    true,
				Description: "Target host",
			},
			wants: []string{"target", "*", "string", "Target host"},
		},
		{
			name: "optional with default",
			param: core.Parameter{
				Name:        "port",
				Type:        "int",
				Required:    false,
				Default:     "8080",
				Description: "Port number",
			},
			wants: []string{"port", "int", "8080", "Port number"},
		},
		{
			name: "bool param",
			param: core.Parameter{
				Name:     "verbose",
				Type:     "bool",
				Required: false,
				Default:  "false",
			},
			wants: []string{"verbose", "bool", "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderParameter(tt.param, 50)
			for _, want := range tt.wants {
				if !strings.Contains(result, want) {
					t.Errorf("expected parameter render to contain %q, got: %s", want, result)
				}
			}
		})
	}
}

func TestDetailsModel_ViewInteractiveScript(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 40)

	script := &core.Script{
		Name:        "interactive-tool",
		Description: "A tool with its own prompts",
		Category:    "Tools",
		Shell:       "bash",
		Path:        "/scripts/interactive-tool.sh",
		Interactive: true,
	}
	m.SetScript(script)

	view := m.View()

	// Should contain "Mode" label and "Interactive" value
	if !strings.Contains(view, "Mode") {
		t.Error("expected view to contain 'Mode' label for interactive script")
	}
	if !strings.Contains(view, "Interactive") {
		t.Error("expected view to contain 'Interactive' mode value")
	}
}

func TestDetailsModel_ViewNonInteractiveScript(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 40)

	script := &core.Script{
		Name:        "deploy",
		Description: "Deploy application",
		Category:    "Deploy",
		Shell:       "bash",
		Path:        "/scripts/deploy.sh",
		Interactive: false,
	}
	m.SetScript(script)

	view := m.View()

	// Should NOT contain "Interactive" mode indicator
	if strings.Contains(view, "Interactive") {
		t.Error("expected view to NOT contain 'Interactive' for non-interactive script")
	}
}

func TestDetailsModel_ViewWithPythonScript(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)

	script := &core.Script{
		Name:        "analyze.py",
		Description: "Analyze data files",
		Category:    "Utils",
		Shell:       "python",
		Path:        "/scripts/analyze.py",
	}
	m.SetScript(script)

	view := m.View()

	// Should contain script name
	if !strings.Contains(view, "analyze.py") {
		t.Error("expected view to contain script name")
	}

	// Should contain shell info
	if !strings.Contains(view, "python") {
		t.Error("expected view to contain 'python' shell")
	}
}

func TestDetailsModel_Init(t *testing.T) {
	m := NewDetailsModel()

	// Init should return nil (no initial command)
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestDetailsModel_Update(t *testing.T) {
	m := NewDetailsModel()

	// Update is a no-op for details panel (read-only)
	// Just verify it doesn't panic and returns nil command
	_, cmd := m.Update(nil)
	if cmd != nil {
		t.Error("expected Update to return nil command")
	}
}

func TestDetailsModel_ViewWithLongPath(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(50, 30)

	script := &core.Script{
		Name:        "deploy.sh",
		Description: "Deploy",
		Category:    "Deploy",
		Shell:       "bash",
		Path:        "/very/long/path/that/goes/on/and/on/with/many/directories/scripts/deploy.sh",
	}
	m.SetScript(script)

	view := m.View()

	// View should render without error
	if len(view) == 0 {
		t.Error("expected view to have content")
	}
	// Path should be truncated or shortened
	if !strings.Contains(view, "Path") {
		t.Error("expected view to contain 'Path' label")
	}
}

func TestDetailsModel_ViewWithNoDescription(t *testing.T) {
	m := NewDetailsModel()
	m.SetSize(60, 30)

	script := &core.Script{
		Name:        "simple.sh",
		Description: "",
		Category:    "Utils",
		Shell:       "bash",
		Path:        "/scripts/simple.sh",
	}
	m.SetScript(script)

	view := m.View()

	// Should still render properly
	if !strings.Contains(view, "simple.sh") {
		t.Error("expected view to contain script name")
	}
	// Should contain category even without description
	if !strings.Contains(view, "Category") {
		t.Error("expected view to contain 'Category' label")
	}
}
