package main

import (
	"os"

	"capact.io/capact/cmd/populator/cmd"
)

const CLIName = "populator"

func main() {
	if err := cmd.NewRoot(CLIName).Execute(); err != nil {
		os.Exit(1)
	}
}
