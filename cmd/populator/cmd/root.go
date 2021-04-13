package cmd

import (
	"log"

	"capact.io/capact/cmd/populator/cmd/register"

	"github.com/spf13/cobra"
)

func NewRoot(cliName string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          cliName,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(
		register.NewRegister(cliName),
	)

	return rootCmd
}
