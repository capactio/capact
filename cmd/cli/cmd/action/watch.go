package action

import (
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/argoproj/argo/v2/cmd/argo/commands"
	"github.com/argoproj/argo/v2/cmd/argo/commands/client"
	"github.com/spf13/cobra"
)

// NewWatch returns a cobra.Command for the "action watch" command.
func NewWatch() *cobra.Command {
	cmd := commands.NewWatchCommand()
	cmd.Use = "watch ACTION"
	cmd.Short = "Watch an Action until it has completed execution"
	cmd.Long = `
    Watch an Action until it has completed execution

    NOTE:   An action needs to be created and run in order to run this command.
            This command calls the Kubernetes API directly. As a result, KUBECONFIG has to be configured
            with the same cluster as the one which the Gateway points to.
    `
	cmd.Example = heredoc.WithCLIName(`
        # Watch an Action:
        <cli> action watch ACTION

        # Watch the Action which was created last:
        <cli> action watch @latest
    `, cli.Name)

	client.AddKubectlFlagsToCmd(cmd)

	for _, hide := range argoHiddenFlags {
		// set flags exits
		_ = cmd.PersistentFlags().MarkHidden(hide)
	}
	return cmd
}
