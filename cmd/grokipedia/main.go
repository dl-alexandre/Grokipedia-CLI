package main

import "github.com/dl-alexandre/Grokipedia-CLI/internal/app"

var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	app.Run(version, gitCommit, buildTime)
}
