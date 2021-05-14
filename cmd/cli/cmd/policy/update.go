package policy

import (
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/policy"
	"github.com/spf13/cobra"
)

func NewUpdate() *cobra.Command {
	var opts policy.UpdateOptions

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates current Policy with new value",
		Example: heredoc.WithCLIName(`
		# Updates the Policy using default editor
		<cli> policy update
		
		# Updates the Policy using content from file
		<cli> policy update --from-file=/tmp/policy.yaml
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return policy.Update(cmd.Context(), opts, os.Stdout)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.PolicyFilePath, "from-file", "", "The new Policy content in YAML format")
	return cmd
}
