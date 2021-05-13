package main

import (
	"os"

	"capact.io/capact/cmd/cli/cmd"
	"capact.io/capact/internal/cli/config"
)

func main() {
	rootCmd := cmd.NewRoot()

	if err := config.ReadConfig(); err != nil {
		panic(err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
