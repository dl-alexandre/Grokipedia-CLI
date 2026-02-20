package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	flags := GlobalFlags{}
	cfg, err := Load(flags)

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults
	if cfg.API.URL != "https://grokipedia.com" {
		t.Errorf("Expected default API URL, got %q", cfg.API.URL)
	}
	if cfg.API.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.API.Timeout)
	}
	if !cfg.Cache.Enabled {
		t.Error("Expected cache enabled by default")
	}
	if cfg.Cache.TTL != 604800 {
		t.Errorf("Expected default TTL 604800, got %d", cfg.Cache.TTL)
	}
	if cfg.Output.Format != "table" {
		t.Errorf("Expected default format 'table', got %q", cfg.Output.Format)
	}
	if cfg.Output.Color != "auto" {
		t.Errorf("Expected default color 'auto', got %q", cfg.Output.Color)
	}
}

func TestLoadWithFlags(t *testing.T) {
	flags := GlobalFlags{
		APIURL:   "https://custom.api.com",
		Timeout:  60,
		NoCache:  true,
		CacheDir: "/custom/cache",
		CacheTTL: 3600,
		Color:    "never",
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.API.URL != "https://custom.api.com" {
		t.Errorf("Expected API URL from flag, got %q", cfg.API.URL)
	}
	if cfg.API.Timeout != 60 {
		t.Errorf("Expected timeout 60 from flag, got %d", cfg.API.Timeout)
	}
	if cfg.Cache.Enabled {
		t.Error("Expected cache disabled from --no-cache flag")
	}
	if cfg.Cache.TTL != 3600 {
		t.Errorf("Expected TTL 3600 from flag, got %d", cfg.Cache.TTL)
	}
	if cfg.Output.Color != "never" {
		t.Errorf("Expected color 'never' from flag, got %q", cfg.Output.Color)
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configContent := `
api:
  url: "https://file.api.com"
  timeout: 45
cache:
  enabled: false
  ttl: 1800
output:
  format: "json"
  color: "always"
`
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	flags := GlobalFlags{
		ConfigFile: configPath,
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.API.URL != "https://file.api.com" {
		t.Errorf("Expected API URL from file, got %q", cfg.API.URL)
	}
	if cfg.API.Timeout != 45 {
		t.Errorf("Expected timeout 45 from file, got %d", cfg.API.Timeout)
	}
	if cfg.Cache.Enabled {
		t.Error("Expected cache disabled from file")
	}
	if cfg.Output.Format != "json" {
		t.Errorf("Expected format 'json' from file, got %q", cfg.Output.Format)
	}
}

func TestLoadFlagsOverrideConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configContent := `
api:
  url: "https://file.api.com"
  timeout: 45
`
	configPath := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	flags := GlobalFlags{
		ConfigFile: configPath,
		APIURL:     "https://flag.api.com",
		Timeout:    90,
	}

	cfg, err := Load(flags)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Flags should override config file
	if cfg.API.URL != "https://flag.api.com" {
		t.Errorf("Expected API URL from flag to override file, got %q", cfg.API.URL)
	}
	if cfg.API.Timeout != 90 {
		t.Errorf("Expected timeout 90 from flag to override file, got %d", cfg.API.Timeout)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "~/.grokipedia",
			expected: filepath.Join(home, ".grokipedia"),
		},
		{
			input:    "~/test/config.yml",
			expected: filepath.Join(home, "test", "config.yml"),
		},
		{
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetCacheDir(t *testing.T) {
	tests := []struct {
		name         string
		cacheDir     string
		wantNotEmpty bool
	}{
		{
			name:         "with custom dir",
			cacheDir:     "/custom/cache",
			wantNotEmpty: true,
		},
		{
			name:         "with empty dir",
			cacheDir:     "",
			wantNotEmpty: true, // Should return default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Cache: CacheConfig{
					Dir: tt.cacheDir,
				},
			}
			got := cfg.GetCacheDir()
			if tt.wantNotEmpty && got == "" {
				t.Error("GetCacheDir() returned empty string")
			}
		})
	}
}

func TestIsCacheEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		ttl      int
		expected bool
	}{
		{
			name:     "enabled with positive TTL",
			enabled:  true,
			ttl:      3600,
			expected: true,
		},
		{
			name:     "disabled by flag",
			enabled:  false,
			ttl:      3600,
			expected: false,
		},
		{
			name:     "disabled by zero TTL",
			enabled:  true,
			ttl:      0,
			expected: false,
		},
		{
			name:     "disabled by negative TTL",
			enabled:  true,
			ttl:      -1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Cache: CacheConfig{
					Enabled: tt.enabled,
					TTL:     tt.ttl,
				},
			}
			if got := cfg.IsCacheEnabled(); got != tt.expected {
				t.Errorf("IsCacheEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetCacheTTL(t *testing.T) {
	tests := []struct {
		name     string
		ttl      int
		expected int
	}{
		{
			name:     "positive TTL",
			ttl:      3600,
			expected: 3600,
		},
		{
			name:     "zero TTL",
			ttl:      0,
			expected: 0,
		},
		{
			name:     "negative TTL returns default",
			ttl:      -1,
			expected: 604800, // default 7 days
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Cache: CacheConfig{
					TTL: tt.ttl,
				},
			}
			if got := cfg.GetCacheTTL(); got != tt.expected {
				t.Errorf("GetCacheTTL() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCommandDefaults(t *testing.T) {
	flags := GlobalFlags{}
	cfg, err := Load(flags)

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check command defaults
	if cfg.Commands.Search.Limit != 12 {
		t.Errorf("Expected search limit 12, got %d", cfg.Commands.Search.Limit)
	}
	if cfg.Commands.Search.Offset != 0 {
		t.Errorf("Expected search offset 0, got %d", cfg.Commands.Search.Offset)
	}
	if cfg.Commands.Search.Format != "table" {
		t.Errorf("Expected search format 'table', got %q", cfg.Commands.Search.Format)
	}
	if cfg.Commands.Edits.Limit != 20 {
		t.Errorf("Expected edits limit 20, got %d", cfg.Commands.Edits.Limit)
	}
}

func TestLoadNonExistentConfigFile(t *testing.T) {
	// Loading without a config file should not error (uses defaults)
	flags := GlobalFlags{}
	_, err := Load(flags)

	if err != nil {
		t.Errorf("Load() with no config file should not error, got %v", err)
	}
}
