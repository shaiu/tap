// Package core provides the core data structures and business logic for tap.
package core

// ParamType constants define the supported parameter types.
const (
	ParamTypeString = "string"
	ParamTypeInt    = "int"
	ParamTypeFloat  = "float"
	ParamTypeBool   = "bool"
	ParamTypePath   = "path"
)

// Script represents a tap-enabled script with its metadata.
type Script struct {
	// Identity
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`

	// Metadata
	Author      string   `yaml:"author,omitempty"`
	Version     string   `yaml:"version,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Interactive bool     `yaml:"interactive,omitempty"`

	// Parameters
	Parameters []Parameter `yaml:"parameters,omitempty"`

	// Examples
	Examples []Example `yaml:"examples,omitempty"`

	// Runtime (not from YAML)
	Path    string `yaml:"-"` // Absolute path to script file
	Shell   string `yaml:"-"` // Detected shell (bash, python, etc.)
	Source  string `yaml:"-"` // "scanned" or "registered"
	AutoGen bool   `yaml:"-"` // true if metadata was auto-generated
}

// Parameter represents a script parameter definition.
type Parameter struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`              // string, int, float, bool, path
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default,omitempty"`
	Choices     []any  `yaml:"choices,omitempty"`
	Description string `yaml:"description,omitempty"`
	Short       string `yaml:"short,omitempty"`
}

// Example represents a usage example for a script.
type Example struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// ValidParamTypes returns the list of valid parameter types.
func ValidParamTypes() []string {
	return []string{
		ParamTypeString,
		ParamTypeInt,
		ParamTypeFloat,
		ParamTypeBool,
		ParamTypePath,
	}
}

// IsValidParamType checks if the given type is a valid parameter type.
func IsValidParamType(t string) bool {
	for _, valid := range ValidParamTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
