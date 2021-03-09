package config

import (
	"github.com/spf13/cobra"
)

func NewConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage configuration for ocftool",
		Long:  "Display or change configuration settings for ocftool",
	}

	cmd.AddCommand(NewGet())
	cmd.AddCommand(NewSet())

	return cmd
}
