package manifestgen

import (
	"errors"
	"log"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/spf13/cobra"
)

var tfContentCfg manifestgen.TerraformConfig

// NewTerraform returns a cobra.Command to bootstrap Terraform based manifests.
func NewTerraform() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform [MANIFEST_PATH] [TERRAFORM_MODULE_PATH]",
		Short: "Generate Terraform based manifests",
		Long:  "Generate Terraform based manifests based on a Terraform module",
		Example: heredoc.WithCLIName(`
		# Generate Implementation manifests 
		<cli> alpha manifest-gen implementation terraform cap.implementation.aws.rds.deploy ./terraform-modules/aws-rds

		# Generate Implementation manifests for an AWS Terraform module
		<cli> alpha manifest-gen implementation terraform cap.implementation.aws.rds.deploy ./terraform-modules/aws-rds -p aws
	
		# Generate Implementation manifests for an GCP Terraform module
		<cli> alpha manifest-gen implementation terraform cap.implementation.gcp.cloudsql.deploy ./terraform-modules/cloud-sql -p gcp`, cli.Name),

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("accepts two arguments: [MANIFEST_PATH] [MODULE_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.implementation.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.implementation.[PREFIX].[NAME]"`)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			tfContentCfg.ManifestPath = args[0]
			tfContentCfg.ModulePath = args[1]

			files, err := manifestgen.GenerateTerraformManifests(&tfContentCfg)
			if err != nil {
				log.Fatalf("while generating content files: %v", err)
			}

			if err := manifestgen.WriteManifestFiles(manifestOutputDirectory, files, overrideExistingManifest); err != nil {
				log.Fatalf("while writing manifest files: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&tfContentCfg.InterfacePathWithRevision, "interface", "i", "", "Path with revision of the Interface, which is implemented by this Implementation")
	cmd.Flags().StringVarP(&tfContentCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")
	cmd.Flags().StringVarP(&tfContentCfg.ModuleSourceURL, "source", "s", "https://example.com/terraform-module.tgz", "URL to the tarball with the Terraform module")
	cmd.Flags().VarP(&tfContentCfg.Provider, "provider", "p", `Create a provider-specific workflow. Possible values: "aws", "gcp"`)

	return cmd
}
