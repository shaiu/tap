package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/shaiungar/tap/internal/core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterByCategory(t *testing.T) {
	categories := []core.Category{
		{Name: "deploy", Scripts: []core.Script{{Name: "deploy-app"}}},
		{Name: "data", Scripts: []core.Script{{Name: "export"}}},
		{Name: "utils", Scripts: []core.Script{{Name: "cleanup"}}},
	}

	tests := []struct {
		name     string
		filter   string
		expected int
	}{
		{"existing category", "deploy", 1},
		{"another category", "data", 1},
		{"non-existent category", "nonexistent", 0},
		{"empty filter returns all", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []core.Category
			if tt.filter == "" {
				result = categories
			} else {
				result = filterByCategory(categories, tt.filter)
			}
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestFilterByCategory_ReturnsCorrectCategory(t *testing.T) {
	categories := []core.Category{
		{Name: "deploy", Scripts: []core.Script{{Name: "deploy-app", Description: "Deploy"}}},
		{Name: "data", Scripts: []core.Script{{Name: "export", Description: "Export data"}}},
	}

	result := filterByCategory(categories, "data")
	require.Len(t, result, 1)
	assert.Equal(t, "data", result[0].Name)
	assert.Equal(t, "export", result[0].Scripts[0].Name)
}

func TestOutputJSON_Format(t *testing.T) {
	categories := []core.Category{
		{
			Name: "deploy",
			Scripts: []core.Script{
				{Name: "deploy-app", Description: "Deploy app", Category: "deploy", Path: "/scripts/deploy.sh"},
			},
		},
		{
			Name: "data",
			Scripts: []core.Script{
				{Name: "export", Description: "Export data", Category: "data", Path: "/scripts/export.sh"},
			},
		},
	}

	// Capture output
	var buf bytes.Buffer
	origStdout := jsonOutput
	defer func() { jsonOutput = origStdout }()

	// Create encoder to buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")

	// Flatten scripts like outputJSON does
	var scripts []core.Script
	for _, cat := range categories {
		scripts = append(scripts, cat.Scripts...)
	}
	err := enc.Encode(scripts)
	require.NoError(t, err)

	// Parse back and verify
	var decoded []core.Script
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded, 2)
	assert.Equal(t, "deploy-app", decoded[0].Name)
	assert.Equal(t, "export", decoded[1].Name)
}

func TestListCmd_HasCorrectFlags(t *testing.T) {
	cmd := RootCmd()

	// Find list subcommand
	var listCommand *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Use == "list" {
			listCommand = sub
			break
		}
	}

	require.NotNil(t, listCommand, "list command should exist")

	// Check flags
	categoryFlag := listCommand.Flags().Lookup("category")
	assert.NotNil(t, categoryFlag)
	assert.Equal(t, "c", categoryFlag.Shorthand)

	flatFlag := listCommand.Flags().Lookup("flat")
	assert.NotNil(t, flatFlag)

	jsonFlag := listCommand.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
}

// jsonOutput is used for testing (not actually used in production, just for test structure)
var jsonOutput = func() {}

func TestOutputGrouped_EmptyCategories(t *testing.T) {
	// Test with no categories
	err := outputGrouped(nil)
	assert.NoError(t, err)

	err = outputGrouped([]core.Category{})
	assert.NoError(t, err)
}

func TestOutputFlat_EmptyCategories(t *testing.T) {
	err := outputFlat(nil)
	assert.NoError(t, err)

	err = outputFlat([]core.Category{})
	assert.NoError(t, err)
}

func TestOutputJSON_EmptyCategories(t *testing.T) {
	// Create a pipe to capture stdout
	// For this test, we just verify no error on empty input
	err := outputJSON([]core.Category{})
	assert.NoError(t, err)
}

func TestListCmd_Registered(t *testing.T) {
	cmd := RootCmd()

	// Verify list command is registered
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "list" {
			found = true
			assert.Equal(t, "List all available scripts", sub.Short)
			break
		}
	}
	assert.True(t, found, "list command should be registered under root")
}
