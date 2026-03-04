package output

import (
	"fmt"
)

// Printer handles output formatting
type Printer struct {
	format   string
	useColor bool
}

// NewPrinter creates a new output printer
func NewPrinter(format string, useColor bool) *Printer {
	return &Printer{
		format:   format,
		useColor: useColor,
	}
}

// Print prints data in the specified format (placeholder for future use)
func (p *Printer) Print(data interface{}) error {
	return fmt.Errorf("print not implemented for format: %s", p.format)
}

// ValidateFormat checks if a format is supported
func ValidateFormat(format string, allowed []string) error {
	for _, f := range allowed {
		if f == format {
			return nil
		}
	}
	return fmt.Errorf("invalid format '%s', must be one of: %v", format, allowed)
}
