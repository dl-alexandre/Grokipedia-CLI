package api

// SearchResponse represents the response from /api/full-text-search
type SearchResponse struct {
	Results          []SearchResult `json:"results"`
	TotalCount       int            `json:"totalCount"`
	Facets           []interface{}  `json:"facets"`
	SearchTimeMs     float64        `json:"searchTimeMs"`
	DetectedLanguage string         `json:"detectedLanguage"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title          string  `json:"title"`
	Slug           string  `json:"slug"`
	Snippet        string  `json:"snippet"`
	RelevanceScore float64 `json:"relevanceScore"`
	ViewCount      int     `json:"viewCount"`
}

// PageResponse represents the response from /api/page
type PageResponse struct {
	Page  PageData `json:"page"`
	Found bool     `json:"found"`
}

// PageData represents the page content and metadata
type PageData struct {
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Content     string       `json:"content"`
	Description string       `json:"description"`
	Citations   []Citation   `json:"citations"`
	Images      []Image      `json:"images"`
	Metadata    PageMetadata `json:"metadata"`
	Stats       PageStats    `json:"stats"`
	LinkedPages LinkedPages  `json:"linkedPages"`
}

// Citation represents a citation in a page
type Citation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Image represents an image in a page
type Image struct {
	Caption string `json:"caption"`
	URL     string `json:"url"`
}

// PageMetadata represents page metadata
type PageMetadata struct {
	Categories   []string `json:"categories"`
	LastModified int64    `json:"lastModified"`
	Version      string   `json:"version"`
}

// PageStats represents page statistics
type PageStats struct {
	TotalViews   int     `json:"totalViews"`
	QualityScore float64 `json:"qualityScore"`
}

// LinkedPages represents linked page slugs
type LinkedPages struct {
	IndexedSlugs   []string `json:"indexedSlugs"`
	UnindexedSlugs []string `json:"unindexedSlugs"`
}

// TypeaheadResponse represents the response from /api/typeahead
type TypeaheadResponse struct {
	Suggestions []string `json:"suggestions"`
}

// ConstantsResponse represents the response from /api/constants
// The structure is dynamic, so we use a map
type ConstantsResponse map[string]interface{}

// EditsResponse represents the response from /api/list-edit-requests
type EditsResponse struct {
	EditRequests         []EditRequest `json:"editRequests"`
	TotalCount           int           `json:"totalCount"`
	HasMore              bool          `json:"hasMore"`
	TotalCountUnfiltered int           `json:"totalCountUnfiltered"`
}

// EditRequest represents a single edit request
type EditRequest struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Editor    string `json:"editor"`
}

// EditsBySlugResponse represents the response from /api/list-edit-requests-by-slug
// Same structure as EditsResponse
type EditsBySlugResponse EditsResponse
