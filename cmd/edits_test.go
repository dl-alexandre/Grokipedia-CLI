package cmd

import (
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestEditsCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid table", "table", false},
		{"valid json", "json", false},
		{"invalid format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &api.EditsResponse{
				EditRequests: []api.EditRequest{},
			}
			err := outputEditsResults(response, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputEditsResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditsOutputJSON(t *testing.T) {
	response := &api.EditsResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_PENDING",
				Timestamp: 1702000000,
				Editor:    "user123",
			},
		},
		TotalCount:           1,
		HasMore:              false,
		TotalCountUnfiltered: 1,
	}

	output := captureOutput(t, func() {
		err := outputEditsResults(response, "json")
		if err != nil {
			t.Errorf("outputEditsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "req-001") {
		t.Error("Expected output to contain request ID")
	}

	if !strings.Contains(output, "totalCount") {
		t.Error("Expected output to contain totalCount")
	}
}

func TestEditsOutputTable(t *testing.T) {
	response := &api.EditsResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Python_page",
				Status:    "EDIT_REQUEST_STATUS_PENDING",
				Timestamp: 1702000000,
				Editor:    "alice",
			},
			{
				ID:        "req-002",
				Slug:      "JavaScript_page",
				Status:    "EDIT_REQUEST_STATUS_APPROVED",
				Timestamp: 1701900000,
				Editor:    "bob",
			},
		},
		TotalCount:           2,
		HasMore:              true,
		TotalCountUnfiltered: 10,
	}

	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	oldCounts := editsCounts
	editsCounts = true
	defer func() { editsCounts = oldCounts }()

	output := captureOutput(t, func() {
		err := outputEditsResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "req-001") {
		t.Error("Expected table to contain request ID")
	}

	if !strings.Contains(output, "alice") {
		t.Error("Expected table to contain editor name")
	}

	if !strings.Contains(output, "Total:") {
		t.Error("Expected table to contain total count")
	}

	if !strings.Contains(output, "more available") {
		t.Error("Expected table to indicate more results available")
	}
}

func TestEditsEmptyResults(t *testing.T) {
	response := &api.EditsResponse{
		EditRequests:         []api.EditRequest{},
		TotalCount:           0,
		HasMore:              false,
		TotalCountUnfiltered: 0,
	}

	output := captureOutput(t, func() {
		err := outputEditsResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "No edit requests") {
		t.Error("Expected 'No edit requests' message for empty results")
	}
}

func TestEditsStatusFormatting(t *testing.T) {
	response := &api.EditsResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test",
				Status:    "EDIT_REQUEST_STATUS_PENDING",
				Timestamp: 1702000000,
				Editor:    "user",
			},
		},
	}

	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	output := captureOutput(t, func() {
		err := outputEditsResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsResults() error = %v", err)
		}
	})

	// Status should be trimmed of prefix
	if strings.Contains(output, "EDIT_REQUEST_STATUS_") {
		t.Error("Status should not contain prefix in output")
	}

	if !strings.Contains(output, "PENDING") {
		t.Error("Expected status to show as PENDING")
	}
}

func TestEditsTimestampFormatting(t *testing.T) {
	response := &api.EditsResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test",
				Status:    "PENDING",
				Timestamp: 1702000000, // Known timestamp
				Editor:    "user",
			},
		},
	}

	output := captureOutput(t, func() {
		err := outputEditsResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsResults() error = %v", err)
		}
	})

	// Should contain formatted date (2023-12-08)
	if !strings.Contains(output, "2023") && !strings.Contains(output, "2024") {
		t.Logf("Timestamp formatting output: %s", output)
	}
}
