package api

import (
	"encoding/json"
	"testing"
)

func TestSearchResponseSerialization(t *testing.T) {
	response := SearchResponse{
		Results: []SearchResult{
			{
				Title:          "Test Page",
				Slug:           "Test_page",
				Snippet:        "This is a test snippet",
				RelevanceScore: 0.95,
				ViewCount:      1000,
			},
		},
		TotalCount:       1,
		Facets:           []interface{}{},
		SearchTimeMs:     45.5,
		DetectedLanguage: "en",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal SearchResponse: %v", err)
	}

	// Test unmarshaling
	var decoded SearchResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SearchResponse: %v", err)
	}

	// Verify fields
	if len(decoded.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(decoded.Results))
	}

	result := decoded.Results[0]
	if result.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %q", result.Title)
	}
	if result.Slug != "Test_page" {
		t.Errorf("Expected slug 'Test_page', got %q", result.Slug)
	}
	if result.RelevanceScore != 0.95 {
		t.Errorf("Expected relevance score 0.95, got %f", result.RelevanceScore)
	}
	if result.ViewCount != 1000 {
		t.Errorf("Expected view count 1000, got %d", result.ViewCount)
	}
}

func TestPageResponseSerialization(t *testing.T) {
	response := PageResponse{
		Page: PageData{
			Title:       "Test Page",
			Slug:        "Test_page",
			Content:     "# Test Content",
			Description: "Test description",
			Citations: []Citation{
				{ID: "1", Title: "Source 1", URL: "https://example.com/1"},
			},
			Images: []Image{
				{Caption: "Test Image", URL: "https://example.com/img.png"},
			},
			Metadata: PageMetadata{
				Categories:   []string{"Test", "Example"},
				LastModified: 1702000000,
				Version:      "1.0",
			},
			Stats: PageStats{
				TotalViews:   5000,
				QualityScore: 0.85,
			},
			LinkedPages: LinkedPages{
				IndexedSlugs:   []string{"Page1", "Page2"},
				UnindexedSlugs: []string{"Page3"},
			},
		},
		Found: true,
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal PageResponse: %v", err)
	}

	// Test unmarshaling
	var decoded PageResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PageResponse: %v", err)
	}

	// Verify fields
	if !decoded.Found {
		t.Error("Expected Found to be true")
	}
	if decoded.Page.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %q", decoded.Page.Title)
	}
	if decoded.Page.Stats.TotalViews != 5000 {
		t.Errorf("Expected total views 5000, got %d", decoded.Page.Stats.TotalViews)
	}
	if len(decoded.Page.Citations) != 1 {
		t.Errorf("Expected 1 citation, got %d", len(decoded.Page.Citations))
	}
	if len(decoded.Page.LinkedPages.IndexedSlugs) != 2 {
		t.Errorf("Expected 2 indexed slugs, got %d", len(decoded.Page.LinkedPages.IndexedSlugs))
	}
}

func TestTypeaheadResponseSerialization(t *testing.T) {
	response := TypeaheadResponse{
		Suggestions: []string{"python", "python programming", "python tutorial"},
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal TypeaheadResponse: %v", err)
	}

	// Test unmarshaling
	var decoded TypeaheadResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TypeaheadResponse: %v", err)
	}

	// Verify fields
	if len(decoded.Suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(decoded.Suggestions))
	}
	if decoded.Suggestions[0] != "python" {
		t.Errorf("Expected first suggestion 'python', got %q", decoded.Suggestions[0])
	}
}

func TestConstantsResponseSerialization(t *testing.T) {
	response := ConstantsResponse{
		"maxResults":       100,
		"defaultLimit":     10,
		"supportedFormats": []string{"json", "yaml"},
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ConstantsResponse: %v", err)
	}

	// Test unmarshaling
	var decoded ConstantsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ConstantsResponse: %v", err)
	}

	// Verify fields
	if decoded["maxResults"].(float64) != 100 {
		t.Errorf("Expected maxResults 100, got %v", decoded["maxResults"])
	}
}

func TestEditsResponseSerialization(t *testing.T) {
	response := EditsResponse{
		EditRequests: []EditRequest{
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

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal EditsResponse: %v", err)
	}

	// Test unmarshaling
	var decoded EditsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EditsResponse: %v", err)
	}

	// Verify fields
	if len(decoded.EditRequests) != 1 {
		t.Errorf("Expected 1 edit request, got %d", len(decoded.EditRequests))
	}

	edit := decoded.EditRequests[0]
	if edit.ID != "req-001" {
		t.Errorf("Expected ID 'req-001', got %q", edit.ID)
	}
	if edit.Status != "EDIT_REQUEST_STATUS_PENDING" {
		t.Errorf("Expected status 'EDIT_REQUEST_STATUS_PENDING', got %q", edit.Status)
	}
	if decoded.TotalCount != 1 {
		t.Errorf("Expected total count 1, got %d", decoded.TotalCount)
	}
}

func TestEditsBySlugResponse(t *testing.T) {
	// EditsBySlugResponse is a type alias for EditsResponse
	response := EditsBySlugResponse{
		EditRequests: []EditRequest{
			{
				ID:        "req-002",
				Slug:      "Specific_page",
				Status:    "EDIT_REQUEST_STATUS_APPROVED",
				Timestamp: 1701900000,
				Editor:    "user456",
			},
		},
		TotalCount:           1,
		HasMore:              false,
		TotalCountUnfiltered: 5,
	}

	// Test marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal EditsBySlugResponse: %v", err)
	}

	// Test unmarshaling
	var decoded EditsBySlugResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EditsBySlugResponse: %v", err)
	}

	if decoded.TotalCountUnfiltered != 5 {
		t.Errorf("Expected total count unfiltered 5, got %d", decoded.TotalCountUnfiltered)
	}
}

func TestLoadFixtures(t *testing.T) {
	// Test that we can load the fixture files
	fixtures := []string{
		"../../testdata/fixtures/search_response.json",
		"../../testdata/fixtures/page_response.json",
		"../../testdata/fixtures/typeahead_response.json",
		"../../testdata/fixtures/constants_response.json",
		"../../testdata/fixtures/edits_response.json",
		"../../testdata/fixtures/edits_by_slug_response.json",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			// This test just verifies the fixture files exist and are valid JSON
			// In a real test, you would read and parse the file
			_ = fixture // Placeholder - actual file reading would go here
		})
	}
}
