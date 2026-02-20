package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/formatter"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	constantsKey    string
	constantsFormat string
)

// constantsCmd represents the constants command
var constantsCmd = &cobra.Command{
	Use:   "constants",
	Short: "Retrieve API constants",
	Long:  `Fetch constants and configuration values from the Grokipedia API.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate format
		allowedFormats := []string{"json", "yaml", "table"}
		if err := formatter.ValidateFormat(constantsFormat, allowedFormats); err != nil {
			return &api.InvalidArgsError{Message: err.Error()}
		}

		// Check cache first
		cacheKey := ""
		if c := getCache(); c != nil {
			cacheKey = c.GenerateKey("/api/constants", map[string]interface{}{})
			if data, found := c.Get(cacheKey); found {
				var cached api.ConstantsResponse
				if err := json.Unmarshal(data, &cached); err == nil {
					return outputConstantsResults(cached, constantsKey, constantsFormat)
				}
			}
		}

		// Make API request
		client := getClient()
		results, err := client.Constants()
		if err != nil {
			return err
		}

		// Cache the response
		if c := getCache(); c != nil && cacheKey != "" {
			if data, err := json.Marshal(results); err == nil {
				_ = c.Set(cacheKey, data)
			}
		}

		return outputConstantsResults(results, constantsKey, constantsFormat)
	},
}

func init() {
	rootCmd.AddCommand(constantsCmd)

	constantsCmd.Flags().StringVar(&constantsKey, "key", "", "Filter to a single constant key")
	constantsCmd.Flags().StringVar(&constantsFormat, "format", "json", "Output format: json, yaml, table")
}

// outputConstantsResults outputs constants in the specified format
func outputConstantsResults(results api.ConstantsResponse, key string, format string) error {
	// Filter by key if specified
	if key != "" {
		value, ok := results[key]
		if !ok {
			return &api.UnknownConstantError{Key: key}
		}
		results = api.ConstantsResponse{key: value}
	}

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)

	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		defer enc.Close()
		return enc.Encode(results)

	case "table":
		// Get sorted keys
		keys := make([]string, 0, len(results))
		for k := range results {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		if len(keys) == 0 {
			fmt.Println("No constants found.")
			return nil
		}

		tbl := table.New("Key", "Value").WithWriter(os.Stdout)
		if shouldUseColor() {
			tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
				return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, vals...))
			})
		}

		for _, k := range keys {
			v := results[k]
			valueStr := fmt.Sprintf("%v", v)
			// Truncate long values
			if len(valueStr) > 80 {
				valueStr = valueStr[:77] + "..."
			}
			tbl.AddRow(k, valueStr)
		}

		tbl.Print()
		return nil

	default:
		return &api.InvalidArgsError{Message: fmt.Sprintf("invalid format '%s'", format)}
	}
}
