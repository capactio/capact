package policy

import (
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/policy"
	"github.com/spf13/cobra"
)

// NewApply returns a cobra.Command for applying Capact Global policy on a Capact environment.
func NewApply() *cobra.Command {
	var opts policy.ApplyOptions

	cmd := &cobra.Command{
		Use:   "apply -f {path}",
		Short: "Updates current Policy with new value",
		Example: heredoc.WithCLIName(`
		# Updates the Policy using content from file
		<cli> policy apply -f /tmp/policy.yaml
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return policy.Apply(cmd.Context(), opts, os.Stdout)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.PolicyFilePath, cli.FromFileFlagName, "f", "", "The path to new Policy in YAML format")
	panicOnError(cmd.MarkFlagRequired(cli.FromFileFlagName)) // this cannot happen

	return cmd
}
