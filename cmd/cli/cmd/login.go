package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	capactCLI "capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/internal/cli/heredoc"
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

// NewLogin returns a cobra.Command for logging into a Hub.
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
			return runLogin(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := login.Flags()

	flags.StringVarP(&opts.user, "username", "u", "", "Username")
	flags.StringVarP(&opts.password, "password", "p", "", "Password")

	return login
}

func runLogin(ctx context.Context, opts loginOptions, w io.Writer) error {
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

	answers.Server = normalizeServerEndpoint(answers.Server)

	creds := credstore.Credentials{
		Username: answers.Username,
		Secret:   answers.Password,
	}
	if err := loginClientSide(ctx, answers.Server, &creds); err != nil {
		return errors.Wrap(err, "while verifying provided credentials")
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

func loginClientSide(ctx context.Context, serverURL string, creds *credstore.Credentials) error {
	cli, err := client.NewClusterWithCreds(serverURL, creds)
	if err != nil {
		return err
	}

	// Only test the credentials, the actual response is irrelevant.
	_, err = cli.GetAction(ctx, "logintest")
	if err != nil {
		return errors.Wrap(err, "while executing get action to test credentials")
	}

	return nil
}

func normalizeServerEndpoint(server string) string {
	if strings.HasPrefix(server, "http://") || strings.HasPrefix(server, "https://") {
		return server
	}

	return fmt.Sprintf("https://%s", server)
}
