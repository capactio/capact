package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const CLIName = "populator"

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          CLIName,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(
		NewHubDatabase(),
		NewInstallTypeInstances(),
	)

	return rootCmd
}
