package cmd

import (
	"context"
	"io"
	"os"

	capactCLI "capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/pkg/engine/api/graphql"
	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	serverAddress string
	user          string
	password      string
}

// TODO: Validate the Gateway Hub URL

func NewLogin() *cobra.Command {
	var opts loginOptions

	login := &cobra.Command{
		Use:   "login [OPTIONS] [SERVER]",
		Short: "Login to a Hub (Gateway) server",
		Example: heredoc.WithCLIName(`
			# start interactive setup
			<cli> login

			# Specify server name and specify the user
			<cli> login localhost:8080 -u user
		`, capactCLI.Name),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.serverAddress = args[0]
			}
			return runLogin(opts, os.Stdout)
		},
	}

	flags := login.Flags()

	flags.StringVarP(&opts.user, "username", "u", "", "Username")
	flags.StringVarP(&opts.password, "password", "p", "", "Password")

	return login
}

func runLogin(opts loginOptions, w io.Writer) error {
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
				Message: "Gateway's server address: ",
			},
			Validate: survey.Required,
		})
	}
	if answers.Username == "" {
		qs = append(qs, &survey.Question{
			Name: "username",
			Prompt: &survey.Input{
				Message: "Username: ",
			},
			Validate: survey.Required,
		})
	}

	if answers.Password == "" {
		qs = append(qs, &survey.Question{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password: ",
			},
			Validate: survey.Required,
		})
	}

	// perform the questions if needed
	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	creds := credstore.Credentials{
		Username: answers.Username,
		Secret:   answers.Password,
	}
	if err := loginClientSide(answers.Server, &creds); err != nil {
		return err
	}

	if err = credstore.AddHub(answers.Server, creds); err != nil {
		return err
	}

	if err = config.SetAsDefaultContext(answers.Server, false); err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Login Succeeded\n")

	return nil
}

func loginClientSide(serverURL string, creds *credstore.Credentials) error {
	cli, err := client.NewClusterWithCreds(serverURL, creds)
	if err != nil {
		return err
	}

	_, err = cli.ListActions(context.Background(), &graphql.ActionFilter{})
	if err != nil {
		return errors.Wrap(err, "while executing get action to test credentials")
	}

	return nil
}
