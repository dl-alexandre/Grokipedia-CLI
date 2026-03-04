package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/grokipedia/cli/internal/cache"
	"github.com/grokipedia/cli/internal/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info for update checking
	cli.Version = version
	cli.GitCommit = gitCommit
	cli.BuildTime = buildTime
	cli.BinaryName = "grokipedia"
	cli.GitHubRepo = "grokipedia-cli"

	var c cli.CLI
	ctx := kong.Parse(&c,
		kong.Name("grokipedia"),
		kong.Description("A CLI for the Grokipedia API"),
		kong.UsageOnError(),
	)

	// If version flag was passed, print version and exit
	if ctx.Command() == "version" || (len(ctx.Args) > 0 && ctx.Args[0] == "--version") {
		fmt.Printf("grokipedia %s (%s) built %s\n", version, gitCommit, buildTime)
		return
	}

	// Run auto update check in background (non-blocking)
	// Initialize cache for update checking if not disabled
	var updateCache *cache.Cache
	if !c.NoCache {
		updateCache = cache.New("", 24*60*60) // Default cache dir, 24 hour TTL
	}
	cli.AutoUpdateCheck(updateCache)

	ctx.FatalIfErrorf(ctx.Run(&c.Globals))
}
