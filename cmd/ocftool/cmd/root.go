package cmd

import (
	"log"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd/config"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/hub"
	"projectvoltron.dev/voltron/internal/ocftool"

	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          ocftool.CLIName,
		Short:        "CLI tool for working with OCF manifest files",
		Version:      ocftool.Version,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(
		NewValidate(),
		NewDocs(),
		NewLogin(),
		NewLogout(),
		hub.NewHub(),
		config.NewConfig(),
	)

	return rootCmd
}
