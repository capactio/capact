package cmd

import (
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/docker/cli/cli"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

// NewLoginCommand creates a new `och login` command
func NewLogin() *cobra.Command {
	var opts loginOptions

	cmd := &cobra.Command{
		Use:   "login [OPTIONS] [SERVER]",
		Short: "Log in to a Gateway server",
		Example: heredoc.Doc(`
			# start interactive setup
			$ ocftool login

			# specify server name and user 
			$ ocftool login localhost:8080 -u user
		`),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.serverAddress = args[0]
			}
			return runLogin(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.user, "username", "u", "", "Username")
	flags.StringVarP(&opts.password, "password", "p", "", "Password")

	return cmd
}

func runLogin(opts loginOptions) error {
	answers := struct {
		Server   string `survey:"server-address"`
		Username string
		Password string
	}{
		Server:   opts.serverAddress,
		Username: opts.user,
		Password: opts.password,
	}

	var qs []*survey.Question
	if answers.Server == "" {
		qs = append(qs, &survey.Question{
			Name: "server-address",
			Prompt: &survey.Input{
				Message: "Please type Gateway server address",
			},
			Validate: survey.Required,
		})
	}
	if answers.Username == "" {
		qs = append(qs, &survey.Question{
			Name: "username",
			Prompt: &survey.Input{
				Message: "Please type username",
			},
			Validate: survey.Required,
		})
	}

	if answers.Password == "" {
		qs = append(qs, &survey.Question{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Please type password",
			},
			Validate: survey.Required,
		})
	}

	store := credstore.NewOCH()

	// perform the questions if needed
	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	creds := &credentials.Credentials{
		ServerURL: answers.Server,
		Username:  answers.Username,
		Secret:    answers.Password,
	}
	if err := loginClientSide(creds); err != nil {
		return err
	}

	if err = store.Add(creds); err != nil {
		return err
	}

	return config.SetAsDefaultContext(creds.ServerURL, false)
}

func loginClientSide(creds *credentials.Credentials) error {
	// TODO check whether provided creds allow us to auth into the given server
	return nil
}
