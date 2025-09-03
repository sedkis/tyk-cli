package main

import (
	"fmt"
	"os"

	"github.com/tyktech/tyk-cli/internal/cli"
)

// Build-time variables (set by ldflags)
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	rootCmd := cli.NewRootCommand(version, commit, buildTime)
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}