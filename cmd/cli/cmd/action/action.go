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
	"tls-server-name",
	"token",
	"user",
	"username",
	"node-field-selector",
}

// NewCmd returns a new cobra.Command subcommand for Action related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "action",
		Aliases: []string{"act", "actions"},
		Short:   "This command consists of multiple subcommands to interact with target Actions",
	}

	root.AddCommand(
		NewCreate(),
		NewDelete(),
		NewRun(),
		NewGet(),
		NewWatch(),
		NewWait(),
		NewLogs(),
	)
	return root
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
