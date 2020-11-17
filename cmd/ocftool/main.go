package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/pkg/ocftool"
)

const (
	appName = "ocftool"
)

var (
	schemaRootDir string

	rootCmd = &cobra.Command{
		Use: appName,

		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic(err)
			}
		},
	}

	validateCmd = &cobra.Command{
		Use:  "validate",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validator, err := ocftool.NewFilesystemManifestValidator(schemaRootDir)

			fmt.Println("RESULTS:")

			if err != nil {
				panic(err)
			}

			for _, filepath := range args {
				fp, err := os.Open(filepath)
				if err != nil {
					panic(err)
				}
				defer fp.Close()

				result, err := validator.ValidateYaml(fp)
				if err != nil {
					panic(err)
				}

				fmt.Printf("- %s: %v\n", filepath, result.Valid())
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
