package core

import "testing"

func TestValidParamTypes(t *testing.T) {
	types := ValidParamTypes()
	expected := []string{"string", "int", "float", "bool", "path"}

	if len(types) != len(expected) {
		t.Errorf("expected %d types, got %d", len(expected), len(types))
	}

	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("expected types[%d] = %q, got %q", i, expected[i], typ)
		}
	}
}

func TestIsValidParamType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"string", true},
		{"int", true},
		{"float", true},
		{"bool", true},
		{"path", true},
		{"invalid", false},
		{"", false},
		{"String", false}, // case-sensitive
		{"INT", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValidParamType(tt.input)
			if got != tt.want {
				t.Errorf("IsValidParamType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
