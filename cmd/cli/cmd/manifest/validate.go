package manifest

import (
	"os"

	"capact.io/capact/internal/cli/validate"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/spf13/cobra"
)

const defaultMaxConcurrency int = 5

// NewValidate returns a cobra.Command for validating Hub Manifests.
func NewValidate() *cobra.Command {
	var opts validate.Options

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate OCF manifests",
		Example: heredoc.WithCLIName(`
			# Validate interface-group.yaml file with OCF specification in default location
			<cli> manifest validate ocf-spec/0.0.1/examples/interface-group.yaml

			# Validate multiple files inside test_manifests directory with additional server-side checks
			<cli> manifest validate --server-side pkg/cli/test_manifests/*.yaml

			# Validate all Hub manifests with additional server-side checks
			<cli> manifest validate --server-side ./manifests/**/*.yaml
			
			# Validate interface-group.yaml file with custom OCF specification location 
			<cli> manifest validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml`, cli.Name),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			validation, err := validate.New(os.Stdout, opts)
			if err != nil {
				return err
			}

			return validation.Run(cmd.Context(), args)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.SchemaLocation, "schemas", "s", "", "Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.")
	flags.BoolVar(&opts.ServerSide, "server-side", false, "Executes additional manifests checks against Capact Hub.")
	flags.IntVar(&opts.MaxConcurrency, "concurrency", defaultMaxConcurrency, "Maximum number of concurrent workers.")

	return cmd
}
