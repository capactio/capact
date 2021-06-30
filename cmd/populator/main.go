package main

import (
	"os"

	"capact.io/capact/cmd/populator/cmd"
)

const cliName = "populator"

func main() {
	if err := cmd.NewRoot(cliName).Execute(); err != nil {
		os.Exit(1)
	}
}
