package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

const (
	appName = "ocftool"
	version = "0.0.1"
)

func validateManifest(validator ocftool.ManifestValidator, filepath string) (failed bool) {
	fp, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open manifest: %v\n", err)
		return failed
	}
	defer fp.Close()

	result, err := validator.ValidateYaml(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed validate manifest: %v\n", err)
		return true
	}

	if result.Valid() {
		fmt.Printf("- %s: PASSED\n", filepath)
	} else {
		fmt.Printf("- %s: FAILED\n", filepath)
		for _, err := range result.Errors() {
			fmt.Printf("    %s\n", err)
		}
	}
	return !result.Valid()
}

var (
	schemaRootDir string

	rootCmd = &cobra.Command{
		Use:     appName,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic(err)
			}
		},
	}

	validateCmd = &cobra.Command{
		Use: "validate",
		Example: `ocftool validate ocf-spec/0.0.1/examples/interface-group.yaml
ocftool validate pkg/ocftool/test_manifests/*.yaml`,
		Short: "Validate OCF manifests",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validator := ocftool.NewFilesystemManifestValidator(schemaRootDir)

			fmt.Println("Validation result:")

			shouldFail := false
			for _, filepath := range args {
				if validateManifest(validator, filepath) {
					shouldFail = true
				}
			}

			if shouldFail {
				os.Exit(1)
			}
		},
	}
)

func main() {
	cobra.OnInitialize()

	validateCmd.PersistentFlags().StringVarP(&schemaRootDir, "schemas", "s", "ocf-spec", "Path to the ocf-spec directory")

	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
