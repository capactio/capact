package cmd

import (
	"log"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/config"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/och"
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

	rootCmd.AddCommand(NewValidate())
	rootCmd.AddCommand(NewDocs())
	rootCmd.AddCommand(NewLogin())
	rootCmd.AddCommand(NewLogout())
	rootCmd.AddCommand(och.NewOCH())
	rootCmd.AddCommand(config.NewConfig())

	return rootCmd
}
