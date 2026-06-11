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

// SuggestArticleRequest represents the request to suggest a new article
type SuggestArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
	Sources string `json:"sources,omitempty"`
}

// SuggestArticleResponse represents the response from /api/create-article-request
type SuggestArticleResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// ListPagesResponse represents the response from /api/list-pages
type ListPagesResponse struct {
	Pages      []ListPageItem `json:"pages"`
	TotalCount int            `json:"totalCount"`
	HasMore    bool           `json:"hasMore"`
}

// ListPageItem represents a single page in list-pages results
type ListPageItem struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Metadata    PageMetadata  `json:"metadata"`
	Stats       ListPageStats `json:"stats"`
	Citations   []Citation    `json:"citations"`
	Images      []Image       `json:"images"`
}

// ListPageStats is the stats shape returned by /api/list-pages (values are strings)
type ListPageStats struct {
	TotalViews    string  `json:"totalViews"`
	RecentViews   string  `json:"recentViews"`
	DailyAvgViews float64 `json:"dailyAvgViews"`
	QualityScore  float64 `json:"qualityScore"`
	LastViewed    string  `json:"lastViewed"`
}

// StatsResponse represents the response from /api/stats
type StatsResponse struct {
	TotalPages      string `json:"totalPages"`
	TotalViews      string `json:"totalViews"`
	AvgViewsPerPage int    `json:"avgViewsPerPage"`
	IndexSizeBytes  string `json:"indexSizeBytes"`
	StatsTimestamp  string `json:"statsTimestamp"` // API returns as string
}

// PagePreviewResponse represents the response from /api/page-preview
type PagePreviewResponse struct {
	Found bool            `json:"found"`
	Page  PreviewPageData `json:"page"`
}

// PreviewPageData is a lighter page shape returned by /api/page-preview (stats values may be strings)
type PreviewPageData struct {
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	Content     string        `json:"content"`
	Description string        `json:"description"`
	Citations   []Citation    `json:"citations"`
	Images      []Image       `json:"images"`
	Metadata    PageMetadata  `json:"metadata"`
	Stats       ListPageStats `json:"stats"`
	LinkedPages LinkedPages   `json:"linkedPages"`
}

// TTSResponse represents the response from /api/tts
type TTSResponse struct {
	Slug     string       `json:"slug"`
	Sections []TTSSection `json:"sections"`
}

// TTSSection represents one speakable section of an article
type TTSSection struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	PartCount int    `json:"partCount"`
}

// CreateEditRequest represents the payload for /api/create-edit-request
type CreateEditRequest struct {
	Slug         string `json:"slug"`
	Summary      string `json:"summary"`
	Content      string `json:"content,omitempty"`
	Sources      string `json:"sources,omitempty"`
	OriginalText string `json:"originalText,omitempty"`
}

// CreateEditResponse represents the response from /api/create-edit-request
type CreateEditResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}
