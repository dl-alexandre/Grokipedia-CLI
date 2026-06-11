package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		opts        ClientOptions
		wantURL     string
		wantTimeout time.Duration
	}{
		{
			name:        "default options",
			opts:        ClientOptions{},
			wantURL:     "https://grokipedia.com",
			wantTimeout: 30 * time.Second,
		},
		{
			name: "custom options",
			opts: ClientOptions{
				BaseURL: "https://custom.api.com",
				Timeout: 60,
			},
			wantURL:     "https://custom.api.com",
			wantTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.opts)
			if client == nil {
				t.Fatal("NewClient() returned nil")
			}
			if client.baseURL != tt.wantURL {
				t.Errorf("baseURL = %q, want %q", client.baseURL, tt.wantURL)
			}
			if client.timeout != tt.wantTimeout {
				t.Errorf("timeout = %v, want %v", client.timeout, tt.wantTimeout)
			}
		})
	}
}

func TestClientSearch(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/full-text-search" {
			t.Errorf("Expected path /api/full-text-search, got %s", r.URL.Path)
		}

		// Check query params
		q := r.URL.Query().Get("q")
		if q != "test query" {
			t.Errorf("Expected query 'test query', got %s", q)
		}

		// Return mock response
		response := SearchResponse{
			Results: []SearchResult{
				{
					Title:          "Test Page",
					Slug:           "Test_page",
					Snippet:        "Test snippet",
					RelevanceScore: 0.95,
					ViewCount:      1000,
				},
			},
			TotalCount:   1,
			SearchTimeMs: 45.5,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.Search("test query", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %s", result.Results[0].Title)
	}
}

func TestClientPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/page" {
			t.Errorf("Expected path /api/page, got %s", r.URL.Path)
		}

		slug := r.URL.Query().Get("slug")
		if slug != "Test_page" {
			t.Errorf("Expected slug 'Test_page', got %s", slug)
		}

		response := PageResponse{
			Page: PageData{
				Title: "Test Page",
				Slug:  "Test_page",
			},
			Found: true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.Page("Test_page", true, true)
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}

	if !result.Found {
		t.Error("Expected page to be found")
	}

	if result.Page.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %s", result.Page.Title)
	}
}

func TestClientPageNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	_, err := client.Page("NonExistent", true, true)
	if err == nil {
		t.Fatal("Expected error for 404 response")
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestClientTypeahead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/typeahead" {
			t.Errorf("Expected path /api/typeahead, got %s", r.URL.Path)
		}

		response := TypeaheadResponse{
			Suggestions: []string{"python", "python programming"},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.Typeahead("pyt", 5)
	if err != nil {
		t.Fatalf("Typeahead() error = %v", err)
	}

	if len(result.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(result.Suggestions))
	}
}

func TestClientConstants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/constants" {
			t.Errorf("Expected path /api/constants, got %s", r.URL.Path)
		}

		response := ConstantsResponse{
			"maxResults": 100,
			"version":    "v1",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.Constants()
	if err != nil {
		t.Fatalf("Constants() error = %v", err)
	}

	if result["maxResults"].(float64) != 100 {
		t.Errorf("Expected maxResults 100, got %v", result["maxResults"])
	}
}

func TestClientEdits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/list-edit-requests" {
			t.Errorf("Expected path /api/list-edit-requests, got %s", r.URL.Path)
		}

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
			TotalCount: 1,
			HasMore:    false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.Edits(10, []string{"pending"}, []string{}, true)
	if err != nil {
		t.Fatalf("Edits() error = %v", err)
	}

	if len(result.EditRequests) != 1 {
		t.Errorf("Expected 1 edit request, got %d", len(result.EditRequests))
	}
}

func TestClientEditsBySlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/list-edit-requests-by-slug" {
			t.Errorf("Expected path /api/list-edit-requests-by-slug, got %s", r.URL.Path)
		}

		slug := r.URL.Query().Get("slug")
		if slug != "Test_page" {
			t.Errorf("Expected slug 'Test_page', got %s", slug)
		}

		response := EditsBySlugResponse{
			EditRequests: []EditRequest{
				{
					ID:        "req-002",
					Slug:      "Test_page",
					Status:    "EDIT_REQUEST_STATUS_APPROVED",
					Timestamp: 1701900000,
					Editor:    "user456",
				},
			},
			TotalCount: 1,
			HasMore:    false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	result, err := client.EditsBySlug("Test_page", 10, 0)
	if err != nil {
		t.Fatalf("EditsBySlug() error = %v", err)
	}

	if len(result.EditRequests) != 1 {
		t.Errorf("Expected 1 edit request, got %d", len(result.EditRequests))
	}
}

func TestClientRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL:       server.URL,
		Timeout:       30,
		MaxRetryDelay: 10 * time.Millisecond,
	})

	_, err := client.Search("test", 10, 0)
	if err == nil {
		t.Fatal("Expected error for 429 response")
	}

	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Errorf("Expected RateLimitError, got %T", err)
	}
}

func TestClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	_, err := client.Search("test", 10, 0)
	if err == nil {
		t.Fatal("Expected error for 500 response")
	}
}

func TestClientInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	_, err := client.Search("test", 10, 0)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestShouldRetry(t *testing.T) {
	client := NewClient(ClientOptions{})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      &NetworkError{Message: "connection timeout"},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      &NetworkError{Message: "connection refused"},
			expected: true,
		},
		{
			name:     "no such host",
			err:      &NetworkError{Message: "no such host"},
			expected: true,
		},
		{
			name:     "temporary error",
			err:      &NetworkError{Message: "temporary failure"},
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "other error",
			err:      &NetworkError{Message: "some other error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.shouldRetry(tt.err)
			if got != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClientSuggestArticle(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/create-article-request" {
			t.Errorf("Expected path /api/create-article-request, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
		}

		// Parse request body
		var req SuggestArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Title != "Test Article" {
			t.Errorf("Expected title 'Test Article', got %s", req.Title)
		}

		if req.Content != "Test content" {
			t.Errorf("Expected content 'Test content', got %s", req.Content)
		}

		if req.Sources != "https://example.com" {
			t.Errorf("Expected sources 'https://example.com', got %s", req.Sources)
		}

		// Return mock response
		response := SuggestArticleResponse{
			Success: true,
			ID:      "req-789",
			Status:  "PENDING",
			Message: "Request submitted successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	req := &SuggestArticleRequest{
		Title:   "Test Article",
		Content: "Test content",
		Sources: "https://example.com",
	}

	result, err := client.SuggestArticle(req)
	if err != nil {
		t.Fatalf("SuggestArticle() error = %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.ID != "req-789" {
		t.Errorf("Expected ID 'req-789', got %s", result.ID)
	}

	if result.Status != "PENDING" {
		t.Errorf("Expected status 'PENDING', got %s", result.Status)
	}
}

func TestClientSuggestArticleFailure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SuggestArticleResponse{
			Success: false,
			Message: "Article already exists",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		BaseURL: server.URL,
		Timeout: 30,
	})

	req := &SuggestArticleRequest{
		Title: "Existing Article",
	}

	result, err := client.SuggestArticle(req)
	if err != nil {
		t.Fatalf("SuggestArticle() error = %v", err)
	}

	if result.Success {
		t.Error("Expected success to be false")
	}

	if result.Message != "Article already exists" {
		t.Errorf("Expected message 'Article already exists', got %s", result.Message)
	}
}

// =============================================================================
// Tests for new endpoints (list-pages, stats, page-preview, tts, create-edit-request)
// =============================================================================

func TestClientListPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/list-pages" {
			t.Errorf("Expected path /api/list-pages, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit 10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "5" {
			t.Errorf("Expected offset 5, got %s", r.URL.Query().Get("offset"))
		}
		if r.URL.Query().Get("category") != "Physics" {
			t.Errorf("Expected category Physics, got %s", r.URL.Query().Get("category"))
		}

		response := ListPagesResponse{
			Pages: []ListPageItem{
				{
					Slug:  "quantum_computing",
					Title: "Quantum Computing",
					Metadata: PageMetadata{
						Categories: []string{"Physics", "Computing"},
					},
					Stats: ListPageStats{
						TotalViews:   "12500",
						QualityScore: 0.92,
					},
				},
			},
			TotalCount: 125000,
			HasMore:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	result, err := client.ListPages(10, 5, "Physics")
	if err != nil {
		t.Fatalf("ListPages() error = %v", err)
	}
	if len(result.Pages) != 1 {
		t.Fatalf("Expected 1 page, got %d", len(result.Pages))
	}
	if result.Pages[0].Title != "Quantum Computing" {
		t.Errorf("Expected title 'Quantum Computing', got %s", result.Pages[0].Title)
	}
	if result.TotalCount != 125000 {
		t.Errorf("Expected TotalCount 125000, got %d", result.TotalCount)
	}
	if !result.HasMore {
		t.Error("Expected HasMore to be true")
	}
}

func TestClientStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats" {
			t.Errorf("Expected path /api/stats, got %s", r.URL.Path)
		}

		response := StatsResponse{
			TotalPages:      "6092140",
			TotalViews:      "0",
			AvgViewsPerPage: 0,
			IndexSizeBytes:  "194716652399",
			StatsTimestamp:  "1778900115",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	result, err := client.Stats()
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if result.TotalPages != "6092140" {
		t.Errorf("Expected TotalPages '6092140', got %s", result.TotalPages)
	}
	if result.IndexSizeBytes != "194716652399" {
		t.Errorf("Unexpected IndexSizeBytes")
	}
}

func TestClientPagePreview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/page-preview" {
			t.Errorf("Expected path /api/page-preview, got %s", r.URL.Path)
		}
		slug := r.URL.Query().Get("slug")
		if slug != "Test_Article" {
			t.Errorf("Expected slug Test_Article, got %s", slug)
		}

		response := PagePreviewResponse{
			Found: true,
			Page: PreviewPageData{
				Title: "Test Article",
				Slug:  "Test_Article",
				Stats: ListPageStats{
					TotalViews:   "4500",
					QualityScore: 0.88,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	result, err := client.PagePreview("Test_Article")
	if err != nil {
		t.Fatalf("PagePreview() error = %v", err)
	}
	if !result.Found {
		t.Error("Expected Found to be true")
	}
	if result.Page.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got %s", result.Page.Title)
	}
}

func TestClientTTS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tts" {
			t.Errorf("Expected path /api/tts, got %s", r.URL.Path)
		}

		response := TTSResponse{
			Slug: "Quantum_computing",
			Sections: []TTSSection{
				{ID: "intro", Title: "Quantum computing", PartCount: 1},
				{ID: "historical-development", Title: "Historical Development", PartCount: 1},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	result, err := client.TTS("Quantum_computing")
	if err != nil {
		t.Fatalf("TTS() error = %v", err)
	}
	if result.Slug != "Quantum_computing" {
		t.Errorf("Unexpected slug")
	}
	if len(result.Sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(result.Sections))
	}
}

func TestClientCreateEditRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/create-edit-request" {
			t.Errorf("Expected path /api/create-edit-request, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		response := CreateEditResponse{
			Success: true,
			ID:      "edit-456",
			Status:  "PENDING",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	req := &CreateEditRequest{
		Slug:    "Test_Article",
		Summary: "Fixed a small error",
		Content: "Updated paragraph",
	}

	result, err := client.CreateEditRequest(req)
	if err != nil {
		t.Fatalf("CreateEditRequest() error = %v", err)
	}
	if !result.Success {
		t.Error("Expected success to be true")
	}
	if result.ID != "edit-456" {
		t.Errorf("Expected ID 'edit-456', got %s", result.ID)
	}
}

func TestClientCreateEditRequest_AuthRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":"Authentication required"}`))
	}))
	defer server.Close()

	client := NewClient(ClientOptions{BaseURL: server.URL})

	req := &CreateEditRequest{Slug: "x", Summary: "test"}

	_, err := client.CreateEditRequest(req)
	if err == nil {
		t.Fatal("Expected error for auth required")
	}
	if !strings.Contains(err.Error(), "401") && !strings.Contains(err.Error(), "Authentication") {
		t.Errorf("Expected auth-related error, got: %v", err)
	}
}
