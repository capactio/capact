package content

import (
	"log"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/content"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/spf13/cobra"
)

var tfContentCfg content.TerraformConfig

// NewTerraform returns a cobra.Command to bootstrap Terraform based manifests.
func NewTerraform() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform [PREFIX] [NAME] [TERRAFORM_MODULE_PATH]",
		Short: "Bootstrap Terraform based manifests",
		Long:  "Bootstrap Terraform based manifests based on a Terraform module",
		Example: heredoc.WithCLIName(`
		# Bootstrap manifests 
			<cli> alpha content terraform aws.rds deploy ./terraform-modules/aws-rds`, cli.Name),
		Args: cobra.ExactArgs(3),
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

	cmd.Flags().StringVarP(&tfContentCfg.InterfacePathWithRevision, "interface", "i", "", "Path with revision of the Interface, which is implemented by this Implementation")
	cmd.Flags().StringVarP(&tfContentCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")
	cmd.Flags().StringVarP(&tfContentCfg.ModuleSourceURL, "source", "s", "https://example.com/terraform-module.tgz", "URL to the tarball with the Terraform module")

	return cmd
}
