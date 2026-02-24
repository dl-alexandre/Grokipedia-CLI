package cli

import (
	"encoding/json"
	"fmt"
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

	Search     SearchCmd     `cmd:"" help:"Search for pages in Grokipedia"`
	Page       PageCmd       `cmd:"" help:"Retrieve a page by slug"`
	Edits      EditsCmd      `cmd:"" help:"List edit requests"`
	Typeahead  TypeaheadCmd  `cmd:"" help:"Typeahead search for page titles"`
	Constants  ConstantsCmd  `cmd:"" help:"List API constants and enums"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completion script"`
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
	// Skip initialization for help and completion commands
	if g.ConfigFile == "" && g.APIURL == "" && g.Timeout == 0 {
		// Probably just showing help
		return nil
	}

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
			cfg.GetCacheTTL(),
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
		kong.Description("A CLI for the Grokipedia API"),
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
