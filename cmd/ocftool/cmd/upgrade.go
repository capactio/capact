package cmd

import (
	"fmt"
	"os"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"
	"projectvoltron.dev/voltron/internal/ocftool/upgrade"

	"github.com/spf13/cobra"
)

func NewUpgrade() *cobra.Command {
	var opts upgrade.Options

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrades Capact",
		Long:  "Use this command to upgrade the Capact version on a cluster.",
		Example: heredoc.WithCLIName(`
			# Upgrade Capact components to newest available version
			<cli> upgrade

			# Upgrade Capact components to 0.1.0 version
			<cli> upgrade --version 0.1.0`, ocftool.CLIName),
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
	flags.StringVar(&opts.Parameters.Override.HelmRepoURL, "helm-repo-url", upgrade.CapactioHelmRepoOfficial, fmt.Sprintf("Capact Helm chart repository URL. Use alias %s to select repository which holds master Helm chart versions.", upgrade.CapactioHelmRepoMasterTag))
	flags.StringVar(&opts.Parameters.Override.Docker.Tag, "override-capact-image-tag", "", "Allows you to override Docker image tag for Capact components. By default Docker image tag from Helm chart is used.")
	flags.StringVar(&opts.Parameters.Override.Docker.Repository, "override-capact-image-repo", "", "Allows you to override Docker image repository for Capact components. By default Docker image repository from Helm chart is used.")
	flags.BoolVar(&opts.Parameters.IncreaseResourceLimits, "increase-resource-limits", true, "Enables higher resource requests and limits for components.")
	flags.DurationVar(&opts.Timeout, "timeout", 10*time.Minute, `Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVarP(&opts.Wait, "wait", "w", false, `Waits for the upgrade process until it finish or the defined "--timeout" occurs.`)

	return cmd
}
