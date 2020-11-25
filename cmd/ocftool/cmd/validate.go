package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/pkg/sdk/manifest"
)

var (
	schemaRootDir string
)

func NewValidate() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate OCF manifests",
		Example: `# Validate interface-group.yaml file with OCF specification in default location
ocftool validate ocf-spec/0.0.1/examples/interface-group.yaml

# Validate multiple files inside test_manifests directory
ocftool validate pkg/ocftool/test_manifests/*.yaml

# Validate interface-group.yaml file with custom OCF specification location 
ocftool validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml

# Validate all OCH manifests
ocftool validate ./och-content/**/*.yaml`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validator := manifest.NewFilesystemValidator(schemaRootDir)

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

	validateCmd.PersistentFlags().StringVarP(&schemaRootDir, "schemas", "s", "./ocf-spec", "Path to the ocf-spec directory")

	return validateCmd
}
