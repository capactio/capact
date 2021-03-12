package cmd

import (
	"io"
	"os"

	"github.com/fatih/color"
	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

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
		`, ocftool.CLIName),
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
	if err := loginClientSide(creds); err != nil {
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

func loginClientSide(_ credstore.Credentials) error {
	// TODO check whether provided creds allow us to auth into the given server
	return nil
}
