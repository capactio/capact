package action

import (
	"os"
	"time"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/action"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/spf13/cobra"
)

// NewWait returns a new cobra.Command for waiting for a given Action's condition.
func NewWait() *cobra.Command {
	var opts action.WaitOptions

	cmd := &cobra.Command{
		Use:   "wait ACTION",
		Short: "Wait for a specific condition on a given Action",
		Args:  cobra.ExactArgs(1),
		Example: heredoc.WithCLIName(`
			# Wait for the Actin "example" to contain the phase "READY_TO_RUN"
			<cli> act wait --for=phase=READY_TO_RUN example
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ActionName = args[0]
			return action.Wait(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.For, "for", "", "The filed condition to wait on. Currently, only the 'phase' filed is supported 'phase=phase-name'.")
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where the Action was created.")
	flags.DurationVar(&opts.Timeout, "wait-timeout", 10*time.Minute, `Maximum time to wait before giving up. "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	client.RegisterFlags(flags)

	panicOnError(cmd.MarkFlagRequired("for")) // this cannot happen

	return cmd
}
