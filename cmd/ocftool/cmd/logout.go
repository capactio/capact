package cmd

import (
	"projectvoltron.dev/voltron/internal/ocftool/credstore"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/docker/cli/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewLogout creates a new `ocftool logout` command
func NewLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [SERVER]",
		Short: "Log out from a Gateway server",
		Example: heredoc.Doc(`
			ocftool logout
			# => select what server to log out of via a prompt
			
			ocftool logout localhost:8080
			# => log out of specified server
		`),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var serverAddress string
			if len(args) > 0 {
				serverAddress = args[0]
			}

			return runLogout(serverAddress)
		},
	}

	return cmd
}

func runLogout(serverAddress string) error {
	store := credstore.NewOCH()

	if serverAddress == "" {
		answer, err := askWhatServerToLogout(store)
		if err != nil {
			return err
		}
		serverAddress = answer
	}

	if err := store.Delete(serverAddress); err != nil {
		return errors.Wrap(err, "could not erase credentials")
	}

	// TODO: handle current context update

	return nil
}

func askWhatServerToLogout(store credstore.Store) (string, error) {
	out, err := store.List()
	if err != nil {
		return "", err
	}

	var candidates []string
	for k := range out {
		candidates = append(candidates, k)
	}

	if len(candidates) == 0 {
		return "", errors.New("not logged in to any server")
	}

	var serverAddress string
	err = survey.AskOne(&survey.Select{
		Message: "What server do you want to log out of?",
		Options: candidates,
	}, &serverAddress)
	if err != nil {
		return "", err
	}

	return serverAddress, nil
}
