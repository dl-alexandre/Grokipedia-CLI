package cmd

import (
	"testing"
)

// Integration tests for command workflows
// These tests verify that commands work together properly

func TestCommandHierarchy(t *testing.T) {
	// Verify all expected commands are registered
	expectedCommands := map[string]bool{
		"search":        false,
		"page":          false,
		"typeahead":     false,
		"constants":     false,
		"edits":         false,
		"edits-by-slug": false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, exists := expectedCommands[cmd.Name()]; exists {
			expectedCommands[cmd.Name()] = true
		}
	}

	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Command %q not found in root command", name)
		}
	}
}

func TestGlobalFlagsPersistence(t *testing.T) {
	// Verify global flags are available on all subcommands
	subcommands := []string{"search", "page", "typeahead", "constants", "edits", "edits-by-slug"}

	for _, cmdName := range subcommands {
		cmd, _, err := rootCmd.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Failed to find command %q: %v", cmdName, err)
			continue
		}

		// Check that global flags are inherited
		flags := []string{"api-url", "timeout", "no-cache", "verbose", "debug", "color"}
		for _, flagName := range flags {
			if cmd.PersistentFlags().Lookup(flagName) == nil && rootCmd.PersistentFlags().Lookup(flagName) == nil {
				t.Errorf("Command %q missing global flag %q", cmdName, flagName)
			}
		}
	}
}

func TestCommandUsage(t *testing.T) {
	// Verify commands have proper usage strings
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Fatal("No commands registered")
	}

	for _, cmd := range commands {
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			continue // Skip built-in commands
		}

		if cmd.Short == "" {
			t.Errorf("Command %q missing Short description", cmd.Name())
		}

		if cmd.Long == "" && cmd.Short == "" {
			t.Errorf("Command %q missing both Short and Long descriptions", cmd.Name())
		}
	}
}

func TestFlagDefaults(t *testing.T) {
	// Test that flag defaults are properly set
	tests := []struct {
		flagName string
		expected string
	}{
		{"color", "auto"},
	}

	for _, tt := range tests {
		flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("Flag %q not found", tt.flagName)
			continue
		}

		if flag.DefValue != tt.expected {
			t.Errorf("Flag %q default value = %q, want %q", tt.flagName, flag.DefValue, tt.expected)
		}
	}
}
