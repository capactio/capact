package main

import (
	"os"

	"projectvoltron.dev/voltron/cmd/populator/cmd"
)

func main() {
	if err := cmd.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
