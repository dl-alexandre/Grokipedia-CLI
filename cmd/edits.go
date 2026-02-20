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
	editsLimit       int
	editsStatus      string
	editsExcludeUser []string
	editsCounts      bool
	editsFormat      string
)

// editsCmd represents the edits command
var editsCmd = &cobra.Command{
	Use:   "edits",
	Short: "List edit requests",
	Long:  `Retrieve a list of edit requests from the Grokipedia API.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate format
		allowedFormats := []string{"table", "json"}
		if err := formatter.ValidateFormat(editsFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Parse status filter
		var statusList []string
		if editsStatus != "" {
			statusList = strings.Split(editsStatus, ",")
			for i := range statusList {
				statusList[i] = strings.TrimSpace(statusList[i])
			}
		}

		// Build cache key params
		cacheParams := map[string]interface{}{
			"limit":         editsLimit,
			"includeCounts": editsCounts,
		}
		if len(statusList) > 0 {
			cacheParams["status"] = editsStatus
		}
		if len(editsExcludeUser) > 0 {
			cacheParams["excludeUsers"] = strings.Join(editsExcludeUser, ",")
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/list-edit-requests", cacheParams)
			if data, found := c.Get(cacheKey); found {
				var cached api.EditsResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputEditsResults(&cached, editsFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		results, err := client.Edits(editsLimit, statusList, editsExcludeUser, editsCounts)
		if err != nil {
			return err
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(results); err == nil {
				_ = c.Set(cacheKey, data)
			}
		}

		return outputEditsResults(results, editsFormat)
	},
}

func init() {
	rootCmd.AddCommand(editsCmd)

	// Get defaults from config
	cfg := getConfig()
	defaultLimit := 20
	if cfg != nil {
		defaultLimit = cfg.Commands.Edits.Limit
	}

	editsCmd.Flags().IntVar(&editsLimit, "limit", defaultLimit, "Maximum number of results (1-100)")
	editsCmd.Flags().StringVar(&editsStatus, "status", "", "Filter by status (comma-separated: approved,implemented,pending)")
	editsCmd.Flags().StringArrayVar(&editsExcludeUser, "exclude-user", []string{}, "Exclude edits by username (repeatable)")
	editsCmd.Flags().BoolVar(&editsCounts, "counts", true, "Include count metadata")
	editsCmd.Flags().StringVar(&editsFormat, "format", "table", "Output format: table, json")
}

// outputEditsResults outputs edit results in the specified format
func outputEditsResults(results *api.EditsResponse, format string) error {
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
		if shouldUseColor() {
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

		if editsCounts {
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
