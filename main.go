package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/grokipedia/cli/internal/cli"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
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

	ctx.FatalIfErrorf(ctx.Run(&c.Globals))
}
