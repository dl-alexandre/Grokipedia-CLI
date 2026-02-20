package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	API      APIConfig      `mapstructure:"api"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Output   OutputConfig   `mapstructure:"output"`
	Commands CommandsConfig `mapstructure:"commands"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	TTL     int    `mapstructure:"ttl"`
	Dir     string `mapstructure:"dir"`
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  string `mapstructure:"color"`
}

// CommandsConfig holds command-specific defaults
type CommandsConfig struct {
	Search SearchConfig `mapstructure:"search"`
	Edits  EditsConfig  `mapstructure:"edits"`
}

// SearchConfig holds search command defaults
type SearchConfig struct {
	Limit  int    `mapstructure:"limit"`
	Offset int    `mapstructure:"offset"`
	Format string `mapstructure:"format"`
}

// EditsConfig holds edits command defaults
type EditsConfig struct {
	Limit int `mapstructure:"limit"`
}

// GlobalFlags holds CLI flag values that override config
type GlobalFlags struct {
	APIURL     string
	Timeout    int
	NoCache    bool
	CacheDir   string
	CacheTTL   int
	Verbose    bool
	Debug      bool
	ConfigFile string
	Color      string
}

// Load loads configuration from file, environment, and flags
// Precedence: flags > env vars > config file > defaults
func Load(flags GlobalFlags) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file if provided
	if flags.ConfigFile != "" {
		v.SetConfigFile(expandPath(flags.ConfigFile))
	} else {
		// Default config location
		configDir := getDefaultConfigDir()
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	bindEnvVars(v)

	// Override with CLI flags
	applyFlags(v, flags)

	// Unmarshal to struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand paths in config
	cfg.Cache.Dir = expandPath(cfg.Cache.Dir)

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("api.url", "https://grokipedia.com")
	v.SetDefault("api.timeout", 30)

	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.ttl", 604800) // 7 days
	v.SetDefault("cache.dir", "~/.grokipedia/cache")

	v.SetDefault("output.format", "table")
	v.SetDefault("output.color", "auto")

	v.SetDefault("commands.search.limit", 12)
	v.SetDefault("commands.search.offset", 0)
	v.SetDefault("commands.search.format", "table")

	v.SetDefault("commands.edits.limit", 20)
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("GROKIPEDIA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit bindings for nested keys
	v.BindEnv("api.url", "GROKIPEDIA_API_URL")
	v.BindEnv("api.timeout", "GROKIPEDIA_TIMEOUT")
	v.BindEnv("cache.enabled", "GROKIPEDIA_NO_CACHE")
	v.BindEnv("cache.ttl", "GROKIPEDIA_CACHE_TTL")
	v.BindEnv("cache.dir", "GROKIPEDIA_CACHE_DIR")
	v.BindEnv("output.color", "GROKIPEDIA_COLOR")
}

// applyFlags applies CLI flag values to viper
func applyFlags(v *viper.Viper, flags GlobalFlags) {
	if flags.APIURL != "" {
		v.Set("api.url", flags.APIURL)
	}
	if flags.Timeout > 0 {
		v.Set("api.timeout", flags.Timeout)
	}
	if flags.NoCache {
		v.Set("cache.enabled", false)
	}
	if flags.CacheDir != "" {
		v.Set("cache.dir", flags.CacheDir)
	}
	if flags.CacheTTL != 0 {
		v.Set("cache.ttl", flags.CacheTTL)
	}
	if flags.Color != "" {
		v.Set("output.color", flags.Color)
	}
}

// getDefaultConfigDir returns the default configuration directory
func getDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".grokipedia")
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}

// GetCacheDir returns the expanded cache directory path
func (c *Config) GetCacheDir() string {
	if c.Cache.Dir == "" {
		return filepath.Join(getDefaultConfigDir(), "cache")
	}
	return c.Cache.Dir
}

// IsCacheEnabled returns true if caching is enabled
func (c *Config) IsCacheEnabled() bool {
	return c.Cache.Enabled && c.Cache.TTL > 0
}

// GetCacheTTL returns the cache TTL in seconds
func (c *Config) GetCacheTTL() int {
	if c.Cache.TTL < 0 {
		return 604800 // default 7 days
	}
	return c.Cache.TTL
}
