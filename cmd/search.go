package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/formatter"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var (
	searchLimit  int
	searchOffset int
	searchFormat string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for pages in Grokipedia",
	Long:  `Perform a full-text search across all Grokipedia pages.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		// Validate format
		allowedFormats := []string{"table", "json", "markdown"}
		if err := formatter.ValidateFormat(searchFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/full-text-search", map[string]interface{}{
				"q":      query,
				"limit":  searchLimit,
				"offset": searchOffset,
			})
			if data, found := c.Get(cacheKey); found {
				var cached api.SearchResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputSearchResults(&cached, searchFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		results, err := client.Search(query, searchLimit, searchOffset)
		if err != nil {
			return err
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(results); err == nil {
				c.Set(cacheKey, data)
			}
		}

		return outputSearchResults(results, searchFormat)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Get defaults from config
	cfg := getConfig()
	defaultLimit := 12
	defaultOffset := 0
	defaultFormat := "table"
	if cfg != nil {
		defaultLimit = cfg.Commands.Search.Limit
		defaultOffset = cfg.Commands.Search.Offset
		if cfg.Commands.Search.Format != "" {
			defaultFormat = cfg.Commands.Search.Format
		}
	}

	searchCmd.Flags().IntVar(&searchLimit, "limit", defaultLimit, "Maximum number of results (1-100)")
	searchCmd.Flags().IntVar(&searchOffset, "offset", defaultOffset, "Offset for pagination")
	searchCmd.Flags().StringVar(&searchFormat, "format", defaultFormat, "Output format: table, json, markdown")
}

// outputSearchResults outputs search results in the specified format
func outputSearchResults(results *api.SearchResponse, format string) error {
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
		if shouldUseColor() {
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
