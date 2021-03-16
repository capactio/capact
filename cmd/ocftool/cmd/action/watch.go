package action

import (
	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

	"github.com/argoproj/argo/cmd/argo/commands"
	"github.com/argoproj/argo/cmd/argo/commands/client"
	"github.com/spf13/cobra"
)

func NewWatch() *cobra.Command {
	cmd := commands.NewWatchCommand()
	cmd.Use = "watch ACTION"
	cmd.Short = "Watch an Action until it completes"
	cmd.Example = heredoc.WithCLIName(`
		# Watch an Action:
		<cli> action watch my-action
		
		# Watch the latest Action:
		<cli> action watch @latest
	`, ocftool.CLIName)

	client.AddKubectlFlagsToCmd(cmd)

	for _, hide := range hiddenFlags {
		// set flags exits
		_ = cmd.PersistentFlags().MarkHidden(hide)
	}
	return cmd
}
