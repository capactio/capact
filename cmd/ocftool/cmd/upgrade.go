package cmd

import (
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
			# Upgrade Capact components to newest version
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
	flags.StringVar(&opts.Version, "version", "0.1.0", "Capact version")
	flags.BoolVar(&opts.IncreaseResourceLimits, "increase-resource-limits", true, "Enables higher resource requests and limits for components.")
	flags.DurationVar(&opts.Timeout, "timeout", 5*time.Minute, `Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVarP(&opts.Wait, "wait", "w", false, `Waits for the upgrade process until it finish or the defined "--timeout" occurs.`)

	return cmd
}
