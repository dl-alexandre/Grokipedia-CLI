package formatter

import (
	"encoding/json"
	"fmt"
	"io"
)

// Formatter is the interface for all output formatters
type Formatter interface {
	Format(data interface{}, w io.Writer) error
}

// JSONFormatter outputs data as JSON
type JSONFormatter struct {
	Indent bool
}

// Format implements the Formatter interface
func (f *JSONFormatter) Format(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// TableFormatter outputs data as an ASCII table
type TableFormatter struct {
	Headers []string
}

// Format implements the Formatter interface
func (f *TableFormatter) Format(data interface{}, w io.Writer) error {
	// Table formatting is command-specific
	// Each command will implement its own table formatting
	return fmt.Errorf("table formatter requires command-specific implementation")
}

// MarkdownFormatter outputs data as Markdown
type MarkdownFormatter struct{}

// Format implements the Formatter interface
func (f *MarkdownFormatter) Format(data interface{}, w io.Writer) error {
	// Markdown formatting is command-specific
	return fmt.Errorf("markdown formatter requires command-specific implementation")
}

// PlainFormatter outputs data as plain text
type PlainFormatter struct{}

// Format implements the Formatter interface
func (f *PlainFormatter) Format(data interface{}, w io.Writer) error {
	// Plain text formatting is command-specific
	return fmt.Errorf("plain formatter requires command-specific implementation")
}

// YAMLFormatter outputs data as YAML
type YAMLFormatter struct{}

// Format implements the Formatter interface
func (f *YAMLFormatter) Format(data interface{}, w io.Writer) error {
	// YAML formatting uses the yaml package
	// This is a placeholder - actual implementation would use gopkg.in/yaml.v3
	return fmt.Errorf("yaml formatter requires command-specific implementation")
}

// ListFormatter outputs data as a newline-delimited list
type ListFormatter struct{}

// Format implements the Formatter interface
func (f *ListFormatter) Format(data interface{}, w io.Writer) error {
	// List formatting is command-specific
	return fmt.Errorf("list formatter requires command-specific implementation")
}

// FormatType represents the output format type
type FormatType string

const (
	FormatJSON     FormatType = "json"
	FormatTable    FormatType = "table"
	FormatMarkdown FormatType = "markdown"
	FormatPlain    FormatType = "plain"
	FormatYAML     FormatType = "yaml"
	FormatList     FormatType = "list"
)

// NewFormatter creates a formatter for the given format type
func NewFormatter(format FormatType) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{Indent: true}
	case FormatTable:
		return &TableFormatter{}
	case FormatMarkdown:
		return &MarkdownFormatter{}
	case FormatPlain:
		return &PlainFormatter{}
	case FormatYAML:
		return &YAMLFormatter{}
	case FormatList:
		return &ListFormatter{}
	default:
		return &JSONFormatter{Indent: true}
	}
}

// ValidateFormat checks if a format is valid for a command
func ValidateFormat(format string, allowed []string) error {
	for _, f := range allowed {
		if f == format {
			return nil
		}
	}
	return fmt.Errorf("invalid format '%s'; allowed: %v", format, allowed)
}
