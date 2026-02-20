package cmd

import (
	"fmt"
	"os"

	"github.com/grokipedia/cli/internal/api"
	"github.com/grokipedia/cli/internal/cache"
	"github.com/grokipedia/cli/internal/config"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	apiURL    string
	timeout   int
	noCache   bool
	cacheDir  string
	cacheTTL  int
	verbose   bool
	debug     bool
	colorMode string

	appConfig *config.Config
	appCache  *cache.Cache
	appClient *api.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "grokipedia",
	Short: "A CLI for the Grokipedia API",
	Long: `Grokipedia CLI provides access to the Grokipedia knowledge base.

Use the search command to find pages, the page command to view content,
and other commands to interact with edit requests and API constants.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip initialization for help and completion commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		// Load configuration
		flags := config.GlobalFlags{
			APIURL:     apiURL,
			Timeout:    timeout,
			NoCache:    noCache,
			CacheDir:   cacheDir,
			CacheTTL:   cacheTTL,
			Verbose:    verbose,
			Debug:      debug,
			ConfigFile: cfgFile,
			Color:      colorMode,
		}

		var err error
		appConfig, err = config.Load(flags)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize cache if enabled
		if !noCache && appConfig.IsCacheEnabled() {
			appCache = cache.New(
				appConfig.GetCacheDir(),
				appConfig.GetCacheTTL(),
			)
		}

		// Initialize API client
		appClient = api.NewClient(api.ClientOptions{
			BaseURL: appConfig.API.URL,
			Timeout: appConfig.API.Timeout,
			Verbose: verbose,
			Debug:   debug,
		})

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		exitCode := api.GetExitCode(err)
		os.Exit(exitCode)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grokipedia/config.yml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API base URL (env: GROKIPEDIA_API_URL)")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 0, "Request timeout in seconds (env: GROKIPEDIA_TIMEOUT)")
	rootCmd.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Disable caching (env: GROKIPEDIA_NO_CACHE)")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "Cache directory (env: GROKIPEDIA_CACHE_DIR)")
	rootCmd.PersistentFlags().IntVar(&cacheTTL, "cache-ttl", 0, "Cache TTL in seconds (env: GROKIPEDIA_CACHE_TTL)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output (env: GROKIPEDIA_VERBOSE)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (env: GROKIPEDIA_DEBUG)")
	rootCmd.PersistentFlags().StringVar(&colorMode, "color", "auto", "Color mode: auto, always, never (env: GROKIPEDIA_COLOR)")
}

func initConfig() {
	// Configuration is loaded in PersistentPreRunE
}

// shouldUseColor determines if color output should be used
func shouldUseColor() bool {
	switch colorMode {
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

// getCache returns the cache instance if enabled
func getCache() *cache.Cache {
	return appCache
}

// getClient returns the API client
func getClient() *api.Client {
	return appClient
}

// getConfig returns the loaded configuration
func getConfig() *config.Config {
	return appConfig
}
