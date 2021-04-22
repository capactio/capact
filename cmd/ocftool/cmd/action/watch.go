package action

import (
	"capact.io/capact/internal/ocftool"
	"capact.io/capact/internal/ocftool/heredoc"

	"github.com/argoproj/argo/v2/cmd/argo/commands"
	"github.com/argoproj/argo/v2/cmd/argo/commands/client"
	"github.com/spf13/cobra"
)

// TODO: Would be great to have this as a `-w` parameter with the status/details action

func NewWatch() *cobra.Command {
	cmd := commands.NewWatchCommand()
	cmd.Use = "watch ACTION"
	cmd.Short = "Watch an Action until it has completed execution"
	cmd.Long = `
    Watch an Action until it has completed execution

    NOTE:   An action needs to be created and run in order to run this command.
            Furthermore, 'kubectl' has to be configured with the context and default
            namespace set to be the same as the one which the Gateway points to. 
    `
	cmd.Example = heredoc.WithCLIName(`
        # Watch an Action:
        <cli> action watch ACTION

        # Watch the Action which was created last:
        <cli> action watch @latest
    `, ocftool.CLIName)

	client.AddKubectlFlagsToCmd(cmd)

	for _, hide := range hiddenFlags {
		// set flags exits
		_ = cmd.PersistentFlags().MarkHidden(hide)
	}
	return cmd
}
