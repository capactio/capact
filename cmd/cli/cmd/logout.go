package cmd

import (
	"io"
	"os"

	capactCLI "capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/fatih/color"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/cli/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [SERVER]",
		Short: "Logout from the Hub (Gateway) server",
		Example: heredoc.WithCLIName(`
			# Select what server to log out of via a prompt			
			<cli> logout
			
			# Logout of a specified Hub server
			<cli> logout localhost:8080
		`, capactCLI.Name),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var serverAddress string
			if len(args) > 0 {
				serverAddress = args[0]
			}

			return runLogout(serverAddress, os.Stdout)
		},
	}

	return cmd
}

func runLogout(serverAddress string, w io.Writer) error {
	if serverAddress == "" {
		answer, err := askWhatServerToLogout()
		if err != nil {
			return err
		}
		serverAddress = answer
	}

	if err := credstore.DeleteHub(serverAddress); err != nil {
		return errors.Wrap(err, "could not erase credentials")
	}

	// TODO: handle current context update

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Logout Succeeded\n")

	return nil
}

func askWhatServerToLogout() (string, error) {
	candidates, err := credstore.ListHubServer()
	if err != nil {
		return "", err
	}

	if len(candidates) == 0 {
		return "", errors.New("Not logged in to any server")
	}

	var serverAddress string
	err = survey.AskOne(&survey.Select{
		Message: "What server do you want to log out of? ",
		Options: candidates,
	}, &serverAddress)
	if err != nil {
		return "", err
	}

	return serverAddress, nil
}
