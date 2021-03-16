package main

import (
	"os"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
