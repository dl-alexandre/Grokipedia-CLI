package cmd

import (
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestEditsBySlugCommandValidation(t *testing.T) {
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
			response := &api.EditsBySlugResponse{
				EditRequests: []api.EditRequest{},
			}
			err := outputEditsBySlugResults(response, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputEditsBySlugResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditsBySlugOutputJSON(t *testing.T) {
	response := &api.EditsBySlugResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Python_programming_language",
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
		err := outputEditsBySlugResults(response, "json")
		if err != nil {
			t.Errorf("outputEditsBySlugResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "Python_programming_language") {
		t.Error("Expected output to contain slug")
	}

	if !strings.Contains(output, "totalCount") {
		t.Error("Expected output to contain totalCount")
	}
}

func TestEditsBySlugOutputTable(t *testing.T) {
	response := &api.EditsBySlugResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_APPROVED",
				Timestamp: 1702000000,
				Editor:    "alice",
			},
			{
				ID:        "req-002",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_IMPLEMENTED",
				Timestamp: 1701900000,
				Editor:    "bob",
			},
		},
		TotalCount:           2,
		HasMore:              false,
		TotalCountUnfiltered: 2,
	}

	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	output := captureOutput(t, func() {
		err := outputEditsBySlugResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsBySlugResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "req-001") {
		t.Error("Expected table to contain first request ID")
	}

	if !strings.Contains(output, "req-002") {
		t.Error("Expected table to contain second request ID")
	}

	if !strings.Contains(output, "Total:") {
		t.Error("Expected table to show total count")
	}
}

func TestEditsBySlugEmptyResults(t *testing.T) {
	response := &api.EditsBySlugResponse{
		EditRequests:         []api.EditRequest{},
		TotalCount:           0,
		HasMore:              false,
		TotalCountUnfiltered: 0,
	}

	output := captureOutput(t, func() {
		err := outputEditsBySlugResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsBySlugResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "No edit requests") {
		t.Error("Expected 'No edit requests' message for empty results")
	}

	if !strings.Contains(output, "for this page") {
		t.Error("Expected message to mention 'this page'")
	}
}

func TestEditsBySlugHasMore(t *testing.T) {
	response := &api.EditsBySlugResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test_page",
				Status:    "PENDING",
				Timestamp: 1702000000,
				Editor:    "user",
			},
		},
		TotalCount:           1,
		HasMore:              true,
		TotalCountUnfiltered: 5,
	}

	output := captureOutput(t, func() {
		err := outputEditsBySlugResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsBySlugResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "more available") {
		t.Error("Expected output to indicate more results available")
	}
}

func TestEditsBySlugMultipleStatuses(t *testing.T) {
	response := &api.EditsBySlugResponse{
		EditRequests: []api.EditRequest{
			{
				ID:        "req-001",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_PENDING",
				Timestamp: 1702000000,
				Editor:    "alice",
			},
			{
				ID:        "req-002",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_APPROVED",
				Timestamp: 1701900000,
				Editor:    "bob",
			},
			{
				ID:        "req-003",
				Slug:      "Test_page",
				Status:    "EDIT_REQUEST_STATUS_IMPLEMENTED",
				Timestamp: 1701800000,
				Editor:    "charlie",
			},
		},
		TotalCount: 3,
	}

	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	output := captureOutput(t, func() {
		err := outputEditsBySlugResults(response, "table")
		if err != nil {
			t.Errorf("outputEditsBySlugResults() error = %v", err)
		}
	})

	statuses := []string{"PENDING", "APPROVED", "IMPLEMENTED"}
	for _, status := range statuses {
		if !strings.Contains(output, status) {
			t.Errorf("Expected output to contain status '%s'", status)
		}
	}
}
