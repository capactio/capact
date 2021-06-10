package action

import (
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/action"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/spf13/cobra"
)

func NewDelete() *cobra.Command {
	var opts action.DeleteOptions

	cmd := &cobra.Command{
		Use:   "delete [ACTION_NAME...]",
		Short: "Deletes the Action",
		Example: heredoc.WithCLIName(`
		# Deletes the foo Action in the default namespace
		<cli> action delete foo
		
		# Deletes all Actions with 'upgrade-' prefix in the foo namespace
		<cli> action delete --name-regex='upgrade-*' --namespace=foo
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ActionNames = args
			err := action.Delete(cmd.Context(), opts, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where the Action was created")
	flags.StringVar(&opts.NameRegex, "name-regex", "", "Deletes all Actions whose names are matched by the given regular expression. To check the regex syntax, read: https://golang.org/s/re2syntax")
	flags.DurationVar(&opts.Timeout, "timeout", 10*time.Minute, `Maximum time during which the deletion process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVarP(&opts.Wait, "wait", "w", true, `Waits for the deletion process until it finish or the defined "--timeout" occurs.`)
	// TODO: support phases also when exact name is used.
	flags.StringVar(&opts.Phase, "phase", "", fmt.Sprintf("Deletes Actions only in the given phase. Supported only when the --name-regex flag is used. Allowed values: %s", action.AllowedPhases()))

	return cmd
}
