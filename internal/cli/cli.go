package cli

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/cache"
	"github.com/grokipedia/cli/internal/config"
	"github.com/grokipedia/cli/internal/formatter"
	"github.com/mattn/go-isatty"
	"github.com/rodaine/table"
)

// CLI is the main command-line interface structure using Kong
type CLI struct {
	Globals

	Search       SearchCmd      `cmd:"" help:"Search for pages in Grokipedia"`
	Page         PageCmd        `cmd:"" help:"Retrieve a page by slug"`
	Links        LinksCmd       `cmd:"" help:"List links from a page"`
	Edits        EditsCmd       `cmd:"" help:"List edit requests"`
	Typeahead    TypeaheadCmd   `cmd:"" help:"Typeahead search for page titles"`
	Constants    ConstantsCmd   `cmd:"" help:"List API constants and enums"`
	Suggest      SuggestCmd     `cmd:"" help:"Suggest a new article"`
	Edit         EditCmd        `cmd:"" help:"Suggest an edit to an existing article"`
	List         ListCmd        `cmd:"" help:"List pages (browse / discover)"`
	Stats        StatsCmd       `cmd:"" help:"Show global Grokipedia statistics"`
	Preview      PreviewCmd     `cmd:"" help:"Lightweight page preview"`
	TTS          TTSCmd         `cmd:"" help:"List text-to-speech sections for an article"`
	Random       RandomCmd      `cmd:"" help:"Show a random page"`
	Doctor       DoctorCmd      `cmd:"" help:"Run diagnostics (API, cache, config). Try 'grokipedia doctor -v' for details"`
	Completion   CompletionCmd  `cmd:"" help:"Generate shell completion script"`
	CheckUpdates UpdateCheckCmd `cmd:"" name:"check-updates" help:"Check for available updates"`
}

// Globals contains global flags available to all commands
type Globals struct {
	ConfigFile string `help:"Config file (default is ~/.grokipedia/config.yml)" short:"c" env:"GROKIPEDIA_CONFIG"`
	APIURL     string `help:"API base URL" env:"GROKIPEDIA_API_URL"`
	Timeout    int    `help:"Request timeout in seconds" env:"GROKIPEDIA_TIMEOUT"`
	NoCache    bool   `help:"Disable caching" env:"GROKIPEDIA_NO_CACHE"`
	CacheDir   string `help:"Cache directory" env:"GROKIPEDIA_CACHE_DIR"`
	CacheTTL   int    `help:"Cache TTL in seconds" env:"GROKIPEDIA_CACHE_TTL"`
	Verbose    bool   `help:"Enable verbose output" short:"v" env:"GROKIPEDIA_VERBOSE"`
	Debug      bool   `help:"Enable debug output" env:"GROKIPEDIA_DEBUG"`
	Color      string `help:"Color mode: auto, always, never" default:"auto" env:"GROKIPEDIA_COLOR"`

	// Runtime dependencies (initialized by AfterApply)
	appConfig *config.Config
	appCache  *cache.Cache
	appClient *api.Client
}

func (g *Globals) AfterApply() error {
	// Load configuration
	flags := config.GlobalFlags{
		APIURL:     g.APIURL,
		Timeout:    g.Timeout,
		NoCache:    g.NoCache,
		CacheDir:   g.CacheDir,
		CacheTTL:   g.CacheTTL,
		Verbose:    g.Verbose,
		Debug:      g.Debug,
		ConfigFile: g.ConfigFile,
		Color:      g.Color,
	}

	cfg, err := config.Load(flags)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	g.appConfig = cfg

	// Initialize cache if enabled
	if !g.NoCache && cfg.IsCacheEnabled() {
		g.appCache = cache.New(
			cfg.GetCacheDir(),
			int(cfg.GetCacheTTL()),
		)
	}

	// Initialize API client
	g.appClient = api.NewClient(api.ClientOptions{
		BaseURL: cfg.API.URL,
		Timeout: cfg.API.Timeout,
		Verbose: g.Verbose,
		Debug:   g.Debug,
	})

	return nil
}

func (g *Globals) getCache() *cache.Cache {
	return g.appCache
}

func (g *Globals) getClient() *api.Client {
	return g.appClient
}

func (g *Globals) shouldUseColor() bool {
	switch g.Color {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		return isatty.IsTerminal(os.Stdout.Fd())
	default:
		return isatty.IsTerminal(os.Stdout.Fd())
	}
}

// newTestGlobals is a helper for testing command Run methods.
// It allows tests to inject a custom API client.
func newTestGlobals(client *api.Client) *Globals {
	return &Globals{
		appClient: client,
	}
}

// SearchCmd handles the search command
type SearchCmd struct {
	Query  string `arg:"" help:"Search query"`
	Limit  int    `help:"Maximum number of results (1-100)" default:"12"`
	Offset int    `help:"Offset for pagination" default:"0"`
	Format string `help:"Output format: table, json, markdown" default:"table"`
}

func (c *SearchCmd) Run(globals *Globals) error {
	// Validate format
	allowedFormats := []string{"table", "json", "markdown"}
	if err := formatter.ValidateFormat(c.Format, allowedFormats); err != nil {
		return &api.InvalidArgsError{Message: err.Error()}
	}

	// Check cache first
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/full-text-search", map[string]interface{}{
			"q":      c.Query,
			"limit":  c.Limit,
			"offset": c.Offset,
		})
		if data, found := cache.Get(cacheKey); found {
			var cached api.SearchResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputSearchResults(&cached, c.Format, globals.shouldUseColor())
			}
		}
	}

	// Make API request
	client := globals.getClient()
	results, err := client.Search(c.Query, c.Limit, c.Offset)
	if err != nil {
		return err
	}

	// Cache the response
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(results); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputSearchResults(results, c.Format, globals.shouldUseColor())
}

// PageCmd handles the page command
type PageCmd struct {
	Slug    string `arg:"" help:"Page slug"`
	Content bool   `help:"Show page content"`
	NoLinks bool   `help:"Skip link validation"`
	Format  string `help:"Output format: markdown, plain, json" default:"markdown"`
}

func (c *PageCmd) Run(globals *Globals) error {
	// Validate format
	allowedFormats := []string{"markdown", "plain", "json"}
	if err := formatter.ValidateFormat(c.Format, allowedFormats); err != nil {
		return &api.InvalidArgsError{Message: err.Error()}
	}

	// Check cache first
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/page", map[string]interface{}{
			"slug":           c.Slug,
			"includeContent": c.Content,
			"validateLinks":  !c.NoLinks,
		})
		if data, found := cache.Get(cacheKey); found {
			var cached api.PageResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputPageResults(&cached, c.Format, c.Content)
			}
		}
	}

	// Make API request
	client := globals.getClient()
	result, err := client.Page(c.Slug, c.Content, !c.NoLinks)
	if err != nil {
		return err
	}

	// Check if page was found
	if !result.Found {
		return &api.NotFoundError{Resource: c.Slug}
	}

	// Cache the response
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(result); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputPageResults(result, c.Format, c.Content)
}

// EditsCmd handles the edits command
type EditsCmd struct {
	Limit       int      `help:"Maximum number of results (1-100)" default:"20"`
	Status      string   `help:"Filter by status (comma-separated: approved,implemented,pending)"`
	ExcludeUser []string `help:"Exclude edits by username (repeatable)"`
	Counts      bool     `help:"Include count metadata" default:"true"`
	Format      string   `help:"Output format: table, json" default:"table"`
}

func (c *EditsCmd) Run(globals *Globals) error {
	// Validate format
	allowedFormats := []string{"table", "json"}
	if err := formatter.ValidateFormat(c.Format, allowedFormats); err != nil {
		return &api.InvalidArgsError{Message: err.Error()}
	}

	// Parse status filter
	var statusList []string
	if c.Status != "" {
		statusList = strings.Split(c.Status, ",")
		for i := range statusList {
			statusList[i] = strings.TrimSpace(statusList[i])
		}
	}

	// Build cache key params
	cacheParams := map[string]interface{}{
		"limit":         c.Limit,
		"includeCounts": c.Counts,
	}
	if len(statusList) > 0 {
		cacheParams["status"] = c.Status
	}
	if len(c.ExcludeUser) > 0 {
		cacheParams["excludeUsers"] = strings.Join(c.ExcludeUser, ",")
	}

	// Check cache first
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/list-edit-requests", cacheParams)
		if data, found := cache.Get(cacheKey); found {
			var cached api.EditsResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputEditsResults(&cached, c.Format, c.Counts, globals.shouldUseColor())
			}
		}
	}

	// Make API request
	client := globals.getClient()
	results, err := client.Edits(c.Limit, statusList, c.ExcludeUser, c.Counts)
	if err != nil {
		return err
	}

	// Cache the response
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(results); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputEditsResults(results, c.Format, c.Counts, globals.shouldUseColor())
}

// TypeaheadCmd handles the typeahead command
type TypeaheadCmd struct {
	Query string `arg:"" help:"Search query prefix"`
	Limit int    `help:"Maximum number of results" default:"10"`
}

func (c *TypeaheadCmd) Run(globals *Globals) error {
	client := globals.getClient()
	results, err := client.Typeahead(c.Query, c.Limit)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

// ConstantsCmd handles the constants command
type ConstantsCmd struct {
	Format string `help:"Output format: table, json" default:"table"`
}

func (c *ConstantsCmd) Run(globals *Globals) error {
	client := globals.getClient()
	constants, err := client.Constants()
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(constants)
}

// SuggestCmd handles the suggest command
type SuggestCmd struct {
	Title   string `arg:"" help:"Title of the article to suggest"`
	Content string `help:"Optional content or details for the article suggestion"`
	Sources string `help:"Optional sources or references (URLs or citations)"`
	Format  string `help:"Output format: text, json" default:"text"`
}

func (c *SuggestCmd) Run(globals *Globals) error {
	// Validate format
	allowedFormats := []string{"text", "json"}
	if err := formatter.ValidateFormat(c.Format, allowedFormats); err != nil {
		return &api.InvalidArgsError{Message: err.Error()}
	}

	// Validate title is not empty
	if c.Title == "" {
		return &api.InvalidArgsError{Message: "title cannot be empty"}
	}

	// Make API request
	client := globals.getClient()
	req := &api.SuggestArticleRequest{
		Title:   c.Title,
		Content: c.Content,
		Sources: c.Sources,
	}

	resp, err := client.SuggestArticle(req)
	if err != nil {
		return err
	}

	return outputSuggestResults(resp, c.Format, c.Title)
}

func outputSuggestResults(resp *api.SuggestArticleResponse, format string, title string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)

	case "text":
		if resp.Success {
			fmt.Printf("✓ Successfully submitted suggestion for '%s'\n", title)
			if resp.ID != "" {
				fmt.Printf("  Request ID: %s\n", resp.ID)
			}
			if resp.Status != "" {
				fmt.Printf("  Status: %s\n", resp.Status)
			}
			fmt.Println()
			fmt.Println("Your suggestion has been submitted and will be reviewed.")
			fmt.Println("Note: Grokipedia articles are generated by AI based on suggestions.")
		} else {
			fmt.Printf("✗ Failed to submit suggestion for '%s'\n", title)
			if resp.Message != "" {
				fmt.Printf("  Reason: %s\n", resp.Message)
			}
			return &api.InvalidArgsError{Message: resp.Message}
		}
		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}

// =============================================================================
// NEW COMMANDS - Grokipedia feature coverage expansion
// =============================================================================

// StatsCmd shows global statistics
type StatsCmd struct {
	Format string `help:"Output format: table, json" default:"table"`
}

func (c *StatsCmd) Run(globals *Globals) error {
	// Check cache first (stats change infrequently)
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/stats", nil)
		if data, found := cache.Get(cacheKey); found {
			var cached api.StatsResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputStatsResults(&cached, c.Format)
			}
		}
	}

	client := globals.getClient()
	stats, err := client.Stats()
	if err != nil {
		return err
	}

	// Cache the result
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(stats); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputStatsResults(stats, c.Format)
}

// outputStatsResults handles formatting for the stats command
func outputStatsResults(stats *api.StatsResponse, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(stats)
	case "table":
		fmt.Println("Grokipedia Global Statistics")
		fmt.Println("----------------------------")
		fmt.Printf("Total Pages:      %s\n", stats.TotalPages)
		fmt.Printf("Total Views:      %s\n", stats.TotalViews)
		fmt.Printf("Avg Views/Page:   %d\n", stats.AvgViewsPerPage)
		fmt.Printf("Index Size:       %s bytes\n", stats.IndexSizeBytes)
		fmt.Printf("Stats Timestamp:  %s\n", stats.StatsTimestamp)
		return nil
	default:
		return &api.InvalidArgsError{Message: "format must be table or json"}
	}
}

// outputListResults handles formatting for the list command
func outputListResults(resp *api.ListPagesResponse, format string, showCounts bool, useColor bool) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	case "table":
		if len(resp.Pages) == 0 {
			fmt.Println("No pages found.")
			return nil
		}
		tbl := table.New("Title", "Slug", "Views", "Quality", "Categories").WithWriter(os.Stdout)
		if useColor {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}
		for _, p := range resp.Pages {
			cats := strings.Join(p.Metadata.Categories[:min(2, len(p.Metadata.Categories))], ", ")
			if len(p.Metadata.Categories) > 2 {
				cats += "..."
			}
			tbl.AddRow(p.Title, p.Slug, p.Stats.TotalViews, fmt.Sprintf("%.2f", p.Stats.QualityScore), cats)
		}
		tbl.Print()
		if showCounts {
			fmt.Printf("\nShowing %d of %d total (has more: %v)\n", len(resp.Pages), resp.TotalCount, resp.HasMore)
		}
		return nil
	default:
		return &api.InvalidArgsError{Message: "format must be table or json"}
	}
}

// RandomCmd shows a random page
type RandomCmd struct {
	Content  bool   `help:"Include full page content (uses /api/page)"`
	Category string `help:"Limit random selection to a category (e.g. 'Physics', 'History')"`
	Format   string `help:"Output format: markdown, json, plain" default:"markdown"`
}

func (c *RandomCmd) Run(globals *Globals) error {
	client := globals.getClient()

	// Use a safe upper bound because /api/list-pages appears to only expose
	// a subset of pages (often ~10k-50k) rather than all 6M+ pages.
	const maxListable = 50000

	// Pick a random offset in the listable range
	randOffsetBig, _ := rand.Int(rand.Reader, big.NewInt(maxListable))
	randOffset := int(randOffsetBig.Int64())

	// Try up to 3 times in case of empty result at that offset
	var randomPage api.ListPageItem
	found := false
	for attempt := 0; attempt < 3; attempt++ {
		listResp, err := client.ListPages(1, randOffset, c.Category)
		if err == nil && len(listResp.Pages) > 0 {
			randomPage = listResp.Pages[0]
			found = true
			break
		}
		// Try a lower offset on failure
		randOffset = randOffset / 2
	}
	if !found {
		return &api.NotFoundError{Resource: "random page (no listable pages found)"}
	}

	// If user wants full content, fetch the real page
	if c.Content {
		fullPage, err := client.Page(randomPage.Slug, true, false)
		if err == nil && fullPage.Found {
			// Reuse the existing page output logic by constructing a minimal response
			return outputRandomFullPage(fullPage, c.Format)
		}
	}

	// Otherwise show a nice preview of the random page
	switch c.Format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"random": true,
			"page":   randomPage,
		})
	case "markdown":
		fmt.Printf("# 🎲 Random Page\n\n")
		fmt.Printf("**%s**\n\n", randomPage.Title)
		fmt.Printf("**Slug:** `%s`\n", randomPage.Slug)
		if randomPage.Description != "" {
			fmt.Printf("\n%s\n", randomPage.Description)
		}
		fmt.Printf("\n**Views:** %s  •  **Quality:** %.2f\n", randomPage.Stats.TotalViews, randomPage.Stats.QualityScore)
		if len(randomPage.Metadata.Categories) > 0 {
			fmt.Printf("**Categories:** %s\n", strings.Join(randomPage.Metadata.Categories, ", "))
		}
		fmt.Printf("\n*Run `grokipedia page %s` to read the full article.*\n", randomPage.Slug)
		return nil
	case "plain":
		fmt.Printf("Random: %s\nSlug: %s\nViews: %s\nQuality: %.2f\n", randomPage.Title, randomPage.Slug, randomPage.Stats.TotalViews, randomPage.Stats.QualityScore)
		return nil
	default:
		return &api.InvalidArgsError{Message: "invalid format"}
	}
}

// outputRandomFullPage handles the --content case for random
func outputRandomFullPage(pageResp *api.PageResponse, format string) error {
	// Minimal reuse of existing page output style
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{"random": true, "page": pageResp.Page})
	case "markdown":
		fmt.Printf("# 🎲 Random Page: %s\n\n", pageResp.Page.Title)
		if pageResp.Page.Description != "" {
			fmt.Printf("%s\n\n", pageResp.Page.Description)
		}
		if pageResp.Page.Content != "" {
			fmt.Println(pageResp.Page.Content)
		}
		fmt.Printf("\n**Slug:** %s | **Views:** %d | **Quality:** %.2f\n", pageResp.Page.Slug, pageResp.Page.Stats.TotalViews, pageResp.Page.Stats.QualityScore)
		return nil
	case "plain":
		fmt.Printf("Random: %s (Slug: %s)\n", pageResp.Page.Title, pageResp.Page.Slug)
		return nil
	default:
		return &api.InvalidArgsError{Message: "invalid format"}
	}
}

// ListCmd lists pages with pagination (discovery / browsing)
type ListCmd struct {
	Limit    int    `help:"Maximum results (1-100)" default:"20"`
	Offset   int    `help:"Pagination offset" default:"0"`
	Category string `help:"Filter by category"`
	Format   string `help:"Output format: table, json" default:"table"`
	Count    bool   `help:"Show total count" default:"true"`
}

func (c *ListCmd) Run(globals *Globals) error {
	// Check cache
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/list-pages", map[string]interface{}{
			"limit":    c.Limit,
			"offset":   c.Offset,
			"category": c.Category,
		})
		if data, found := cache.Get(cacheKey); found {
			var cached api.ListPagesResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputListResults(&cached, c.Format, c.Count, globals.shouldUseColor())
			}
		}
	}

	client := globals.getClient()
	resp, err := client.ListPages(c.Limit, c.Offset, c.Category)
	if err != nil {
		return err
	}

	// Cache result
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(resp); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputListResults(resp, c.Format, c.Count, globals.shouldUseColor())
}

// PreviewCmd shows lightweight page information
type PreviewCmd struct {
	Slug   string `arg:"" help:"Page slug"`
	Format string `help:"Output format: markdown, json, plain" default:"markdown"`
}

func (c *PreviewCmd) Run(globals *Globals) error {
	// Check cache first
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/page-preview", map[string]interface{}{
			"slug": c.Slug,
		})
		if data, found := cache.Get(cacheKey); found {
			var cached api.PagePreviewResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				if !cached.Found {
					return &api.NotFoundError{Resource: c.Slug}
				}
				return outputPreviewResults(&cached, c.Format)
			}
		}
	}

	client := globals.getClient()
	resp, err := client.PagePreview(c.Slug)
	if err != nil {
		return err
	}
	if !resp.Found {
		return &api.NotFoundError{Resource: c.Slug}
	}

	// Cache the result
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(resp); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputPreviewResults(resp, c.Format)
}

// outputPreviewResults handles formatting for the preview command
func outputPreviewResults(resp *api.PagePreviewResponse, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	case "markdown":
		fmt.Printf("# %s\n\n", resp.Page.Title)
		if resp.Page.Description != "" {
			fmt.Printf("%s\n\n", resp.Page.Description)
		}
		if resp.Page.Content != "" {
			content := resp.Page.Content
			if len(content) > 800 {
				content = content[:800] + "...\n\n*(content truncated — use 'page' command for full text)*"
			}
			fmt.Println(content)
		}
		fmt.Printf("\n**Slug:** %s | **Views:** %s | **Quality:** %.2f\n", resp.Page.Slug, resp.Page.Stats.TotalViews, resp.Page.Stats.QualityScore)
		return nil
	case "plain":
		fmt.Printf("Title: %s\nSlug: %s\nViews: %s\nQuality: %.2f\n", resp.Page.Title, resp.Page.Slug, resp.Page.Stats.TotalViews, resp.Page.Stats.QualityScore)
		return nil
	default:
		return &api.InvalidArgsError{Message: "invalid format"}
	}
}

// TTSCmd lists text-to-speech sections for an article
type TTSCmd struct {
	Slug   string `arg:"" help:"Page slug"`
	Format string `help:"Output format: table, json, list" default:"table"`
}

func (c *TTSCmd) Run(globals *Globals) error {
	// Check cache
	cacheKey := ""
	if cache := globals.getCache(); cache != nil {
		cacheKey = cache.GenerateKey("/api/tts", map[string]interface{}{"slug": c.Slug})
		if data, found := cache.Get(cacheKey); found {
			var cached api.TTSResponse
			if err := json.Unmarshal(data, &cached); err == nil {
				return outputTTSResults(&cached, c.Format)
			}
		}
	}

	client := globals.getClient()
	resp, err := client.TTS(c.Slug)
	if err != nil {
		return err
	}

	// Cache result
	if cache := globals.getCache(); cache != nil && cacheKey != "" {
		if data, err := json.Marshal(resp); err == nil {
			_ = cache.Set(cacheKey, data)
		}
	}

	return outputTTSResults(resp, c.Format)
}

// outputTTSResults handles formatting for the tts command
func outputTTSResults(resp *api.TTSResponse, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	case "table", "list":
		if len(resp.Sections) == 0 {
			fmt.Println("No TTS sections available for this article.")
			return nil
		}
		fmt.Printf("TTS sections for '%s' (%d sections)\n\n", resp.Slug, len(resp.Sections))
		tbl := table.New("#", "Section ID", "Title", "Parts").WithWriter(os.Stdout)
		for i, s := range resp.Sections {
			tbl.AddRow(strconv.Itoa(i+1), s.ID, s.Title, strconv.Itoa(s.PartCount))
		}
		tbl.Print()
		fmt.Println("\nNote: Actual audio generation may require the web player or future CLI support.")
		return nil
	default:
		return &api.InvalidArgsError{Message: "format must be table, list or json"}
	}
}

// DoctorCmd runs basic diagnostics on the CLI environment.
// Detail level is controlled by the global --verbose flag.
type DoctorCmd struct{}

func (c *DoctorCmd) Run(globals *Globals) error {
	useColor := globals.shouldUseColor()
	verbose := globals.Verbose

	printHeader := func(s string) {
		if useColor {
			fmt.Printf("\033[1m%s\033[0m\n", s)
		} else {
			fmt.Println(s)
		}
	}

	fmt.Println("Grokipedia CLI Doctor")
	fmt.Println("=====================")

	// Version
	fmt.Printf("Version:     %s\n", "dev")

	// API Connectivity
	client := globals.getClient()
	_, err := client.Stats()
	if err != nil {
		fmt.Print("API:         ")
		if useColor {
			fmt.Print("\033[31m❌ Unreachable\033[0m")
		} else {
			fmt.Print("❌ Unreachable")
		}
		fmt.Println()
		if verbose {
			fmt.Printf("             Error: %v\n", err)
		}
	} else {
		fmt.Print("API:         ")
		if useColor {
			fmt.Print("\033[32m✅ Reachable\033[0m")
		} else {
			fmt.Print("✅ Reachable")
		}
		fmt.Println()
	}

	// Cache
	cache := globals.getCache()
	if cache != nil {
		fmt.Print("Cache:       ")
		if useColor {
			fmt.Print("\033[32m✅ Enabled\033[0m")
		} else {
			fmt.Print("✅ Enabled")
		}
		fmt.Println(" (~/.grokipedia/cache)")
	} else {
		fmt.Println("Cache:       ❌ Disabled (--no-cache active)")
	}

	// Config
	fmt.Println("Config:      Loaded from environment + config file + defaults")

	// Verbose details
	if verbose {
		fmt.Println()
		printHeader("Detailed Information:")
		apiURL := globals.APIURL
		if apiURL == "" {
			apiURL = "https://grokipedia.com (default)"
		}
		fmt.Printf("  API Base URL: %s\n", apiURL)

		if cache != nil {
			fmt.Println("  Cache Dir:    ~/.grokipedia/cache")
		}
	}

	// Final status
	fmt.Println()
	allGood := (err == nil) && (cache != nil)
	if allGood {
		if useColor {
			fmt.Println("\033[32mOverall:     ✅ Looks healthy\033[0m")
		} else {
			fmt.Println("Overall:     ✅ Looks healthy")
		}
	} else {
		if useColor {
			fmt.Println("\033[31mOverall:     ⚠️  Some issues detected\033[0m")
		} else {
			fmt.Println("Overall:     ⚠️  Some issues detected")
		}
	}

	return nil
}

// EditCmd submits an edit suggestion for an existing article
type EditCmd struct {
	Slug         string `arg:"" help:"Slug of the article to edit"`
	Summary      string `help:"Short summary of the proposed change (REQUIRED)"`
	Content      string `help:"Proposed replacement content"`
	Sources      string `help:"Supporting sources / references (optional)"`
	OriginalText string `help:"The original text being corrected (recommended for precision)"`
	Format       string `help:"Output format: text, json" default:"text"`
}

func (c *EditCmd) Run(globals *Globals) error {
	if c.Summary == "" {
		return &api.InvalidArgsError{Message: "--summary is required for edit suggestions"}
	}

	client := globals.getClient()
	req := &api.CreateEditRequest{
		Slug:         c.Slug,
		Summary:      c.Summary,
		Content:      c.Content,
		Sources:      c.Sources,
		OriginalText: c.OriginalText,
	}

	resp, err := client.CreateEditRequest(req)
	if err != nil {
		if strings.Contains(err.Error(), "Authentication required") || strings.Contains(err.Error(), "401") {
			fmt.Println("✗ Authentication required to submit edit suggestions.")
			fmt.Println("  Please sign in at https://grokipedia.com first, then try the web UI or authenticated client.")
			return nil
		}
		return err
	}

	return outputEditResults(resp, c.Format, c.Slug)
}

// outputEditResults handles formatting for the edit command
func outputEditResults(resp *api.CreateEditResponse, format string, slug string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	case "text":
		if resp.Success {
			fmt.Printf("✓ Edit suggestion submitted for '%s'\n", slug)
			if resp.ID != "" {
				fmt.Printf("  Request ID: %s\n", resp.ID)
			}
			fmt.Println()
			fmt.Println("Your edit suggestion has been submitted for review.")
			fmt.Println("Note: Grokipedia requires authentication for edit submissions.")
			fmt.Println("If this failed with 'Authentication required', please sign in via the website first.")
		} else {
			fmt.Printf("✗ Failed to submit edit for '%s'\n", slug)
			if resp.Message != "" {
				fmt.Printf("  Reason: %s\n", resp.Message)
			}
			if strings.Contains(strings.ToLower(resp.Message), "auth") || strings.Contains(resp.Message, "Authentication") {
				fmt.Println("\nThis action requires you to be signed in with an xAI account.")
				fmt.Println("Visit https://grokipedia.com and sign in, then try again (or use the web UI).")
			}
			return &api.InvalidArgsError{Message: resp.Message}
		}
		return nil
	default:
		return &api.InvalidArgsError{Message: "format must be text or json"}
	}
}

// Link represents a single link extracted from a page
type Link struct {
	Type   string `json:"type"`
	Slug   string `json:"slug,omitempty"`
	Title  string `json:"title,omitempty"`
	URL    string `json:"url,omitempty"`
	Source string `json:"source"`
}

// LinksCmd handles the links command
type LinksCmd struct {
	Slug     string `arg:"" help:"Page slug to extract links from"`
	Internal bool   `help:"Show only internal page links"`
	External bool   `help:"Show only external links (citations)"`
	Tree     bool   `help:"Display links in a tree structure"`
	Format   string `help:"Output format: table, json, markdown, tree" default:"table"`
}

func (c *LinksCmd) Run(globals *Globals) error {
	// Validate format
	allowedFormats := []string{"table", "json", "markdown", "tree"}
	if err := formatter.ValidateFormat(c.Format, allowedFormats); err != nil {
		return &api.InvalidArgsError{Message: err.Error()}
	}

	// Make API request
	client := globals.getClient()
	result, err := client.Page(c.Slug, false, true)
	if err != nil {
		return err
	}

	// Check if page was found
	if !result.Found {
		return &api.NotFoundError{Resource: c.Slug}
	}

	// Extract links from the page
	links := extractLinks(result.Page, c.Internal, c.External)

	// Determine output format
	outputFormat := c.Format
	if c.Tree {
		outputFormat = "tree"
	}

	return outputLinksResults(links, outputFormat, c.Slug, result.Page.Title, globals.shouldUseColor())
}

// extractLinks extracts links from page data based on filters
func extractLinks(page api.PageData, internalOnly, externalOnly bool) []Link {
	var links []Link

	// Add internal links (indexed slugs)
	if !externalOnly {
		for _, slug := range page.LinkedPages.IndexedSlugs {
			links = append(links, Link{
				Type:   "internal",
				Slug:   slug,
				Title:  slug,
				Source: "linked",
			})
		}
		// Add unindexed slugs
		for _, slug := range page.LinkedPages.UnindexedSlugs {
			links = append(links, Link{
				Type:   "internal",
				Slug:   slug,
				Title:  slug,
				Source: "unindexed",
			})
		}
	}

	// Add external links (citations)
	if !internalOnly {
		for _, citation := range page.Citations {
			links = append(links, Link{
				Type:   "external",
				Title:  citation.Title,
				URL:    citation.URL,
				Source: "citation",
			})
		}
	}

	return links
}

func outputLinksResults(links []Link, format string, slug string, title string, useColor bool) error {
	switch format {
	case "json":
		result := struct {
			Slug  string `json:"slug"`
			Title string `json:"title"`
			Links []Link `json:"links"`
		}{
			Slug:  slug,
			Title: title,
			Links: links,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	case "markdown":
		fmt.Printf("# Links from: %s\n\n", title)
		if len(links) == 0 {
			fmt.Println("No links found.")
			return nil
		}
		fmt.Printf("**Slug:** %s\n\n", slug)

		// Group by type
		internal := filterLinksByType(links, "internal")
		external := filterLinksByType(links, "external")

		if len(internal) > 0 {
			fmt.Println("## Internal Links")
			for _, link := range internal {
				fmt.Printf("- [%s](%s)\n", link.Title, link.Slug)
			}
			fmt.Println()
		}

		if len(external) > 0 {
			fmt.Println("## External Links")
			for _, link := range external {
				fmt.Printf("- [%s](%s)\n", link.Title, link.URL)
			}
		}
		return nil

	case "tree":
		fmt.Printf("%s\n", title)
		if len(links) == 0 {
			fmt.Println("└── No links found")
			return nil
		}

		internal := filterLinksByType(links, "internal")
		external := filterLinksByType(links, "external")

		if len(internal) > 0 {
			fmt.Println("├── Internal Links")
			for i, link := range internal {
				prefix := "│   ├──"
				if i == len(internal)-1 {
					prefix = "│   └──"
				}
				fmt.Printf("%s %s (%s)\n", prefix, link.Title, link.Source)
			}
		}

		if len(external) > 0 {
			if len(internal) > 0 {
				fmt.Println("│")
			}
			fmt.Println("└── External Links")
			for i, link := range external {
				prefix := "    ├──"
				if i == len(external)-1 {
					prefix = "    └──"
				}
				fmt.Printf("%s %s\n", prefix, link.Title)
			}
		}
		return nil

	case "table":
		if len(links) == 0 {
			fmt.Println("No links found.")
			return nil
		}

		tbl := table.New("Type", "Title", "URL/Slug", "Source").WithWriter(os.Stdout)
		if useColor {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}

		for _, link := range links {
			display := link.Slug
			if link.Type == "external" {
				display = link.URL
			}
			tbl.AddRow(link.Type, link.Title, display, link.Source)
		}

		tbl.Print()
		fmt.Printf("\nTotal: %d links\n", len(links))
		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}

// filterLinksByType filters links by their type
func filterLinksByType(links []Link, linkType string) []Link {
	var filtered []Link
	for _, link := range links {
		if link.Type == linkType {
			filtered = append(filtered, link)
		}
	}
	return filtered
}

// CompletionCmd handles shell completion generation
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell: bash, zsh, fish, powershell"`
}

func (c *CompletionCmd) Run() error {
	fmt.Printf("# %s completion for grokipedia\n", c.Shell)
	return nil
}

// Helper functions for output formatting

func outputSearchResults(results *api.SearchResponse, format string, useColor bool) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)

	case "markdown":
		fmt.Println("# Search Results")
		fmt.Println()
		if len(results.Results) == 0 {
			fmt.Println("No results found.")
			return nil
		}
		for _, r := range results.Results {
			fmt.Printf("- [%s](%s)\n", r.Title, r.Slug)
			fmt.Printf("  Score: %.2f, Views: %d\n", r.RelevanceScore, r.ViewCount)
			if r.Snippet != "" {
				fmt.Printf("  %s\n", r.Snippet)
			}
			fmt.Println()
		}
		return nil

	case "table":
		if len(results.Results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		tbl := table.New("Title", "Slug", "Score", "Views").WithWriter(os.Stdout)
		if useColor {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}

		for _, r := range results.Results {
			score := strconv.FormatFloat(r.RelevanceScore, 'f', 2, 64)
			tbl.AddRow(r.Title, r.Slug, score, strconv.Itoa(r.ViewCount))
		}

		tbl.Print()
		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}

func outputPageResults(result *api.PageResponse, format string, showContent bool) error {
	page := result.Page

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	case "markdown":
		fmt.Printf("# %s\n\n", page.Title)

		if page.Description != "" {
			fmt.Printf("%s\n\n", page.Description)
		}

		if showContent && page.Content != "" {
			fmt.Println(page.Content)
			fmt.Println()
		}

		fmt.Printf("**Slug:** %s\n", page.Slug)
		fmt.Printf("**Views:** %d\n", page.Stats.TotalViews)
		fmt.Printf("**Quality Score:** %.2f\n", page.Stats.QualityScore)

		if len(page.Citations) > 0 {
			fmt.Println("\n## Citations")
			for _, c := range page.Citations {
				fmt.Printf("- [%s](%s)\n", c.Title, c.URL)
			}
		}

		return nil

	case "plain":
		fmt.Printf("Title: %s\n", page.Title)

		if page.Description != "" {
			fmt.Printf("Description: %s\n", page.Description)
		}

		if showContent && page.Content != "" {
			fmt.Println("\nContent:")
			fmt.Println(page.Content)
		}

		fmt.Printf("\nSlug: %s\n", page.Slug)
		fmt.Printf("Views: %d\n", page.Stats.TotalViews)
		fmt.Printf("Quality Score: %.2f\n", page.Stats.QualityScore)

		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}

func outputEditsResults(results *api.EditsResponse, format string, showCounts bool, useColor bool) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)

	case "table":
		if len(results.EditRequests) == 0 {
			fmt.Println("No edit requests found.")
			return nil
		}

		tbl := table.New("ID", "Slug", "Status", "Editor", "Timestamp").WithWriter(os.Stdout)
		if useColor {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}

		for _, edit := range results.EditRequests {
			timestamp := time.Unix(edit.Timestamp, 0).Format("2006-01-02 15:04")
			status := strings.TrimPrefix(edit.Status, "EDIT_REQUEST_STATUS_")
			tbl.AddRow(edit.ID, edit.Slug, status, edit.Editor, timestamp)
		}

		tbl.Print()

		if showCounts {
			fmt.Printf("\nTotal: %d", results.TotalCount)
			if results.HasMore {
				fmt.Print(" (more available)")
			}
			fmt.Println()
		}

		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}

// Run parses CLI args and executes the appropriate command
func Run(args []string) error {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("grokipedia"),
		kong.Description("Unofficial CLI for the Grokipedia API"),
		kong.UsageOnError(),
	)
	if err != nil {
		return err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}

	return ctx.Run(&cli.Globals)
}
