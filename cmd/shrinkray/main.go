package main

import (
	"os"

	"github.com/jparkerweb/shrinkray/internal/cli"
)

// Build info — injected via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetBuildInfo(version, commit, date)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
