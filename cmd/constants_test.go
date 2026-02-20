package cmd

import (
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestConstantsCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid json", "json", false},
		{"valid yaml", "yaml", false},
		{"valid table", "table", false},
		{"invalid format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := api.ConstantsResponse{
				"test": "value",
			}
			err := outputConstantsResults(response, "", tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputConstantsResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConstantsOutputJSON(t *testing.T) {
	response := api.ConstantsResponse{
		"maxResults":       100,
		"defaultLimit":     10,
		"apiVersion":       "v1",
		"supportedFormats": []string{"json", "yaml"},
	}

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "", "json")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "maxResults") {
		t.Error("Expected output to contain 'maxResults'")
	}

	if !strings.Contains(output, "100") {
		t.Error("Expected output to contain maxResults value")
	}
}

func TestConstantsOutputYAML(t *testing.T) {
	response := api.ConstantsResponse{
		"version": "1.0.0",
		"debug":   true,
	}

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "", "yaml")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "version") {
		t.Error("Expected output to contain 'version'")
	}
}

func TestConstantsOutputTable(t *testing.T) {
	response := api.ConstantsResponse{
		"KEY_ONE": "value_one",
		"KEY_TWO": "value_two",
	}

	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "", "table")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "KEY_ONE") {
		t.Error("Expected table to contain 'KEY_ONE'")
	}

	if !strings.Contains(output, "value_one") {
		t.Error("Expected table to contain 'value_one'")
	}
}

func TestConstantsFilterByKey(t *testing.T) {
	response := api.ConstantsResponse{
		"wanted":   "this value",
		"unwanted": "not this",
	}

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "wanted", "json")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "this value") {
		t.Error("Expected output to contain filtered value")
	}
}

func TestConstantsFilterUnknownKey(t *testing.T) {
	response := api.ConstantsResponse{
		"existing": "value",
	}

	err := outputConstantsResults(response, "nonexistent", "json")
	if err == nil {
		t.Error("Expected error for unknown constant key")
	}

	if _, ok := err.(*api.UnknownConstantError); !ok {
		t.Errorf("Expected UnknownConstantError, got %T", err)
	}
}

func TestConstantsEmptyResponse(t *testing.T) {
	response := api.ConstantsResponse{}

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "", "table")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "No constants") {
		t.Error("Expected 'No constants' message for empty response")
	}
}

func TestConstantsNestedValues(t *testing.T) {
	response := api.ConstantsResponse{
		"simple": "value",
		"nested": map[string]interface{}{
			"key": "nested_value",
		},
	}

	output := captureOutput(t, func() {
		err := outputConstantsResults(response, "", "json")
		if err != nil {
			t.Errorf("outputConstantsResults() error = %v", err)
		}
	})

	// JSON should handle nested values
	if !strings.Contains(output, "nested") {
		t.Error("Expected output to contain nested key")
	}
}
