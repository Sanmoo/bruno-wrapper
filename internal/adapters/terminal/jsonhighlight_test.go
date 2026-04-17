package terminal

import (
	"strings"
	"testing"
)

func TestHighlightJSONNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	input := `{"name": "Brazil", "id": 2032}`
	result := highlightJSON(input)
	if result != input {
		t.Errorf("highlightJSON() with NO_COLOR = %q, want %q", result, input)
	}
}

func TestHighlightJSONPreservesStructure(t *testing.T) {
	input := `{"a": 1}`
	result := highlightJSON(input)

	for _, char := range []string{"{", "}", ":", "1", "a"} {
		if !strings.Contains(result, char) {
			t.Errorf("highlightJSON() missing %q in result: %q", char, result)
		}
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"-42", true},
		{"3.14", true},
		{"1e10", true},
		{"1E-5", true},
		{"1.5e+3", true},
		{"abc", false},
		{"", false},
		{"-", false},
		{"true", false},
		{"null", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumber(tt.input)
			if got != tt.want {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHighlightJSONMultiline(t *testing.T) {
	input := `{
  "name": "Brazil",
  "code": "BRA",
  "id": 2032
}`
	result := highlightJSON(input)

	if !strings.Contains(result, "\n") {
		t.Error("highlightJSON() should preserve newlines")
	}

	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("highlightJSON() should have 5 lines, got %d", len(lines))
	}
}
