package app

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/Grokipedia-CLI/internal/cli"
	cliver "github.com/dl-alexandre/cli-tools/version"
)

// Run starts the Grokipedia CLI using linker-provided build metadata.
func Run(linkedVersion, linkedCommit, linkedBuildTime string) {
	version := resolveVersion(linkedVersion)
	cliver.Version = version
	cliver.GitCommit = linkedCommit
	cliver.BuildTime = linkedBuildTime
	cliver.BinaryName = "grokipedia"

	var commandLine cli.CLI
	ctx := kong.Parse(&commandLine,
		kong.Name("grokipedia"),
		kong.Description("Unofficial CLI for the Grokipedia API"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	if ctx.Command() == "version" {
		fmt.Printf("grokipedia %s (commit: %s) built %s\n", cliver.Version, cliver.GitCommit, cliver.BuildTime)
		os.Exit(0)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cli.AutoUpdateCheck()
	}()

	if err := ctx.Run(&commandLine.Globals); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func resolveVersion(linkedVersion string) string {
	if linkedVersion != "" && linkedVersion != "dev" {
		return linkedVersion
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		version := buildInfo.Main.Version
		if version != "" && version != "(devel)" {
			return version
		}
	}

	return linkedVersion
}
