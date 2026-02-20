package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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

	if _, ok := err.(*NotFoundError); !ok {
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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

	if _, ok := err.(*RateLimitError); !ok {
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
		w.Write([]byte("invalid json"))
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
