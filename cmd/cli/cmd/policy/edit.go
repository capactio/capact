package policy

import (
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/policy"

	"github.com/spf13/cobra"
)

// NewEdit returns a cobra.Command for interactive editing
// of Capact Global policy on a Capact environment.
func NewEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edits current Policy in place using interactive mode",
		Example: heredoc.WithCLIName(`
		# Updates the Policy using default editor
		<cli> policy edit
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return policy.Edit(cmd.Context(), os.Stdout)
		},
	}

	flags := cmd.Flags()
	client.RegisterFlags(flags)

	return cmd
}
