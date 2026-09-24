package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// SearchResponse represents the response from /api/full-text-search.
type SearchResponse struct {
	Results          []SearchResult `json:"results"`
	TotalCount       int            `json:"totalCount"`
	Facets           []interface{}  `json:"facets"`
	SearchTimeMs     float64        `json:"searchTimeMs"`
	DetectedLanguage string         `json:"detectedLanguage,omitempty"`
}

// SearchResult represents a single search result.
//
// ViewCount is kept as an integer for the CLI's stable output, but the live API
// currently returns it as a string. UnmarshalJSON accepts both representations.
type SearchResult struct {
	Title             string           `json:"title"`
	Slug              string           `json:"slug"`
	Snippet           string           `json:"snippet"`
	RelevanceScore    float64          `json:"relevanceScore"`
	ViewCount         int              `json:"viewCount"`
	TitleHighlights   []string         `json:"titleHighlights,omitempty"`
	SnippetHighlights []string         `json:"snippetHighlights,omitempty"`
	CreationSource    int              `json:"creationSource,omitempty"`
	Visibility        int              `json:"visibility,omitempty"`
	Providers         []int            `json:"providers,omitempty"`
	SnippetVariants   []SnippetVariant `json:"snippetVariants,omitempty"`
	ScrollAnchorText  string           `json:"scrollAnchorText,omitempty"`
}

// SnippetVariant represents an alternate snippet returned by the search index.
type SnippetVariant struct {
	Provider int     `json:"provider"`
	Kind     int     `json:"kind"`
	Text     string  `json:"text"`
	Score    float64 `json:"score"`
}

// UnmarshalJSON accepts both the historical numeric viewCount and the current
// string representation returned by Grokipedia's live API.
func (r *SearchResult) UnmarshalJSON(data []byte) error {
	var wire struct {
		Title             string           `json:"title"`
		Slug              string           `json:"slug"`
		Snippet           string           `json:"snippet"`
		RelevanceScore    float64          `json:"relevanceScore"`
		ViewCount         json.RawMessage  `json:"viewCount"`
		TitleHighlights   []string         `json:"titleHighlights"`
		SnippetHighlights []string         `json:"snippetHighlights"`
		CreationSource    int              `json:"creationSource"`
		Visibility        int              `json:"visibility"`
		Providers         []int            `json:"providers"`
		SnippetVariants   []SnippetVariant `json:"snippetVariants"`
		ScrollAnchorText  string           `json:"scrollAnchorText"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	viewCount, err := parseJSONInt(wire.ViewCount)
	if err != nil {
		return fmt.Errorf("viewCount: %w", err)
	}

	*r = SearchResult{
		Title:             wire.Title,
		Slug:              wire.Slug,
		Snippet:           wire.Snippet,
		RelevanceScore:    wire.RelevanceScore,
		ViewCount:         viewCount,
		TitleHighlights:   wire.TitleHighlights,
		SnippetHighlights: wire.SnippetHighlights,
		CreationSource:    wire.CreationSource,
		Visibility:        wire.Visibility,
		Providers:         wire.Providers,
		SnippetVariants:   wire.SnippetVariants,
		ScrollAnchorText:  wire.ScrollAnchorText,
	}
	return nil
}

// PageResponse represents the response from /api/page-preview.
type PageResponse struct {
	Page  PageData `json:"page"`
	Found bool     `json:"found"`
}

// PageData represents the page content and metadata.
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

// Citation represents a citation in a page.
type Citation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	Favicon     string `json:"favicon,omitempty"`
}

// Image represents an image in a page.
type Image struct {
	ID       string `json:"id,omitempty"`
	Caption  string `json:"caption"`
	URL      string `json:"url"`
	Position string `json:"position,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// PageMetadata represents page metadata.
type PageMetadata struct {
	Categories     []string `json:"categories"`
	LastModified   int64    `json:"lastModified"`
	ContentLength  int      `json:"contentLength,omitempty"`
	Version        string   `json:"version"`
	LastEditor     string   `json:"lastEditor,omitempty"`
	Language       string   `json:"language,omitempty"`
	IsRedirect     bool     `json:"isRedirect"`
	RedirectTarget string   `json:"redirectTarget,omitempty"`
	IsWithheld     bool     `json:"isWithheld"`
	CreationSource int      `json:"creationSource,omitempty"`
	Visibility     int      `json:"visibility,omitempty"`
}

// PageStats represents page statistics.
type PageStats struct {
	TotalViews   int     `json:"totalViews"`
	QualityScore float64 `json:"qualityScore"`
}

// UnmarshalJSON accepts both numeric and string view counts from old and
// current Grokipedia API responses.
func (s *PageStats) UnmarshalJSON(data []byte) error {
	var wire struct {
		TotalViews   json.RawMessage `json:"totalViews"`
		QualityScore float64         `json:"qualityScore"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	totalViews, err := parseJSONInt(wire.TotalViews)
	if err != nil {
		return fmt.Errorf("totalViews: %w", err)
	}
	s.TotalViews = totalViews
	s.QualityScore = wire.QualityScore
	return nil
}

// LinkedPages represents linked page slugs.
type LinkedPages struct {
	IndexedSlugs   []string `json:"indexedSlugs"`
	UnindexedSlugs []string `json:"unindexedSlugs"`
}

// TypeaheadResponse represents the response from /api/typeahead.
//
// Results is the current response shape. Suggestions is retained for
// compatibility with older Grokipedia deployments and cached responses.
type TypeaheadResponse struct {
	Results      []SearchResult `json:"results,omitempty"`
	Suggestions  []string       `json:"suggestions,omitempty"`
	SearchTimeMs float64        `json:"searchTimeMs,omitempty"`
}

// SuggestionTitles returns titles in the response regardless of API shape.
func (r *TypeaheadResponse) SuggestionTitles() []string {
	if r == nil {
		return nil
	}
	if len(r.Results) > 0 {
		titles := make([]string, 0, len(r.Results))
		for _, result := range r.Results {
			titles = append(titles, result.Title)
		}
		return titles
	}
	return r.Suggestions
}

// ConstantsResponse represents a legacy API constants response.
//
// Grokipedia no longer exposes /api/constants on the live site, but the model
// remains so older self-hosted-compatible endpoints can still be used.
type ConstantsResponse map[string]interface{}

// EditsResponse represents the response from /api/list-edit-requests.
type EditsResponse struct {
	EditRequests         []EditRequest `json:"editRequests"`
	TotalCount           int           `json:"totalCount"`
	HasMore              bool          `json:"hasMore"`
	TotalCountUnfiltered int           `json:"totalCountUnfiltered,omitempty"`
}

// EditRequest represents a single edit request.
type EditRequest struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp,omitempty"` // legacy response field
	Editor    string `json:"editor,omitempty"`    // legacy response field

	// Fields used by the current edit-history API.
	UserID             string               `json:"userId,omitempty"`
	Type               string               `json:"type,omitempty"`
	Summary            string               `json:"summary,omitempty"`
	OriginalContent    string               `json:"originalContent,omitempty"`
	ProposedContent    string               `json:"proposedContent,omitempty"`
	SectionTitle       string               `json:"sectionTitle,omitempty"`
	CreatedAt          int64                `json:"createdAt,omitempty"`
	UpdatedAt          int64                `json:"updatedAt,omitempty"`
	ReviewReason       string               `json:"reviewReason,omitempty"`
	UpvoteCount        int                  `json:"upvoteCount,omitempty"`
	DownvoteCount      int                  `json:"downvoteCount,omitempty"`
	SupportingEvidence []SupportingEvidence `json:"supportingEvidence,omitempty"`
}

// SupportingEvidence represents a source attached to an edit request.
type SupportingEvidence struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// EditsBySlugResponse is kept as a distinct named type for source
// compatibility with the original CLI API.
type EditsBySlugResponse EditsResponse

// SuggestArticleRequest represents the request to suggest a new article.
type SuggestArticleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	// Deprecated fields accepted by older Grokipedia deployments.
	Content string `json:"content,omitempty"`
	Sources string `json:"sources,omitempty"`
}

// SuggestArticleResponse represents the response from
// /api/create-article-request.
type SuggestArticleResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListPagesResponse represents the response from /api/list-pages.
type ListPagesResponse struct {
	Pages      []ListPageItem `json:"pages"`
	TotalCount int            `json:"totalCount"`
	HasMore    bool           `json:"hasMore"`
}

// ListPageItem represents a single page in list-pages results.
type ListPageItem struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Metadata    PageMetadata  `json:"metadata"`
	Stats       ListPageStats `json:"stats"`
	Citations   []Citation    `json:"citations"`
	Images      []Image       `json:"images"`
}

// ListPageStats is the stats shape returned by /api/list-pages. Values are
// normally strings, but accepting numbers keeps the client tolerant of older
// deployments and cached responses.
type ListPageStats struct {
	TotalViews    string  `json:"totalViews"`
	RecentViews   string  `json:"recentViews"`
	DailyAvgViews float64 `json:"dailyAvgViews"`
	QualityScore  float64 `json:"qualityScore"`
	LastViewed    string  `json:"lastViewed"`
}

// UnmarshalJSON accepts both string and numeric total/recent view counts.
func (s *ListPageStats) UnmarshalJSON(data []byte) error {
	var wire struct {
		TotalViews    json.RawMessage `json:"totalViews"`
		RecentViews   json.RawMessage `json:"recentViews"`
		DailyAvgViews float64         `json:"dailyAvgViews"`
		QualityScore  float64         `json:"qualityScore"`
		LastViewed    string          `json:"lastViewed"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	totalViews, err := parseJSONInt(wire.TotalViews)
	if err != nil {
		return fmt.Errorf("totalViews: %w", err)
	}
	recentViews, err := parseJSONInt(wire.RecentViews)
	if err != nil {
		return fmt.Errorf("recentViews: %w", err)
	}

	s.TotalViews = strconv.Itoa(totalViews)
	s.RecentViews = strconv.Itoa(recentViews)
	s.DailyAvgViews = wire.DailyAvgViews
	s.QualityScore = wire.QualityScore
	s.LastViewed = wire.LastViewed
	return nil
}

// StatsResponse represents the response from /api/stats.
type StatsResponse struct {
	TotalPages      string `json:"totalPages"`
	TotalViews      string `json:"totalViews"`
	AvgViewsPerPage int    `json:"avgViewsPerPage"`
	IndexSizeBytes  string `json:"indexSizeBytes"`
	StatsTimestamp  string `json:"statsTimestamp"`
}

// PagePreviewResponse represents the response from /api/page-preview.
type PagePreviewResponse struct {
	Found bool            `json:"found"`
	Page  PreviewPageData `json:"page"`
}

// PreviewPageData is the page shape returned by /api/page-preview.
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

// ToPageData converts the preview response to the stable full-page model used
// by the CLI's page and links commands.
func (p PreviewPageData) ToPageData() PageData {
	return PageData{
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		Description: p.Description,
		Citations:   p.Citations,
		Images:      p.Images,
		Metadata:    p.Metadata,
		Stats: PageStats{
			TotalViews:   parseIntString(p.Stats.TotalViews),
			QualityScore: p.Stats.QualityScore,
		},
		LinkedPages: p.LinkedPages,
	}
}

// TTSResponse represents the response from /api/tts.
type TTSResponse struct {
	Slug     string       `json:"slug"`
	Sections []TTSSection `json:"sections"`
}

// TTSSection represents one speakable section of an article.
type TTSSection struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	PartCount int    `json:"partCount"`
}

// CreateEditRequest represents the payload for /api/create-edit-request.
type CreateEditRequest struct {
	Slug               string               `json:"slug"`
	Type               int                  `json:"type,omitempty"`
	Summary            string               `json:"summary"`
	OriginalContent    string               `json:"originalContent,omitempty"`
	ProposedContent    string               `json:"proposedContent,omitempty"`
	SectionTitle       string               `json:"sectionTitle,omitempty"`
	EditStartHeader    string               `json:"editStartHeader,omitempty"`
	EditEndHeader      string               `json:"editEndHeader,omitempty"`
	SupportingEvidence []SupportingEvidence `json:"supportingEvidence,omitempty"`

	// Deprecated fields used by the original CLI payload.
	Content      string `json:"content,omitempty"`
	Sources      string `json:"sources,omitempty"`
	OriginalText string `json:"originalText,omitempty"`
}

// CreateEditResponse represents the response from /api/create-edit-request.
type CreateEditResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func parseJSONInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0, nil
	}

	var number int
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return parseIntStringErr(text)
	}

	return 0, fmt.Errorf("expected an integer or numeric string")
}

func parseIntString(value string) int {
	parsed, err := parseIntStringErr(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseIntStringErr(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
