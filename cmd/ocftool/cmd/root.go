package cmd

import (
	"log"
	"strings"

	"projectvoltron.dev/voltron/cmd/ocftool/cmd/action"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/config"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/hub"
	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   ocftool.CLIName,
		Short: "Collective Capability Manager CLI",
		Long: strings.Join([]string{figure.NewColorFigure(ocftool.CLIName, "mini", "green", true).String(),
			heredoc.WithCLIName(`
        <cli> - Collective Capability Manager CLI

        A utility for managing Capact & assist with authoring OCF content

        To begin working with Capact using the <cli> CLI, start with:

            $ <cli> login

        NOTE: If you would like to use 'pass' for credential storage, be sure to
              set CAPACT_CREDENTIALS_STORE_BACKEND to 'pass' in your shell's env variables.

              In order to watch follow the progress of the workflow execution, it is required
              to have 'kubectl' configured with the default context set to the same cluster where
              the Gateway URL points to.

        Quick Start:

            $ <cli> hub interfaces search                    # Lists available content (generic interfaces)
            $ <cli> hub interfaces browse                    # Interactively browse available content in your terminal
            $ <cli> action search                            # Lists available actions in the 'default' namespace
            $ <cli> action get <action name> -n <namespace>  # Gets the details of a specified action in the specified namespace (table format)
            $ <cli> action get <action name> -o json         # Gets the details of a specified action in the 'default' namespace (JSON format)
            $ <cli> action run <action name>                 # Accepts the rendered action, and sends it to the workflow engine
            $ <cli> action status @latest                    # Gets the status of the last triggered action
            $ <cli> action watch <action name>               # Watches the workflow engine's progress while processing the specified action

            `, ocftool.CLIName)}, "\n"),
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
		NewUpgrade(),
		hub.NewHub(),
		config.NewConfig(),
		action.NewAction(),
	)

	return rootCmd
}
