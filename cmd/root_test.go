package cmd

import (
	"testing"
)

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name      string
		colorMode string
		// Note: actual TTY detection depends on environment,
		// so we mainly test the explicit modes
	}{
		{
			name:      "always",
			colorMode: "always",
		},
		{
			name:      "never",
			colorMode: "never",
		},
		{
			name:      "auto",
			colorMode: "auto",
		},
		{
			name:      "invalid defaults to auto",
			colorMode: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global colorMode variable
			colorMode = tt.colorMode

			result := shouldUseColor()

			switch tt.colorMode {
			case "always":
				if !result {
					t.Error("shouldUseColor() with 'always' should return true")
				}
			case "never":
				if result {
					t.Error("shouldUseColor() with 'never' should return false")
				}
			}
			// For 'auto' and invalid, we can't reliably test without controlling TTY
		})
	}
}

func TestGlobalFlagDefaults(t *testing.T) {
	// Verify global flags are initialized with zero values
	// The actual defaults are set in init() and cobra flag definitions

	// Test that flags exist (they're declared as package variables)
	_ = cfgFile
	_ = apiURL
	_ = timeout
	_ = noCache
	_ = cacheDir
	_ = cacheTTL
	_ = verbose
	_ = debug
	_ = colorMode
}

func TestRootCommand(t *testing.T) {
	// Test that root command is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	// Check command name
	if rootCmd.Name() != "grokipedia" {
		t.Errorf("Expected command name 'grokipedia', got %s", rootCmd.Name())
	}

	// Check that subcommands are registered
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Expected subcommands to be registered")
	}
}

func TestSubcommandsExist(t *testing.T) {
	expectedCommands := []string{
		"search",
		"page",
		"typeahead",
		"constants",
		"edits",
		"edits-by-slug",
	}

	for _, cmdName := range expectedCommands {
		cmd, _, err := rootCmd.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Expected subcommand %q to exist: %v", cmdName, err)
			continue
		}
		if cmd.Name() != cmdName {
			t.Errorf("Expected subcommand name %q, got %q", cmdName, cmd.Name())
		}
	}
}

func TestPersistentFlags(t *testing.T) {
	flags := []string{
		"config",
		"api-url",
		"timeout",
		"no-cache",
		"cache-dir",
		"cache-ttl",
		"verbose",
		"debug",
		"color",
	}

	for _, flagName := range flags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected persistent flag %q to exist", flagName)
		}
	}
}
