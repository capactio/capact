package cmd

import (
	"log"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd/action"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/config"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/och"

	"github.com/spf13/cobra"
)

const (
	appName = "ocftool"
	version = "0.2.0"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          appName,
		Short:        "CLI tool for working with OCF manifest files",
		Version:      version,
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
		och.NewOCH(),
		config.NewConfig(),
		action.NewAction(),
	)

	return rootCmd
}
