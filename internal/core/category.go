package core

import "sort"

// Category represents a grouping of scripts.
type Category struct {
	Name    string
	Scripts []Script
}

// OrganizeByCategory groups scripts by their category field.
// Categories are sorted alphabetically, with "uncategorized" appearing last.
func OrganizeByCategory(scripts []Script) []Category {
	categoryMap := make(map[string][]Script)

	for _, script := range scripts {
		cat := script.Category
		if cat == "" {
			cat = "uncategorized"
		}
		categoryMap[cat] = append(categoryMap[cat], script)
	}

	// Collect category names, excluding "uncategorized"
	var names []string
	for name := range categoryMap {
		if name != "uncategorized" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	// Build categories slice with "uncategorized" last
	categories := make([]Category, 0, len(categoryMap))
	for _, name := range names {
		categories = append(categories, Category{
			Name:    name,
			Scripts: categoryMap[name],
		})
	}

	// Add uncategorized at the end if it exists
	if scripts, ok := categoryMap["uncategorized"]; ok {
		categories = append(categories, Category{
			Name:    "uncategorized",
			Scripts: scripts,
		})
	}

	return categories
}
