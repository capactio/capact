package register

import (
	"capact.io/capact/internal/cli/testing"
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"

	"capact.io/capact/internal/cli/heredoc"
)

// NewTestStorageBackend returns a cobra.Command for populating storage backend TypeInstance for testing purposes.
func NewTestStorageBackend(cliName string) *cobra.Command {
	return &cobra.Command{
		Use:   "test-storage-backend",
		Short: "Produces and uploads TypeInstances which describe storage backend for testing purposes",
		Example: heredoc.WithCLIName(`
			<cli> test-storage-backend
		`, cliName),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			storageRegister, err := testing.NewStorageBackendRegister()
			if err != nil {
				return err
			}
			return storageRegister.RegisterTypeInstances(cmd.Context())
		},
	}
}
