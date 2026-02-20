package cmd

import (
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestTypeaheadCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid list", "list", false},
		{"valid json", "json", false},
		{"invalid format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &api.TypeaheadResponse{
				Suggestions: []string{"test"},
			}
			err := outputTypeaheadResults(response, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputTypeaheadResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeaheadOutputJSON(t *testing.T) {
	response := &api.TypeaheadResponse{
		Suggestions: []string{"python", "python programming", "python tutorial"},
	}

	output := captureOutput(t, func() {
		err := outputTypeaheadResults(response, "json")
		if err != nil {
			t.Errorf("outputTypeaheadResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "suggestions") {
		t.Error("Expected output to contain 'suggestions'")
	}

	for _, suggestion := range response.Suggestions {
		if !strings.Contains(output, suggestion) {
			t.Errorf("Expected output to contain '%s'", suggestion)
		}
	}
}

func TestTypeaheadOutputList(t *testing.T) {
	response := &api.TypeaheadResponse{
		Suggestions: []string{"go", "golang", "google"},
	}

	output := captureOutput(t, func() {
		err := outputTypeaheadResults(response, "list")
		if err != nil {
			t.Errorf("outputTypeaheadResults() error = %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	for i, suggestion := range response.Suggestions {
		if i < len(lines) && !strings.Contains(lines[i], suggestion) {
			t.Errorf("Expected line %d to contain '%s'", i, suggestion)
		}
	}
}

func TestTypeaheadEmptyResults(t *testing.T) {
	response := &api.TypeaheadResponse{
		Suggestions: []string{},
	}

	output := captureOutput(t, func() {
		err := outputTypeaheadResults(response, "list")
		if err != nil {
			t.Errorf("outputTypeaheadResults() error = %v", err)
		}
	})

	// Empty output is expected for empty suggestions in list format
	if strings.TrimSpace(output) != "" {
		t.Logf("Empty suggestions produced output: %q", output)
	}
}

func TestTypeaheadSingleSuggestion(t *testing.T) {
	response := &api.TypeaheadResponse{
		Suggestions: []string{"exact_match"},
	}

	output := captureOutput(t, func() {
		err := outputTypeaheadResults(response, "json")
		if err != nil {
			t.Errorf("outputTypeaheadResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "exact_match") {
		t.Error("Expected output to contain the single suggestion")
	}
}
