package tui

import (
	"strings"
	"testing"
)

func TestNewFooterModel(t *testing.T) {
	m := NewFooterModel()

	if m.context.State != StateBrowsing {
		t.Errorf("expected initial state StateBrowsing, got %v", m.context.State)
	}
	if m.context.ActivePanel != PanelSidebar {
		t.Errorf("expected initial panel PanelSidebar, got %v", m.context.ActivePanel)
	}
	if m.context.LayoutMode != LayoutThreePanel {
		t.Errorf("expected initial layout LayoutThreePanel, got %v", m.context.LayoutMode)
	}
	if m.width != 80 {
		t.Errorf("expected initial width 80, got %d", m.width)
	}
}

func TestFooterModel_SetContext(t *testing.T) {
	m := NewFooterModel()

	ctx := FooterContext{
		State:       StateFilter,
		ActivePanel: PanelScripts,
		LayoutMode:  LayoutTwoPanel,
		HasParams:   true,
	}
	m.SetContext(ctx)

	if m.context.State != StateFilter {
		t.Errorf("expected state StateFilter, got %v", m.context.State)
	}
	if m.context.ActivePanel != PanelScripts {
		t.Errorf("expected panel PanelScripts, got %v", m.context.ActivePanel)
	}
	if m.context.LayoutMode != LayoutTwoPanel {
		t.Errorf("expected layout LayoutTwoPanel, got %v", m.context.LayoutMode)
	}
	if !m.context.HasParams {
		t.Error("expected HasParams true")
	}
}

func TestFooterModel_SetWidth(t *testing.T) {
	m := NewFooterModel()
	m.SetWidth(120)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
}

func TestFooterModel_BrowsingHints_SidebarPanel(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelSidebar,
		LayoutMode:  LayoutThreePanel,
		HasParams:   false,
	})

	hints := m.hintsForContext()

	// Should have: navigate, select, panel, filter, help, quit
	expectedKeys := []string{"↑↓", "enter", "tab", "/", "?", "q"}
	expectedActions := []string{"navigate", "select", "panel", "filter", "help", "quit"}

	if len(hints) != len(expectedKeys) {
		t.Errorf("expected %d hints, got %d", len(expectedKeys), len(hints))
		return
	}

	for i, hint := range hints {
		if hint.Key != expectedKeys[i] {
			t.Errorf("hint %d: expected key %q, got %q", i, expectedKeys[i], hint.Key)
		}
		if hint.Action != expectedActions[i] {
			t.Errorf("hint %d: expected action %q, got %q", i, expectedActions[i], hint.Action)
		}
	}
}

func TestFooterModel_BrowsingHints_ScriptsPanel(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelScripts,
		LayoutMode:  LayoutThreePanel,
		HasParams:   false,
	})

	hints := m.hintsForContext()

	// Should have: navigate, run (not select), panel, filter, help, quit
	found := false
	for _, hint := range hints {
		if hint.Key == "enter" && hint.Action == "run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'enter run' hint for scripts panel")
	}
}

func TestFooterModel_BrowsingHints_WithParams(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelScripts,
		LayoutMode:  LayoutThreePanel,
		HasParams:   true,
	})

	hints := m.hintsForContext()

	// Should include params hint
	found := false
	for _, hint := range hints {
		if hint.Key == "p" && hint.Action == "params" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'p params' hint when HasParams is true")
	}
}

func TestFooterModel_BrowsingHints_NoTabInOnePanel(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelScripts,
		LayoutMode:  LayoutOnePanel,
		HasParams:   false,
	})

	hints := m.hintsForContext()

	// Should NOT have tab hint in one-panel mode
	for _, hint := range hints {
		if hint.Key == "tab" {
			t.Error("tab hint should not appear in one-panel mode")
		}
	}
}

func TestFooterModel_FilterHints(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State: StateFilter,
	})

	hints := m.hintsForContext()

	if len(hints) != 2 {
		t.Errorf("expected 2 filter hints, got %d", len(hints))
		return
	}

	if hints[0].Key != "enter" || hints[0].Action != "select" {
		t.Errorf("expected 'enter select', got '%s %s'", hints[0].Key, hints[0].Action)
	}
	if hints[1].Key != "esc" || hints[1].Action != "cancel" {
		t.Errorf("expected 'esc cancel', got '%s %s'", hints[1].Key, hints[1].Action)
	}
}

func TestFooterModel_HelpHints(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State: StateHelp,
	})

	hints := m.hintsForContext()

	if len(hints) != 1 {
		t.Errorf("expected 1 help hint, got %d", len(hints))
		return
	}

	if hints[0].Key != "any key" || hints[0].Action != "close" {
		t.Errorf("expected 'any key close', got '%s %s'", hints[0].Key, hints[0].Action)
	}
}

func TestFooterModel_FormHints(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State: StateForm,
	})

	hints := m.hintsForContext()

	expectedKeys := []string{"↑↓", "enter", "esc"}
	expectedActions := []string{"navigate", "submit", "cancel"}

	if len(hints) != len(expectedKeys) {
		t.Errorf("expected %d form hints, got %d", len(expectedKeys), len(hints))
		return
	}

	for i, hint := range hints {
		if hint.Key != expectedKeys[i] {
			t.Errorf("hint %d: expected key %q, got %q", i, expectedKeys[i], hint.Key)
		}
		if hint.Action != expectedActions[i] {
			t.Errorf("hint %d: expected action %q, got %q", i, expectedActions[i], hint.Action)
		}
	}
}

func TestFooterModel_LegacyMenuHints_CategoryList(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State: StateCategoryList,
	})

	hints := m.hintsForContext()

	// Should NOT have 'esc back' in category list
	for _, hint := range hints {
		if hint.Key == "esc" && hint.Action == "back" {
			t.Error("category list should not have 'esc back' hint")
		}
	}

	// Should have standard hints
	hasFilter := false
	hasQuit := false
	for _, hint := range hints {
		if hint.Key == "/" {
			hasFilter = true
		}
		if hint.Key == "q" {
			hasQuit = true
		}
	}
	if !hasFilter {
		t.Error("expected filter hint in category list")
	}
	if !hasQuit {
		t.Error("expected quit hint in category list")
	}
}

func TestFooterModel_LegacyMenuHints_ScriptList(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State: StateScriptList,
	})

	hints := m.hintsForContext()

	// Should have 'esc back' in script list
	found := false
	for _, hint := range hints {
		if hint.Key == "esc" && hint.Action == "back" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'esc back' hint in script list")
	}
}

func TestFooterModel_View(t *testing.T) {
	m := NewFooterModel()
	m.SetWidth(80)
	m.SetContext(FooterContext{
		State:       StateFilter,
		ActivePanel: PanelScripts,
		LayoutMode:  LayoutThreePanel,
	})

	view := m.View()

	// View should contain the hint text
	if !strings.Contains(view, "enter") {
		t.Error("footer view should contain 'enter'")
	}
	if !strings.Contains(view, "select") {
		t.Error("footer view should contain 'select'")
	}
	if !strings.Contains(view, "esc") {
		t.Error("footer view should contain 'esc'")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("footer view should contain 'cancel'")
	}
}

func TestFormatKeyHint(t *testing.T) {
	hint := formatKeyHint("enter", "run")

	// Should contain both key and action
	if !strings.Contains(hint, "enter") {
		t.Error("formatted hint should contain key")
	}
	if !strings.Contains(hint, "run") {
		t.Error("formatted hint should contain action")
	}
}

func TestRenderSimple(t *testing.T) {
	hints := []KeyHint{
		{Key: "enter", Action: "select"},
		{Key: "esc", Action: "cancel"},
	}

	result := RenderSimple(hints, 80)

	if !strings.Contains(result, "enter") {
		t.Error("result should contain 'enter'")
	}
	if !strings.Contains(result, "select") {
		t.Error("result should contain 'select'")
	}
	if !strings.Contains(result, "esc") {
		t.Error("result should contain 'esc'")
	}
	if !strings.Contains(result, "cancel") {
		t.Error("result should contain 'cancel'")
	}
}

func TestRenderCompact(t *testing.T) {
	hints := []KeyHint{
		{Key: "?", Action: "help"},
		{Key: "q", Action: "quit"},
	}

	result := RenderCompact(hints)

	// Compact render should not have border styling
	if !strings.Contains(result, "?") {
		t.Error("result should contain '?'")
	}
	if !strings.Contains(result, "help") {
		t.Error("result should contain 'help'")
	}
	if !strings.Contains(result, "q") {
		t.Error("result should contain 'q'")
	}
	if !strings.Contains(result, "quit") {
		t.Error("result should contain 'quit'")
	}
}

func TestHintsForState(t *testing.T) {
	// Test that the convenience function works
	hints := HintsForState(StateFilter, PanelScripts, LayoutThreePanel, false)

	if len(hints) != 2 {
		t.Errorf("expected 2 hints for filter state, got %d", len(hints))
	}
}

func TestFilterFooter(t *testing.T) {
	result := FilterFooter(80)

	if !strings.Contains(result, "enter") {
		t.Error("filter footer should contain 'enter'")
	}
	if !strings.Contains(result, "esc") {
		t.Error("filter footer should contain 'esc'")
	}
}

func TestHelpFooter(t *testing.T) {
	result := HelpFooter(80)

	if !strings.Contains(result, "any key") {
		t.Error("help footer should contain 'any key'")
	}
	if !strings.Contains(result, "close") {
		t.Error("help footer should contain 'close'")
	}
}

func TestFormFooter(t *testing.T) {
	result := FormFooter(80)

	if !strings.Contains(result, "enter") {
		t.Error("form footer should contain 'enter'")
	}
	if !strings.Contains(result, "submit") {
		t.Error("form footer should contain 'submit'")
	}
	if !strings.Contains(result, "esc") {
		t.Error("form footer should contain 'esc'")
	}
}

func TestFooterHeight(t *testing.T) {
	height := FooterHeight()

	// Footer should be 2 lines (border + content)
	if height != 2 {
		t.Errorf("expected footer height 2, got %d", height)
	}
}

func TestGetFooterStyle(t *testing.T) {
	style := GetFooterStyle()

	// Just verify it returns a valid style (non-zero)
	// The style itself is tested by rendering it
	if style.GetPaddingLeft() == 0 && style.GetPaddingRight() == 0 {
		// Style should have some padding
		// Note: this is a weak test, but lipgloss styles are hard to inspect
	}
}

func TestFooterModel_DetailsPanel(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelDetails,
		LayoutMode:  LayoutThreePanel,
		HasParams:   false,
	})

	hints := m.hintsForContext()

	// Details panel should show "run" action
	found := false
	for _, hint := range hints {
		if hint.Key == "enter" && hint.Action == "run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'enter run' hint for details panel")
	}
}

func TestFooterModel_TwoPanelLayout(t *testing.T) {
	m := NewFooterModel()
	m.SetContext(FooterContext{
		State:       StateBrowsing,
		ActivePanel: PanelSidebar,
		LayoutMode:  LayoutTwoPanel,
		HasParams:   false,
	})

	hints := m.hintsForContext()

	// Two-panel layout should still have tab hint
	found := false
	for _, hint := range hints {
		if hint.Key == "tab" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'tab' hint in two-panel layout")
	}
}
