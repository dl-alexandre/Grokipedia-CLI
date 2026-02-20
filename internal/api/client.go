package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := req.Execute(req.Method, endpoint)

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
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
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

// Search performs a full-text search
func (c *Client) Search(query string, limit, offset int) (*SearchResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("q", query).
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset))

	resp, err := c.doRequest(req, "/api/full-text-search")
	if err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &result, nil
}

// Page retrieves a page by slug
func (c *Client) Page(slug string, includeContent, validateLinks bool) (*PageResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("slug", slug).
		SetQueryParam("includeContent", strconv.FormatBool(includeContent)).
		SetQueryParam("validateLinks", strconv.FormatBool(validateLinks))

	resp, err := c.doRequest(req, "/api/page")
	if err != nil {
		return nil, err
	}

	var result PageResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	return &result, nil
}

// Typeahead retrieves search suggestions
func (c *Client) Typeahead(query string, limit int) (*TypeaheadResponse, error) {
	req := c.httpClient.R().
		SetQueryParam("q", query).
		SetQueryParam("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(req, "/api/typeahead")
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

	resp, err := c.doRequest(req, "/api/constants")
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

	resp, err := c.doRequest(req, "/api/list-edit-requests")
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

	resp, err := c.doRequest(req, "/api/list-edit-requests-by-slug")
	if err != nil {
		return nil, err
	}

	var result EditsBySlugResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse edits-by-slug response: %w", err)
	}

	return &result, nil
}
