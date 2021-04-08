package main

import (
	"os"

	"projectvoltron.dev/voltron/cmd/populator/cmd"
)

const CLIName = "populator"

func main() {
	if err := cmd.NewRoot(CLIName).Execute(); err != nil {
		os.Exit(1)
	}
}
