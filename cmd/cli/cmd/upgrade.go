package cmd

import (
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/upgrade"

	"github.com/spf13/cobra"
)

func NewUpgrade() *cobra.Command {
	var opts upgrade.Options

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades Capact",
		Long:  "Use this command to upgrade the Capact version on a cluster.",
		Example: heredoc.WithCLIName(`
			# Upgrade Capact components to the newest available version
			<cli> upgrade

			# Upgrade Capact components to 0.1.0 version
			<cli> upgrade --version 0.1.0`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			upgradeProcess, err := upgrade.New(os.Stdout)
			if err != nil {
				return err
			}

			return upgradeProcess.Run(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Parameters.Version, "version", upgrade.LatestVersionTag, "Capact version.")
	flags.StringVar(&opts.Parameters.Override.HelmRepoURL, "helm-repo-url", upgrade.CapactioHelmRepoOfficial, fmt.Sprintf("Capact Helm chart repository URL. Use %s tag to select repository which holds the latest Helm chart versions.", upgrade.CapactioHelmRepoLatestTag))
	flags.StringVar(&opts.Parameters.Override.CapactValues.Global.ContainerRegistry.Tag, "override-capact-image-tag", "", "Allows you to override Docker image tag for Capact components. By default, Docker image tag from Helm chart is used.")
	flags.StringVar(&opts.Parameters.Override.CapactValues.Global.ContainerRegistry.Path, "override-capact-image-repo", "", "Allows you to override Docker image repository for Capact components. By default, Docker image repository from Helm chart is used.")
	flags.BoolVar(&opts.Parameters.IncreaseResourceLimits, "increase-resource-limits", true, "Enables higher resource requests and limits for components.")
	flags.BoolVar(&opts.Parameters.Override.CapactValues.Engine.TestSetup.Enabled, "enable-test-setup", false, "Enables test setup for the Capact E2E validation scenarios.")
	flags.BoolVar(&opts.Parameters.Override.CapactValues.Notes.PrintInsecure, "print-insecure-helm-release-notes", false, "Prints the base64-encoded Gateway password directly in Helm release notes.")
	flags.StringVar(&opts.ActionNamePrefix, "action-name-prefix", "capact-upgrade-", "Specifies Capact upgrade Action name prefix.")
	flags.DurationVar(&opts.Timeout, "timeout", 10*time.Minute, `Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVarP(&opts.Wait, "wait", "w", true, `Waits for the upgrade process until it finish or the defined "--timeout" occurs.`)
	flags.DurationVar(&opts.MaxQueueTime, "max-queue-time", 10*time.Minute, `Maximum waiting time for the completion of other, currently running upgrade tasks. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)

	return cmd
}
