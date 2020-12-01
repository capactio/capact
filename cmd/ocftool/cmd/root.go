package cmd

import (
	"log"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd/validate"

	"github.com/spf13/cobra"
)

const (
	appName = "ocftool"
	version = "0.0.1"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     appName,
		Short:   "CLI tool for working with OCF manifest files",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(validate.NewCmd())

	return rootCmd
}
