package main

import (
	"github.com/alecthomas/kong"
	"github.com/grokipedia/cli/internal/cli"
)

func main() {
	var c cli.CLI
	ctx := kong.Parse(&c,
		kong.Name("grokipedia"),
		kong.Description("A CLI for the Grokipedia API"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run(&c.Globals))
}
