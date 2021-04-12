package action

import (
	"fmt"
	"os"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/action"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

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
		
		# Deletes all Actions with upgrade- prefix in the foo namespace
		<cli> action delete --name-regex='upgrade-*' --namespace=foo
		`, ocftool.CLIName),
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
	// TODO: support phases also when exact name is used.
	flags.StringVar(&opts.Phase, "phase", "", fmt.Sprintf("Deletes Actions only in the given phase. Supported only when the --name-regex flag is used. Allowed values: %s", action.AllowedPhases()))

	return cmd
}
