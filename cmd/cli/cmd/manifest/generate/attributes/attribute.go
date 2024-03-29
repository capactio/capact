package attributes

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/manifest/generate/common"
	"capact.io/capact/internal/cli/manifestgen"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewAttribute returns a cobra.Command to bootstrap new Attribute manifests.
func NewAttribute() *cobra.Command {
	var attributeCfg manifestgen.AttributeConfig

	cmd := &cobra.Command{
		Use:     "attribute [PATH]",
		Aliases: []string{"attributes"},
		Short:   "Generate new Attribute manifests",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.attribute.cloud.provider.aws Attribute
			<cli> manifest generate attribute cap.attribute.cloud.provider.aws`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts one argument: [MANIFEST_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.attribute.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.attribute.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			attributeCfg.ManifestRef.Path = args[0]
			attributeCfg.Metadata = common.GetDefaultInterfaceMetadata()

			manifests, err := manifestgen.GenerateAttributeTemplatingConfig(&attributeCfg)
			if err != nil {
				return errors.Wrap(err, "while generating attribute templating config")
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

	cmd.Flags().StringVarP(&attributeCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the Attribute manifest")

	return cmd
}

// GenerateAttributeFile generates new Attribute file.
func GenerateAttributeFile(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	attributeCfg := manifestgen.AttributeConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     common.CreateManifestPath(types.AttributeManifestKind, opts.ManifestPath),
				Revision: opts.Revision,
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: opts.Metadata.DocumentationURL,
			SupportURL:       opts.Metadata.SupportURL,
			IconURL:          opts.Metadata.IconURL,
			Maintainers:      opts.Metadata.Maintainers,
		},
	}
	files, err := manifestgen.GenerateAttributeTemplatingConfig(&attributeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating attribute content file")
	}
	return files, nil
}
