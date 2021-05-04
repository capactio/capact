package cmd

import (
	"fmt"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/schema"
	"capact.io/capact/pkg/sdk/manifest"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewValidate() *cobra.Command {
	schemaProvider := schema.Provider{}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate OCF manifests",
		Example: heredoc.WithCLIName(`
			# Validate interface-group.yaml file with OCF specification in default location
			<cli> validate ocf-spec/0.0.1/examples/interface-group.yaml
			
			# Validate multiple files inside test_manifests directory
			<cli> validate pkg/cli/test_manifests/*.yaml
			
			# Validate interface-group.yaml file with custom OCF specification location 
			<cli> validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml
			
			# Validate all OCH manifests
			<cli> validate ./och-content/**/*.yaml`, cli.Name),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validator := manifest.NewFilesystemValidator(schemaProvider.FileSystem())

			fmt.Println("Validating files...")

			shouldFail := false

			for _, filepath := range args {
				result, err := validator.ValidateFile(filepath)

				if err == nil && result.Valid() {
					color.Green("- %s: PASSED\n", filepath)
				} else {
					color.Red("- %s: FAILED\n", filepath)
					for _, err := range append(result.Errors, err) {
						color.Red("    %v", err)
					}

					shouldFail = true
				}
			}

			if shouldFail {
				fmt.Fprintf(os.Stderr, "Some files failed validation\n")
				os.Exit(1)
			}
		},
	}

	schemaProvider.RegisterSchemaFlags(validateCmd.PersistentFlags())

	return validateCmd
}
