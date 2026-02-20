package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/formatter"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var (
	editsBySlugLimit  int
	editsBySlugOffset int
	editsBySlugFormat string
)

// editsBySlugCmd represents the edits-by-slug command
var editsBySlugCmd = &cobra.Command{
	Use:   "edits-by-slug <slug>",
	Short: "List edit requests for a specific page",
	Long:  `Retrieve edit requests for a specific page by its slug.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		// Validate format
		allowedFormats := []string{"table", "json"}
		if err := formatter.ValidateFormat(editsBySlugFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/list-edit-requests-by-slug", map[string]interface{}{
				"slug":   slug,
				"limit":  editsBySlugLimit,
				"offset": editsBySlugOffset,
			})
			if data, found := c.Get(cacheKey); found {
				var cached api.EditsBySlugResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputEditsBySlugResults(&cached, editsBySlugFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		results, err := client.EditsBySlug(slug, editsBySlugLimit, editsBySlugOffset)
		if err != nil {
			return err
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(results); err == nil {
				c.Set(cacheKey, data)
			}
		}

		return outputEditsBySlugResults(results, editsBySlugFormat)
	},
}

func init() {
	rootCmd.AddCommand(editsBySlugCmd)

	editsBySlugCmd.Flags().IntVar(&editsBySlugLimit, "limit", 10, "Maximum number of results (1-100)")
	editsBySlugCmd.Flags().IntVar(&editsBySlugOffset, "offset", 0, "Offset for pagination")
	editsBySlugCmd.Flags().StringVar(&editsBySlugFormat, "format", "table", "Output format: table, json")
}

// outputEditsBySlugResults outputs edits by slug results in the specified format
func outputEditsBySlugResults(results *api.EditsBySlugResponse, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)

	case "table":
		if len(results.EditRequests) == 0 {
			fmt.Println("No edit requests found for this page.")
			return nil
		}

		tbl := table.New("ID", "Status", "Editor", "Timestamp").WithWriter(os.Stdout)
		if shouldUseColor() {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}

		for _, edit := range results.EditRequests {
			timestamp := time.Unix(edit.Timestamp, 0).Format("2006-01-02 15:04")
			status := strings.TrimPrefix(edit.Status, "EDIT_REQUEST_STATUS_")
			tbl.AddRow(edit.ID, status, edit.Editor, timestamp)
		}

		tbl.Print()

		fmt.Printf("\nTotal: %d", results.TotalCount)
		if results.HasMore {
			fmt.Print(" (more available)")
		}
		fmt.Println()

		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}
