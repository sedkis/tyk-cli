package main

import (
	"errors"
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
		// Check for ExitError to use specific exit codes
		var exitError *cli.ExitError
		if errors.As(err, &exitError) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", exitError.Message)
			os.Exit(exitError.Code)
		}
		
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}