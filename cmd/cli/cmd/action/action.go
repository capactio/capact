package action

import (
	"github.com/spf13/cobra"
)

var argoHiddenFlags = []string{
	"as",
	"as-group",
	"certificate-authority",
	"client-certificate",
	"client-key",
	"cluster",
	"context",
	"help",
	"insecure-skip-tls-verify",
	"kubeconfig",
	"no-utf8",
	"node-field-selector",
	"password",
	"request-timeout",
	"server",
	"token",
	"user",
	"username",
	"node-field-selector",
}

// NewCmd returns a new cobra.Command for the "action" subcommand.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "action",
		Aliases: []string{"act"},
		Short:   "This command consists of multiple subcommands to interact with target Actions",
	}

	root.AddCommand(
		NewCreate(),
		NewDelete(),
		NewRun(),
		NewGet(),
		NewWatch(),
	)
	return root
}
