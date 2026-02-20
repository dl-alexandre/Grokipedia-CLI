package cmd

import (
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestPageCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid markdown", "markdown", false},
		{"valid plain", "plain", false},
		{"valid json", "json", false},
		{"invalid format", "xml", true},
		{"empty format", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &api.PageResponse{
				Page: api.PageData{
					Title: "Test",
					Slug:  "Test",
				},
				Found: true,
			}
			err := outputPageResults(response, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("outputPageResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPageOutputJSON(t *testing.T) {
	response := &api.PageResponse{
		Page: api.PageData{
			Title:       "Python",
			Slug:        "Python",
			Content:     "# Python\n\nPython is a language.",
			Description: "Python programming language",
			Stats: api.PageStats{
				TotalViews:   10000,
				QualityScore: 0.95,
			},
		},
		Found: true,
	}

	output := captureOutput(t, func() {
		err := outputPageResults(response, "json")
		if err != nil {
			t.Errorf("outputPageResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "Python") {
		t.Error("Expected output to contain 'Python'")
	}

	if !strings.Contains(output, "totalViews") {
		t.Error("Expected output to contain stats")
	}
}

func TestPageOutputMarkdown(t *testing.T) {
	response := &api.PageResponse{
		Page: api.PageData{
			Title:       "Go",
			Slug:        "Go",
			Content:     "# Go\n\nGo is fast.",
			Description: "Go programming language",
			Stats: api.PageStats{
				TotalViews:   5000,
				QualityScore: 0.88,
			},
			Citations: []api.Citation{
				{ID: "1", Title: "Go Docs", URL: "https://go.dev"},
			},
		},
		Found: true,
	}

	output := captureOutput(t, func() {
		err := outputPageResults(response, "markdown")
		if err != nil {
			t.Errorf("outputPageResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "# Go") {
		t.Error("Expected markdown header")
	}

	if !strings.Contains(output, "Views:") {
		t.Error("Expected output to contain views")
	}

	if !strings.Contains(output, "Go Docs") {
		t.Error("Expected output to contain citations")
	}
}

func TestPageOutputPlain(t *testing.T) {
	response := &api.PageResponse{
		Page: api.PageData{
			Title:       "JavaScript",
			Slug:        "JavaScript",
			Content:     "JS content here",
			Description: "JS language",
			Stats: api.PageStats{
				TotalViews:   8000,
				QualityScore: 0.90,
			},
		},
		Found: true,
	}

	output := captureOutput(t, func() {
		// Disable content flag for plain output test
		pageContent = false
		err := outputPageResults(response, "plain")
		if err != nil {
			t.Errorf("outputPageResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "Title:") {
		t.Error("Expected plain output to contain 'Title:'")
	}

	if !strings.Contains(output, "JavaScript") {
		t.Error("Expected output to contain 'JavaScript'")
	}
}

func TestPageWithContent(t *testing.T) {
	response := &api.PageResponse{
		Page: api.PageData{
			Title:   "Test",
			Slug:    "Test",
			Content: "# Full Content\n\nThis is the full page content.",
		},
		Found: true,
	}

	output := captureOutput(t, func() {
		pageContent = true
		err := outputPageResults(response, "markdown")
		if err != nil {
			t.Errorf("outputPageResults() error = %v", err)
		}
	})

	if !strings.Contains(output, "Full Content") {
		t.Error("Expected output to contain full content")
	}
}

func TestPageNotFound(t *testing.T) {
	response := &api.PageResponse{
		Page:  api.PageData{},
		Found: false,
	}

	// Page not found should return NotFoundError when handled by command
	// but outputPageResults just outputs whatever it receives
	output := captureOutput(t, func() {
		err := outputPageResults(response, "markdown")
		if err != nil {
			t.Errorf("outputPageResults() error = %v", err)
		}
	})

	// Empty page should still produce some output
	if output == "" {
		t.Error("Expected some output even for empty page")
	}
}
