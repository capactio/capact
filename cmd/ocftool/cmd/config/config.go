package config

import (
	"github.com/spf13/cobra"
)

func NewConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Display or change configuration settings for hub and target",
	}

	cmd.AddCommand(NewGet())
	cmd.AddCommand(NewSet())

	return cmd
}
