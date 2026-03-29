package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/grokipedia/cli/internal/api"
)

func TestExtractLinksAll(t *testing.T) {
	page := api.PageData{
		LinkedPages: api.LinkedPages{
			IndexedSlugs:   []string{"Go", "Rust"},
			UnindexedSlugs: []string{"Zig"},
		},
		Citations: []api.Citation{
			{ID: "1", Title: "Go Docs", URL: "https://golang.org"},
		},
	}

	links := extractLinks(page, false, false)

	if len(links) != 4 {
		t.Errorf("Expected 4 links, got %d", len(links))
	}

	// Check internal links
	internal := filterLinksByType(links, "internal")
	if len(internal) != 3 {
		t.Errorf("Expected 3 internal links, got %d", len(internal))
	}

	// Check external links
	external := filterLinksByType(links, "external")
	if len(external) != 1 {
		t.Errorf("Expected 1 external link, got %d", len(external))
	}
}

func TestExtractLinksInternalOnly(t *testing.T) {
	page := api.PageData{
		LinkedPages: api.LinkedPages{
			IndexedSlugs:   []string{"Python"},
			UnindexedSlugs: []string{"Ruby"},
		},
		Citations: []api.Citation{
			{ID: "1", Title: "Python Docs", URL: "https://python.org"},
		},
	}

	links := extractLinks(page, true, false)

	if len(links) != 2 {
		t.Errorf("Expected 2 internal links only, got %d", len(links))
	}

	for _, link := range links {
		if link.Type != "internal" {
			t.Errorf("Expected only internal links, got %s", link.Type)
		}
	}
}

func TestExtractLinksExternalOnly(t *testing.T) {
	page := api.PageData{
		LinkedPages: api.LinkedPages{
			IndexedSlugs: []string{"Java"},
		},
		Citations: []api.Citation{
			{ID: "1", Title: "Oracle Docs", URL: "https://oracle.com"},
			{ID: "2", Title: "Java Tutorials", URL: "https://java.com"},
		},
	}

	links := extractLinks(page, false, true)

	if len(links) != 2 {
		t.Errorf("Expected 2 external links only, got %d", len(links))
	}

	for _, link := range links {
		if link.Type != "external" {
			t.Errorf("Expected only external links, got %s", link.Type)
		}
	}
}

func TestExtractLinksEmpty(t *testing.T) {
	page := api.PageData{
		LinkedPages: api.LinkedPages{},
		Citations:   []api.Citation{},
	}

	links := extractLinks(page, false, false)

	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}

func TestFilterLinksByType(t *testing.T) {
	links := []Link{
		{Type: "internal", Slug: "Go"},
		{Type: "external", URL: "https://go.dev"},
		{Type: "internal", Slug: "Rust"},
		{Type: "external", URL: "https://rust-lang.org"},
	}

	internal := filterLinksByType(links, "internal")
	if len(internal) != 2 {
		t.Errorf("Expected 2 internal links, got %d", len(internal))
	}

	external := filterLinksByType(links, "external")
	if len(external) != 2 {
		t.Errorf("Expected 2 external links, got %d", len(external))
	}
}

func TestOutputLinksResultsJSON(t *testing.T) {
	links := []Link{
		{Type: "internal", Slug: "Go", Title: "Go", Source: "linked"},
		{Type: "external", URL: "https://golang.org", Title: "Go Website", Source: "citation"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinksResults(links, "json", "go", "Go Programming", false)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputLinksResults() error = %v", err)
	}

	var parsed struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
		Links []Link `json:"links"`
	}

	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	if parsed.Slug != "go" {
		t.Errorf("Expected slug 'go', got '%s'", parsed.Slug)
	}

	if len(parsed.Links) != 2 {
		t.Errorf("Expected 2 links in JSON, got %d", len(parsed.Links))
	}
}

func TestOutputLinksResultsMarkdown(t *testing.T) {
	links := []Link{
		{Type: "internal", Slug: "Rust", Title: "Rust", Source: "linked"},
		{Type: "external", URL: "https://rust-lang.org", Title: "Rust Website", Source: "citation"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinksResults(links, "markdown", "rust", "Rust", false)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputLinksResults() error = %v", err)
	}

	if !strings.Contains(output, "# Links from: Rust") {
		t.Error("Expected markdown header")
	}

	if !strings.Contains(output, "## Internal Links") {
		t.Error("Expected 'Internal Links' section")
	}

	if !strings.Contains(output, "## External Links") {
		t.Error("Expected 'External Links' section")
	}

	if !strings.Contains(output, "[Rust](Rust)") {
		t.Error("Expected internal link markdown")
	}
}

func TestOutputLinksResultsTree(t *testing.T) {
	links := []Link{
		{Type: "internal", Slug: "Python", Title: "Python", Source: "linked"},
		{Type: "internal", Slug: "Ruby", Title: "Ruby", Source: "linked"},
		{Type: "external", URL: "https://python.org", Title: "Python.org", Source: "citation"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinksResults(links, "tree", "python", "Python", false)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputLinksResults() error = %v", err)
	}

	if !strings.Contains(output, "Python") {
		t.Error("Expected title in tree output")
	}

	if !strings.Contains(output, "Internal Links") {
		t.Error("Expected 'Internal Links' in tree")
	}

	if !strings.Contains(output, "External Links") {
		t.Error("Expected 'External Links' in tree")
	}

	// Check tree characters
	if !strings.Contains(output, "├──") || !strings.Contains(output, "└──") {
		t.Error("Expected tree branch characters")
	}
}

func TestOutputLinksResultsTable(t *testing.T) {
	links := []Link{
		{Type: "internal", Slug: "Go", Title: "Go", Source: "linked"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinksResults(links, "table", "go", "Go", false)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputLinksResults() error = %v", err)
	}

	if !strings.Contains(output, "internal") {
		t.Error("Expected 'internal' in table output")
	}

	if !strings.Contains(output, "Go") {
		t.Error("Expected 'Go' in table output")
	}

	if !strings.Contains(output, "Total: 1 links") {
		t.Error("Expected total count in table output")
	}
}

func TestOutputLinksResultsEmpty(t *testing.T) {
	links := []Link{}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinksResults(links, "table", "empty", "Empty Page", false)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if err != nil {
		t.Errorf("outputLinksResults() error = %v", err)
	}

	if !strings.Contains(output, "No links found") {
		t.Error("Expected 'No links found' message")
	}
}

func TestOutputLinksResultsInvalidFormat(t *testing.T) {
	links := []Link{{Type: "internal", Slug: "Test"}}

	err := outputLinksResults(links, "xml", "test", "Test", false)

	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected 'invalid format' error, got: %v", err)
	}
}

func TestLinkStruct(t *testing.T) {
	link := Link{
		Type:   "internal",
		Slug:   "Rust",
		Title:  "Rust Programming",
		URL:    "",
		Source: "linked",
	}

	if link.Type != "internal" {
		t.Error("Link type mismatch")
	}

	if link.Slug != "Rust" {
		t.Error("Link slug mismatch")
	}

	if link.Source != "linked" {
		t.Error("Link source mismatch")
	}
}
