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
	typeaheadLimit  int
	typeaheadFormat string
)

// typeaheadCmd represents the typeahead command
var typeaheadCmd = &cobra.Command{
	Use:   "typeahead <query>",
	Short: "Get search suggestions",
	Long:  `Retrieve search suggestions as you type.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		// Validate format
		allowedFormats := []string{"list", "json"}
		if err := formatter.ValidateFormat(typeaheadFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/typeahead", map[string]interface{}{
				"q":     query,
				"limit": typeaheadLimit,
			})
			if data, found := c.Get(cacheKey); found {
				var cached api.TypeaheadResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputTypeaheadResults(&cached, typeaheadFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		results, err := client.Typeahead(query, typeaheadLimit)
		if err != nil {
			return err
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(results); err == nil {
				c.Set(cacheKey, data)
			}
		}

		return outputTypeaheadResults(results, typeaheadFormat)
	},
}

func init() {
	rootCmd.AddCommand(typeaheadCmd)

	typeaheadCmd.Flags().IntVar(&typeaheadLimit, "limit", 5, "Maximum number of suggestions (1-50)")
	typeaheadCmd.Flags().StringVar(&typeaheadFormat, "format", "list", "Output format: list, json")
}

// outputTypeaheadResults outputs typeahead results in the specified format
func outputTypeaheadResults(results *api.TypeaheadResponse, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)

	case "list":
		for _, suggestion := range results.Suggestions {
			fmt.Println(suggestion)
		}
		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}
