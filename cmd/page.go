package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/formatter"
	"github.com/spf13/cobra"
)

var (
	pageContent bool
	pageNoLinks bool
	pageFormat  string
)

// pageCmd represents the page command
var pageCmd = &cobra.Command{
	Use:   "page <slug>",
	Short: "Retrieve a page by slug",
	Long:  `Fetch a Grokipedia page by its slug identifier.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		// Validate format
		allowedFormats := []string{"markdown", "plain", "json"}
		if err := formatter.ValidateFormat(pageFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/page", map[string]interface{}{
				"slug":           slug,
				"includeContent": pageContent,
				"validateLinks":  !pageNoLinks,
			})
			if data, found := c.Get(cacheKey); found {
				var cached api.PageResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputPageResults(&cached, pageFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		result, err := client.Page(slug, pageContent, !pageNoLinks)
		if err != nil {
			return err
		}

		// Check if page was found
		if !result.Found {
			return &api.NotFoundError{Resource: slug}
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(result); err == nil {
				_ = c.Set(cacheKey, data)
			}
		}

		return outputPageResults(result, pageFormat)
	},
}

func init() {
	rootCmd.AddCommand(pageCmd)

	pageCmd.Flags().BoolVar(&pageContent, "content", false, "Show page content")
	pageCmd.Flags().BoolVar(&pageNoLinks, "no-links", false, "Skip link validation")
	pageCmd.Flags().StringVar(&pageFormat, "format", "markdown", "Output format: markdown, plain, json")
}

// outputPageResults outputs page results in the specified format
func outputPageResults(result *api.PageResponse, format string) error {
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

		if pageContent && page.Content != "" {
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

		if pageContent && page.Content != "" {
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
