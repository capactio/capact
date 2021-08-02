package alpha

import (
	"log"
	"os"
	"path"

	"capact.io/capact/internal/cli/alpha/content"
	"github.com/spf13/cobra"
)

var tfContentCfg content.TerraformConfig

// NewTerraform returns a cobra.Command to bootstrap Terraform based manifests.
func NewTerraform() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform",
		Short: "Bootstrap Terraform based manifests",
		Long:  "Bootstrap Terraform based manifests based on a Terraform module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tfContentCfg.ModulePath = args[0]

			files, err := content.GenerateTerraformManifests(&tfContentCfg)
			if err != nil {
				log.Fatalf("while generating content files: %v", err)
			}

			for filename, content := range files {
				if err := os.MkdirAll(path.Dir(filename), 0750); err != nil {
					log.Fatalf("while creating directory for generated manifests: %v", err)
				}

				if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
					log.Fatalf("while writing generated manifest %s: %v", filename, err)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&tfContentCfg.ManifestName, "name", "n", "", "Name of the manifests")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&tfContentCfg.ManifestsPrefix, "prefix", "p", "", "Prefix for the manifests")
	if err := cmd.MarkFlagRequired("prefix"); err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&tfContentCfg.InterfacePath, "interface", "i", "", "Interface path, which is implemented by this Implementation")
	if err := cmd.MarkFlagRequired("interface"); err != nil {
		panic(err)
	}

	return cmd
}
