package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

var (
	schemaRootDir string
)

func NewValidate() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate OCF manifests",
		Example: `ocftool validate ocf-spec/0.0.1/examples/interface-group.yaml
ocftool validate pkg/ocftool/test_manifests/*.yaml
ocftool validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validator := ocftool.NewFilesystemManifestValidator(schemaRootDir)

			fmt.Println("Validating files...")

			shouldFail := false

			for _, filepath := range args {
				result := validator.ValidateFile(filepath)

				if result.Valid() {
					color.Green("- %s: PASSED\n", filepath)
				} else {
					color.Red("- %s: FAILED\n", filepath)
					for _, err := range result.Errors {
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

	validateCmd.PersistentFlags().StringVarP(&schemaRootDir, "schemas", "s", "ocf-spec", "Path to the ocf-spec directory")

	return validateCmd
}
