package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/cli-tools/cache"
	cliver "github.com/dl-alexandre/cli-tools/version"
	"github.com/grokipedia/cli/internal/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info in cli-tools
	cliver.Version = version
	cliver.GitCommit = gitCommit
	cliver.BuildTime = buildTime
	cliver.BinaryName = "grokipedia"

	var c cli.CLI
	ctx := kong.Parse(&c,
		kong.Name("grokipedia"),
		kong.Description("A CLI for the Grokipedia API"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	if ctx.Command() == "version" {
		fmt.Printf("grokipedia %s (commit: %s) built %s\n", cliver.Version, cliver.GitCommit, cliver.BuildTime)
		os.Exit(0)
	}

	// Run auto-update check in background (after initialization)
	// This runs asynchronously and won't block the main command
	go func() {
		// Small delay to not interfere with command output
		time.Sleep(100 * time.Millisecond)

		// Use a minimal cache for update checks
		updateCache := cache.New(cache.DefaultDir("grokipedia"), 24*time.Hour)
		cli.AutoUpdateCheck(updateCache)
	}()

	if err := ctx.Run(&c.Globals); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
