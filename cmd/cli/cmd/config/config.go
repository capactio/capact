package config

import (
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for CLI configuration related operations.
// TODO: Add support for target configuration (instead of relying on current default context in kubectl)
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Display or change configuration settings for the Hub",
	}

	cmd.AddCommand(NewGet())
	cmd.AddCommand(NewSet())

	return cmd
}
