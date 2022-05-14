package action

import (
	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands"
	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/spf13/cobra"
)

// NewLogs returns a new cobra.Command for getting Action's logs.
// TODO: this should be done via Gateway once the subscription is implemented.
func NewLogs() *cobra.Command {
	cmd := commands.NewLogsCommand()
	cmd.Use = "logs ACTION [POD]"
	cmd.Short = "Print the Action's logs"
	cmd.Long = heredoc.Doc(`
    Print the Action's logs

    NOTE:   An action needs to be created and run in order to run this command.
            This command calls the Kubernetes API directly. As a result, KUBECONFIG has to be configured
            with the same cluster as the one which the Gateway points to.`)

	cmd.Example = heredoc.WithCLIName(`
			# Print the logs of an Action:
			<cli> logs example

			# Follow the logs of an Action:
			<cli> logs example --follow

			# Print the logs of single container in a pod
			<cli> logs example step-pod -c step-pod-container

			# Print the logs of an Action's step:
			<cli> logs example step-pod

			# Print the logs of the latest executed Action:
			<cli> logs @latest
		`, cli.Name)

	client.AddKubectlFlagsToCmd(cmd)

	for _, hide := range argoHiddenFlags {
		_ = cmd.PersistentFlags().MarkHidden(hide)
	}

	return cmd
}
