package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestSearchCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid table", "table", false},
		{"valid json", "json", false},
		{"valid markdown", "markdown", false},
		{"invalid format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := outputSearchResults(&api.SearchResponse{}, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputSearchResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchOutputJSON(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{
			{
				Title:          "Test",
				Slug:           "Test",
				Snippet:        "Test snippet",
				RelevanceScore: 0.95,
				ViewCount:      100,
			},
		},
		TotalCount: 1,
	}

	output := captureOutput(t, func() {
		err := outputSearchResults(response, "json")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	if output == "" {
		t.Error("Expected JSON output, got empty string")
	}

	if !strings.Contains(output, "Test") {
		t.Error("Expected output to contain 'Test'")
	}

	// Verify it's valid JSON structure
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Error("Expected valid JSON output")
	}
}

func TestSearchOutputMarkdown(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{
			{
				Title:          "Python",
				Slug:           "Python",
				RelevanceScore: 0.95,
				ViewCount:      1000,
			},
		},
	}

	output := captureOutput(t, func() {
		err := outputSearchResults(response, "markdown")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "# Search Results") {
		t.Error("Expected markdown header")
	}

	if !strings.Contains(output, "Python") {
		t.Error("Expected output to contain 'Python'")
	}

	if !strings.Contains(output, "Score:") {
		t.Error("Expected output to contain score")
	}
}

func TestSearchOutputTable(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{
			{
				Title:          "Go",
				Slug:           "Go",
				RelevanceScore: 0.88,
				ViewCount:      500,
			},
		},
	}

	// Save and restore color mode
	oldColorMode := colorMode
	colorMode = "never"
	defer func() { colorMode = oldColorMode }()

	output := captureOutput(t, func() {
		err := outputSearchResults(response, "table")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	// Table output should contain the data
	if !strings.Contains(output, "Go") {
		t.Errorf("Expected table to contain 'Go', got:\n%s", output)
	}

	// Should have header row
	if !strings.Contains(output, "Title") {
		t.Error("Expected table to contain 'Title' header")
	}
}

func TestSearchEmptyResults(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{},
	}

	output := captureOutput(t, func() {
		err := outputSearchResults(response, "table")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "No results") {
		t.Errorf("Expected 'No results' message, got:\n%s", output)
	}
}

func TestSearchMultipleResults(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{
			{
				Title:          "Python",
				Slug:           "Python",
				RelevanceScore: 0.95,
				ViewCount:      1000,
			},
			{
				Title:          "JavaScript",
				Slug:           "JavaScript",
				RelevanceScore: 0.88,
				ViewCount:      800,
			},
			{
				Title:          "Go",
				Slug:           "Go",
				RelevanceScore: 0.82,
				ViewCount:      500,
			},
		},
		TotalCount:       3,
		SearchTimeMs:     45.5,
		DetectedLanguage: "en",
	}

	// Test JSON format with multiple results
	output := captureOutput(t, func() {
		err := outputSearchResults(response, "json")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	// Should contain all titles
	for _, title := range []string{"Python", "JavaScript", "Go"} {
		if !strings.Contains(output, title) {
			t.Errorf("Expected output to contain '%s'", title)
		}
	}

	// Should contain metadata
	if !strings.Contains(output, "totalCount") {
		t.Error("Expected output to contain totalCount")
	}
}

func TestSearchResultsWithSnippet(t *testing.T) {
	response := &api.SearchResponse{
		Results: []api.SearchResult{
			{
				Title:          "Test Page",
				Slug:           "Test_page",
				Snippet:        "This is a <em>highlighted</em> snippet",
				RelevanceScore: 0.95,
				ViewCount:      1000,
			},
		},
	}

	output := captureOutput(t, func() {
		err := outputSearchResults(response, "markdown")
		if err != nil {
			t.Errorf("outputSearchResults() error = %v", err)
		}
	})

	// Markdown should include snippet
	if !strings.Contains(output, "highlighted") {
		t.Error("Expected markdown to contain snippet content")
	}
}

// captureOutput captures stdout during test execution
func captureOutput(t *testing.T, f func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	return buf.String()
}
