package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"capact.io/capact/cmd/cli/cmd/policy"

	"capact.io/capact/cmd/cli/cmd/action"
	"capact.io/capact/cmd/cli/cmd/config"
	"capact.io/capact/cmd/cli/cmd/hub"
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   cli.Name,
		Short: "Collective Capability Manager CLI",
		Long: strings.Join(
			[]string{
				"```",
				figure.NewColorFigure(cli.Name, "mini", "green", true).String(),
				"```\n",
				heredoc.WithCLIName(`
        <cli> - Collective Capability Manager CLI

        A utility that manages Capact resources and assists with creating OCF content.

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

            `, cli.Name),
			},
			"\n",
		),
		Version:      cli.Version,
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
		NewCompletion(),
		NewVersion(),
		hub.NewHub(),
		config.NewConfig(),
		action.NewAction(),
		policy.NewCmd(),
	)

	cobra.OnInitialize(initConfig)

	return rootCmd
}

func initConfig() {
	configPath, err := getConfigPath()
	if err != nil {
		handleError(err)
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(configPath, 0700); err != nil {
				handleError(err)
			}

			if err := viper.SafeWriteConfig(); err != nil {
				handleError(err)
			}
		} else {
			handleError(err)
		}
	}
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".config", "capact"), nil
}

func handleError(err error) {
	fmt.Println(err)
	os.Exit(1)
}
