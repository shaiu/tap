package tui

import (
	"strings"
	"testing"
)

func TestNewHelpModel(t *testing.T) {
	m := NewHelpModel()

	if m.width != 80 {
		t.Errorf("expected default width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("expected default height 24, got %d", m.height)
	}
}

func TestHelpModel_SetSize(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestHelpModel_Sections(t *testing.T) {
	m := NewHelpModel()
	sections := m.Sections()

	// Verify we have the expected sections
	expectedSections := []string{"Navigation", "Panels", "Filtering", "Actions"}

	if len(sections) != len(expectedSections) {
		t.Errorf("expected %d sections, got %d", len(expectedSections), len(sections))
	}

	for i, expected := range expectedSections {
		if sections[i].Title != expected {
			t.Errorf("section %d: expected title %q, got %q", i, expected, sections[i].Title)
		}
	}
}

func TestHelpModel_Sections_HasBindings(t *testing.T) {
	m := NewHelpModel()
	sections := m.Sections()

	for _, section := range sections {
		if len(section.Bindings) == 0 {
			t.Errorf("section %q has no bindings", section.Title)
		}

		for _, binding := range section.Bindings {
			if binding.Key == "" {
				t.Errorf("section %q has binding with empty key", section.Title)
			}
			if binding.Desc == "" {
				t.Errorf("section %q has binding with empty description", section.Title)
			}
		}
	}
}

func TestHelpModel_View(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(100, 40)

	view := m.View()

	// Check that view contains expected elements
	expectedContent := []string{
		"Keyboard Shortcuts",
		"Navigation",
		"Panels",
		"Filtering",
		"Actions",
		"↑ / k",
		"↓ / j",
		"enter",
		"esc",
		"tab",
		"shift+tab",
		"/",
		"r",
		"?",
		"q",
		"Press any key to close",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(view, expected) {
			t.Errorf("view should contain %q", expected)
		}
	}
}

func TestHelpModel_View_HasBorder(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(100, 40)

	view := m.View()

	// Check for rounded border characters
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╮") {
		t.Error("view should have rounded border (top corners)")
	}
	if !strings.Contains(view, "╰") || !strings.Contains(view, "╯") {
		t.Error("view should have rounded border (bottom corners)")
	}
}

func TestHelpModel_View_Centered(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(120, 50)

	view := m.View()
	lines := strings.Split(view, "\n")

	// The first lines should be empty (vertical padding)
	if len(lines) < 5 {
		t.Fatal("view should have multiple lines for centering")
	}

	// First non-empty line should have leading spaces (horizontal centering)
	foundContent := false
	for _, line := range lines {
		if strings.Contains(line, "╭") {
			foundContent = true
			if !strings.HasPrefix(line, " ") {
				t.Error("content should be horizontally centered with leading spaces")
			}
			break
		}
	}

	if !foundContent {
		t.Error("could not find border in view")
	}
}

func TestRenderHelp(t *testing.T) {
	view := RenderHelp(100, 40)

	// Should contain help content
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("RenderHelp should contain help content")
	}

	// Should have border
	if !strings.Contains(view, "╭") {
		t.Error("RenderHelp should have border")
	}
}

func TestHelpModel_View_SmallSize(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(30, 10) // Very small size

	// Should not panic
	view := m.View()

	// Should still render content
	if !strings.Contains(view, "Keyboard") {
		t.Error("small size should still render help content")
	}
}

func TestHelpSection_Navigation(t *testing.T) {
	m := NewHelpModel()
	sections := m.Sections()

	// Find Navigation section
	var navSection *HelpSection
	for i := range sections {
		if sections[i].Title == "Navigation" {
			navSection = &sections[i]
			break
		}
	}

	if navSection == nil {
		t.Fatal("Navigation section not found")
	}

	// Check for expected bindings
	expectedKeys := map[string]bool{
		"↑ / k": false,
		"↓ / j": false,
		"enter": false,
		"esc":   false,
	}

	for _, binding := range navSection.Bindings {
		if _, ok := expectedKeys[binding.Key]; ok {
			expectedKeys[binding.Key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Navigation section should have key %q", key)
		}
	}
}

func TestHelpSection_Panels(t *testing.T) {
	m := NewHelpModel()
	sections := m.Sections()

	// Find Panels section
	var panelsSection *HelpSection
	for i := range sections {
		if sections[i].Title == "Panels" {
			panelsSection = &sections[i]
			break
		}
	}

	if panelsSection == nil {
		t.Fatal("Panels section not found")
	}

	// Check for tab bindings
	hasTab := false
	hasShiftTab := false
	for _, binding := range panelsSection.Bindings {
		if binding.Key == "tab" {
			hasTab = true
		}
		if binding.Key == "shift+tab" {
			hasShiftTab = true
		}
	}

	if !hasTab {
		t.Error("Panels section should have tab binding")
	}
	if !hasShiftTab {
		t.Error("Panels section should have shift+tab binding")
	}
}

func TestHelpSection_Actions(t *testing.T) {
	m := NewHelpModel()
	sections := m.Sections()

	// Find Actions section
	var actionsSection *HelpSection
	for i := range sections {
		if sections[i].Title == "Actions" {
			actionsSection = &sections[i]
			break
		}
	}

	if actionsSection == nil {
		t.Fatal("Actions section not found")
	}

	// Check for expected bindings
	expectedKeys := map[string]bool{
		"r": false,
		"?": false,
		"q": false,
	}

	for _, binding := range actionsSection.Bindings {
		if _, ok := expectedKeys[binding.Key]; ok {
			expectedKeys[binding.Key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Actions section should have key %q", key)
		}
	}
}
