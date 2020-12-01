package main

import (
	"log"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}
