package main

import (
	"os"

	"capact.io/capact/cmd/cli/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
