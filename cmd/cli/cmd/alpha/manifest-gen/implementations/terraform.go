package implementations

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewTerraform returns a cobra.Command to bootstrap Terraform based manifests.
func NewTerraform() *cobra.Command {
	var tfContentCfg manifestgen.TerraformConfig

	cmd := &cobra.Command{
		Use:   "terraform [MANIFEST_PATH] [TERRAFORM_MODULE_PATH]",
		Short: "Generate Terraform based manifests",
		Long:  "Generate Implementation manifests based on a Terraform module",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			tfContentCfg.ManifestRef.Path = args[0]
			tfContentCfg.ModulePath = args[1]
			tfContentCfg.ManifestMetadata = common.GetDefaultMetadata()

			manifests, err := manifestgen.GenerateTerraformManifests(&tfContentCfg)
			if err != nil {
				return errors.Wrap(err, "while generating content files")
			}

			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return errors.Wrap(err, "while reading output flag")
			}

			overrideManifests, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return errors.Wrap(err, "while reading overwrite flag")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, manifests, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&tfContentCfg.InterfacePathWithRevision, "interface", "i", "", "Path with revision of the Interface, which is implemented by this Implementation")
	cmd.Flags().StringVarP(&tfContentCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")
	cmd.Flags().StringVarP(&tfContentCfg.ModuleSourceURL, "source", "s", "https://example.com/terraform-module.tgz", "Path to the Terraform module, such as URL to Tarball or Git repository")
	cmd.Flags().VarP(&tfContentCfg.Provider, "provider", "p", `Create a provider-specific workflow. Possible values: "aws", "gcp"`)

	return cmd
}

func generateTerraformManifests(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	terraformModule, err := common.AskForDirectory("Path to Terraform module", "")
	if err != nil {
		return nil, errors.Wrap(err, "while asking for path to Terraform module")
	}

	provider, err := askForProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for provider")
	}

	source, err := askForSource()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for source to Terraform module")
	}

	tfContentCfg := manifestgen.TerraformConfig{
		ImplementationConfig: manifestgen.ImplementationConfig{
			Config: manifestgen.Config{
				ManifestMetadata: opts.Metadata,
				ManifestRef: types.ManifestRef{
					Path:     common.CreateManifestPath(types.ImplementationManifestKind, opts.ManifestPath),
					Revision: opts.Revision,
				},
			},
			InterfacePathWithRevision: opts.InterfacePath,
		},
		ModulePath:      terraformModule,
		Provider:        provider,
		ModuleSourceURL: source,
	}

	files, err := manifestgen.GenerateTerraformManifests(&tfContentCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating Terraform manifests")
	}
	return files, nil
}
