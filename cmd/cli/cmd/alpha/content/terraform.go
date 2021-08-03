package content

import (
	"log"

	"capact.io/capact/internal/cli/alpha/content"
	"github.com/spf13/cobra"
)

var tfContentCfg content.TerraformConfig

// NewTerraform returns a cobra.Command to bootstrap Terraform based manifests.
func NewTerraform() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform [PREFIX] [NAME] [TERRAFORM_MODULE_PATH]",
		Short: "Bootstrap Terraform based manifests",
		Long:  "Bootstrap Terraform based manifests based on a Terraform module",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			tfContentCfg.ManifestsPrefix = args[0]
			tfContentCfg.ManifestName = args[1]
			tfContentCfg.ModulePath = args[2]

			files, err := content.GenerateTerraformManifests(&tfContentCfg)
			if err != nil {
				log.Fatalf("while generating content files: %v", err)
			}

			if err := writeManifestFiles(files); err != nil {
				log.Fatalf("while writing manifest files: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&tfContentCfg.InterfacePath, "interface", "i", "", "Interface path, which is implemented by this Implementation")

	return cmd
}
