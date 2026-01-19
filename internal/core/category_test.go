package core

import "testing"

func TestOrganizeByCategory(t *testing.T) {
	scripts := []Script{
		{Name: "deploy", Category: "deployment"},
		{Name: "backup", Category: "operations"},
		{Name: "lint", Category: "development"},
		{Name: "misc1", Category: ""},         // empty becomes uncategorized
		{Name: "misc2", Category: "uncategorized"},
		{Name: "rollback", Category: "deployment"},
	}

	categories := OrganizeByCategory(scripts)

	// Expect 4 categories: deployment, development, operations, uncategorized (last)
	if len(categories) != 4 {
		t.Fatalf("expected 4 categories, got %d", len(categories))
	}

	// Check alphabetical order (excluding uncategorized)
	expectedOrder := []string{"deployment", "development", "operations", "uncategorized"}
	for i, cat := range categories {
		if cat.Name != expectedOrder[i] {
			t.Errorf("category[%d] = %q, want %q", i, cat.Name, expectedOrder[i])
		}
	}

	// Check deployment has 2 scripts
	if len(categories[0].Scripts) != 2 {
		t.Errorf("deployment category should have 2 scripts, got %d", len(categories[0].Scripts))
	}

	// Check uncategorized has 2 scripts (empty string + "uncategorized")
	if len(categories[3].Scripts) != 2 {
		t.Errorf("uncategorized category should have 2 scripts, got %d", len(categories[3].Scripts))
	}
}

func TestOrganizeByCategory_Empty(t *testing.T) {
	categories := OrganizeByCategory([]Script{})

	if len(categories) != 0 {
		t.Errorf("expected 0 categories for empty input, got %d", len(categories))
	}
}

func TestOrganizeByCategory_AllUncategorized(t *testing.T) {
	scripts := []Script{
		{Name: "a", Category: "uncategorized"},
		{Name: "b", Category: ""},
	}

	categories := OrganizeByCategory(scripts)

	if len(categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(categories))
	}

	if categories[0].Name != "uncategorized" {
		t.Errorf("expected category name 'uncategorized', got %q", categories[0].Name)
	}

	if len(categories[0].Scripts) != 2 {
		t.Errorf("expected 2 scripts in uncategorized, got %d", len(categories[0].Scripts))
	}
}
