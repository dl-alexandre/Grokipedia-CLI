package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	searchEndpoint      = "/api/full-text-search"
	typeaheadEndpoint   = "/api/typeahead"
	pageEndpoint        = "/api/page-preview"
	legacyPageEndpoint  = "/api/page"
	constantsEndpoint   = "/api/constants"
	editsEndpoint       = "/api/list-edit-requests"
	editsBySlugEndpoint = "/api/list-edit-requests-by-slug"
	suggestEndpoint     = "/api/create-article-request"
	createEditEndpoint  = "/api/create-edit-request"
	listPagesEndpoint   = "/api/list-pages"
	statsEndpoint       = "/api/stats"
	ttsEndpoint         = "/api/tts"
)

// Client wraps the HTTP client for Grokipedia API
type Client struct {
	httpClient    *resty.Client
	baseURL       string
	timeout       time.Duration
	verbose       bool
	debug         bool
	maxRetryDelay time.Duration
}

// ClientOptions contains configuration options for the client
type ClientOptions struct {
	BaseURL       string
	Timeout       int // seconds
	Verbose       bool
	Debug         bool
	MaxRetryDelay time.Duration
}

// NewClient creates a new API client
func NewClient(opts ClientOptions) *Client {
	client := resty.New()

	timeout := time.Duration(opts.Timeout) * time.Second
	if opts.Timeout <= 0 {
		timeout = 30 * time.Second
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://grokipedia.com"
	}

	client.SetTimeout(timeout)
	client.SetBaseURL(baseURL)

	if opts.Debug {
		client.SetDebug(true)
	}

	return &Client{
		httpClient:    client,
		baseURL:       baseURL,
		timeout:       timeout,
		verbose:       opts.Verbose,
		debug:         opts.Debug,
		maxRetryDelay: opts.MaxRetryDelay,
	}
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(req *resty.Request, endpoint string) (*resty.Response, error) {
	return c.doRequestWithMethod(req, endpoint, "")
}

// doRequestWithMethod performs an HTTP request with retry logic and explicit method handling
func (c *Client) doRequestWithMethod(req *resty.Request, endpoint string, method string) (*resty.Response, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		var resp *resty.Response
		var err error
		if method != "" {
			resp, err = req.Execute(method, endpoint)
		} else {
			resp, err = req.Execute(req.Method, endpoint)
		}

		if err != nil {
			lastErr = &NetworkError{Message: err.Error()}

			// Check if we should retry
			if attempt < maxRetries-1 && c.shouldRetry(err) {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}

			return nil, lastErr
		}

		// Check status code
		switch resp.StatusCode() {
		case http.StatusOK:
			return resp, nil
		case http.StatusNotFound:
			return nil, &NotFoundError{Resource: endpoint}
		case http.StatusTooManyRequests:
			retryAfter := c.parseRetryAfter(resp)

			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(retryAfter) * time.Second
				if sleepDuration <= 0 {
					sleepDuration = c.calculateBackoff(attempt)
				}
				if c.maxRetryDelay > 0 && sleepDuration > c.maxRetryDelay {
					sleepDuration = c.maxRetryDelay
				}
				time.Sleep(sleepDuration)
				continue
			}

			return nil, &RateLimitError{RetryAfter: retryAfter}
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			if attempt < maxRetries-1 {
				sleepDuration := c.calculateBackoff(attempt)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("server error: %d", resp.StatusCode())
		default:
			if resp.StatusCode() >= 400 {
				return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), string(resp.Body()))
			}
			return resp, nil
		}
	}

	return nil, lastErr
}

// shouldRetry determines if a request should be retried
func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary")
}

// calculateBackoff calculates exponential backoff with jitter
func (c *Client) calculateBackoff(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}

	baseDelay := time.Duration(1<<(attempt-1)) * time.Second
	// Use crypto/rand for secure random jitter
	jitterInt, _ := rand.Int(rand.Reader, big.NewInt(1000))
	jitter := time.Duration(jitterInt.Int64()) * time.Millisecond
	return baseDelay + jitter
}

// parseRetryAfter parses the Retry-After header
func (c *Client) parseRetryAfter(resp *resty.Response) int {
	retryAfter := resp.Header().Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		return 0
	}

	return seconds
}

// Search performs a full-text search.
//
// The current public API uses the `query` parameter. The small fallback to
// `q` keeps the client usable with older compatible deployments and cached
// test servers.
func (c *Client) Search(query string, limit, offset int) (*SearchResponse, error) {
	resp, err := c.doRequest(c.searchRequest(query, limit, offset, "query"), searchEndpoint)
	if err != nil && isMissingQueryError(err) {
		resp, err = c.doRequest(c.searchRequest(query, limit, offset, "q"), searchEndpoint)
	}
	if err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &result, nil
}

func (c *Client) searchRequest(query string, limit, offset int, queryParam string) *resty.Request {
	return c.httpClient.R().
		SetQueryParam(queryParam, query).
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset))
}

func isMissingQueryError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "missing required parameter") &&
		(strings.Contains(message, "query") || strings.Contains(message, "q"))
}

// Page retrieves a page by slug.
//
// Grokipedia's live API now serves full article data from /api/page-preview;
// /api/page is retained as a fallback for older compatible deployments.
func (c *Client) Page(slug string, includeContent, validateLinks bool) (*PageResponse, error) {
	preview, err := c.fetchPagePreview(slug)
	if err == nil {
		result := &PageResponse{
			Found: preview.Found,
			Page:  preview.Page.ToPageData(),
		}
		if !includeContent {
			result.Page.Content = ""
		}
		// validateLinks is kept in the public method for compatibility. The
		// current preview endpoint does not expose link validation as a query
		// parameter, so link extraction is handled by the CLI when needed.
		_ = validateLinks
		return result, nil
	}

	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		return nil, err
	}

	// Fall back to the original endpoint for older deployments.
	req := c.httpClient.R().
		SetQueryParam("slug", slug).
		SetQueryParam("includeContent", strconv.FormatBool(includeContent)).
		SetQueryParam("validateLinks", strconv.FormatBool(validateLinks))
	resp, legacyErr := c.doRequest(req, legacyPageEndpoint)
	if legacyErr != nil {
		return nil, legacyErr
	}

	var result PageResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}
	return &result, nil
}

func (c *Client) fetchPagePreview(slug string) (*PagePreviewResponse, error) {
	req := c.httpClient.R().SetQueryParam("slug", slug)
	resp, err := c.doRequest(req, pageEndpoint)
	if err != nil {
		return nil, err
	}

	var result PagePreviewResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse page-preview response: %w", err)
	}
	return &result, nil
}

// Typeahead retrieves search suggestions.
func (c *Client) Typeahead(query string, limit int) (*TypeaheadResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("query", query).
		SetQueryParam("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(req, typeaheadEndpoint)
	if err != nil {
		return nil, err
	}

	var result TypeaheadResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse typeahead response: %w", err)
	}

	return &result, nil
}

// Constants retrieves API constants
func (c *Client) Constants() (ConstantsResponse, error) {
	req := c.httpClient.R()

	resp, err := c.doRequest(req, constantsEndpoint)
	if err != nil {
		return nil, err
	}

	var result ConstantsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse constants response: %w", err)
	}

	return result, nil
}

// Edits retrieves edit requests
func (c *Client) Edits(limit int, status []string, excludeUsers []string, includeCounts bool) (*EditsResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("includeCounts", strconv.FormatBool(includeCounts))

	for _, s := range status {
		req.SetQueryParam("status[]", "EDIT_REQUEST_STATUS_"+strings.ToUpper(s))
	}

	for _, user := range excludeUsers {
		req.SetQueryParam("excludeUserId[]", user)
	}

	resp, err := c.doRequest(req, editsEndpoint)
	if err != nil {
		return nil, err
	}

	var result EditsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse edits response: %w", err)
	}

	return &result, nil
}

// EditsBySlug retrieves edit requests for a specific slug
func (c *Client) EditsBySlug(slug string, limit, offset int) (*EditsBySlugResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("slug", slug).
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset))

	resp, err := c.doRequest(req, editsBySlugEndpoint)
	if err != nil {
		return nil, err
	}

	var result EditsBySlugResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse edits-by-slug response: %w", err)
	}

	return &result, nil
}

// SuggestArticle submits a request to suggest a new article
func (c *Client) SuggestArticle(req *SuggestArticleRequest) (*SuggestArticleResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggest request: %w", err)
	}

	httpReq := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload)

	resp, err := c.doRequestWithMethod(httpReq, suggestEndpoint, resty.MethodPost)
	if err != nil {
		return nil, err
	}

	var result SuggestArticleResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse suggest article response: %w", err)
	}

	return &result, nil
}

// ListPages retrieves a paginated list of pages (supports optional category filter)
func (c *Client) ListPages(limit, offset int, category string) (*ListPagesResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset))

	if category != "" {
		req.SetQueryParam("category", category)
	}

	resp, err := c.doRequest(req, listPagesEndpoint)
	if err != nil {
		return nil, err
	}

	var result ListPagesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse list-pages response: %w", err)
	}

	return &result, nil
}

// Stats retrieves global Grokipedia statistics
func (c *Client) Stats() (*StatsResponse, error) {
	req := c.httpClient.R()

	resp, err := c.doRequest(req, statsEndpoint)
	if err != nil {
		return nil, err
	}

	var result StatsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse stats response: %w", err)
	}

	return &result, nil
}

// PagePreview retrieves a lightweight page preview by slug.
func (c *Client) PagePreview(slug string) (*PagePreviewResponse, error) {
	return c.fetchPagePreview(slug)
}

// TTS retrieves text-to-speech section information for a page
func (c *Client) TTS(slug string) (*TTSResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("slug", slug)

	resp, err := c.doRequest(req, ttsEndpoint)
	if err != nil {
		return nil, err
	}

	var result TTSResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse tts response: %w", err)
	}

	return &result, nil
}

// CreateEditRequest submits a proposed edit for an existing article
func (c *Client) CreateEditRequest(req *CreateEditRequest) (*CreateEditResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal edit request: %w", err)
	}

	httpReq := c.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload)

	resp, err := c.doRequestWithMethod(httpReq, createEditEndpoint, resty.MethodPost)
	if err != nil {
		return nil, err
	}

	var result CreateEditResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse create-edit response: %w", err)
	}

	return &result, nil
}
