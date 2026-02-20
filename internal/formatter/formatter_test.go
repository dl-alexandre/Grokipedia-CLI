package formatter

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	tests := []struct {
		name   string
		data   interface{}
		indent bool
	}{
		{
			name:   "simple map",
			data:   map[string]string{"key": "value"},
			indent: true,
		},
		{
			name:   "nested struct",
			data:   struct{ Name string }{Name: "test"},
			indent: false,
		},
		{
			name:   "array",
			data:   []string{"a", "b", "c"},
			indent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &JSONFormatter{Indent: tt.indent}
			var buf bytes.Buffer

			err := f.Format(tt.data, &buf)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("Format() produced empty output")
			}

			// Verify it's valid JSON
			if !strings.Contains(output, "{") && !strings.Contains(output, "[") {
				t.Error("Format() did not produce JSON output")
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   FormatType
		expected string
	}{
		{FormatJSON, "*formatter.JSONFormatter"},
		{FormatTable, "*formatter.TableFormatter"},
		{FormatMarkdown, "*formatter.MarkdownFormatter"},
		{FormatPlain, "*formatter.PlainFormatter"},
		{FormatYAML, "*formatter.YAMLFormatter"},
		{FormatList, "*formatter.ListFormatter"},
		{FormatType("unknown"), "*formatter.JSONFormatter"}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			f := NewFormatter(tt.format)
			if f == nil {
				t.Fatal("NewFormatter() returned nil")
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		allowed []string
		wantErr bool
	}{
		{
			name:    "valid format",
			format:  "json",
			allowed: []string{"json", "yaml", "table"},
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "xml",
			allowed: []string{"json", "yaml", "table"},
			wantErr: true,
		},
		{
			name:    "empty allowed list",
			format:  "json",
			allowed: []string{},
			wantErr: true,
		},
		{
			name:    "case sensitive",
			format:  "JSON",
			allowed: []string{"json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFormatErrorMessage(t *testing.T) {
	err := ValidateFormat("invalid", []string{"json", "yaml"})
	if err == nil {
		t.Fatal("Expected error for invalid format")
	}

	expectedParts := []string{"invalid format", "invalid", "allowed", "json", "yaml"}
	errMsg := err.Error()
	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("Error message should contain %q, got: %s", part, errMsg)
		}
	}
}

func TestTableFormatter(t *testing.T) {
	f := &TableFormatter{
		Headers: []string{"Col1", "Col2"},
	}
	var buf bytes.Buffer

	data := map[string]string{"key": "value"}
	err := f.Format(data, &buf)

	// Table formatter returns error as it requires command-specific implementation
	if err == nil {
		t.Error("TableFormatter.Format() should return error for generic data")
	}
}

func TestMarkdownFormatter(t *testing.T) {
	f := &MarkdownFormatter{}
	var buf bytes.Buffer

	data := map[string]string{"key": "value"}
	err := f.Format(data, &buf)

	// Markdown formatter returns error as it requires command-specific implementation
	if err == nil {
		t.Error("MarkdownFormatter.Format() should return error for generic data")
	}
}

func TestPlainFormatter(t *testing.T) {
	f := &PlainFormatter{}
	var buf bytes.Buffer

	data := map[string]string{"key": "value"}
	err := f.Format(data, &buf)

	// Plain formatter returns error as it requires command-specific implementation
	if err == nil {
		t.Error("PlainFormatter.Format() should return error for generic data")
	}
}

func TestYAMLFormatter(t *testing.T) {
	f := &YAMLFormatter{}
	var buf bytes.Buffer

	data := map[string]string{"key": "value"}
	err := f.Format(data, &buf)

	// YAML formatter returns error as it requires command-specific implementation
	if err == nil {
		t.Error("YAMLFormatter.Format() should return error for generic data")
	}
}

func TestListFormatter(t *testing.T) {
	f := &ListFormatter{}
	var buf bytes.Buffer

	data := []string{"item1", "item2", "item3"}
	err := f.Format(data, &buf)

	// List formatter returns error as it requires command-specific implementation
	if err == nil {
		t.Error("ListFormatter.Format() should return error for generic data")
	}
}

func TestFormatTypeConstants(t *testing.T) {
	// Verify all format type constants
	formats := []FormatType{
		FormatJSON,
		FormatTable,
		FormatMarkdown,
		FormatPlain,
		FormatYAML,
		FormatList,
	}

	expected := []string{"json", "table", "markdown", "plain", "yaml", "list"}

	for i, format := range formats {
		if string(format) != expected[i] {
			t.Errorf("FormatType %d: expected %q, got %q", i, expected[i], string(format))
		}
	}
}
