package implementations

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewEmpty returns a cobra.Command to bootstrap empty Implementation manifests.
func NewEmpty() *cobra.Command {
	var emptyCfg manifestgen.EmptyImplementationConfig

	cmd := &cobra.Command{
		Use:   "empty [MANIFEST_PATH]",
		Short: "Generate empty Implementation manifests",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts one argument: [MANIFEST_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.implementation.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.implementation.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			emptyCfg.ManifestRef.Path = args[0]
			emptyCfg.Metadata = common.GetDefaultImplementationMetadata()
			emptyCfg.AdditionalInputTypeName = "additional-parameters"

			manifests, err := manifestgen.GenerateEmptyManifests(&emptyCfg)
			if err != nil {
				return errors.Wrap(err, "while generating empty implementation manifests")
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

	cmd.Flags().StringVarP(&emptyCfg.InterfacePathWithRevision, "interface", "i", "", "Path with revision of the Interface, which is implemented by this Implementation")
	cmd.Flags().StringVarP(&emptyCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")

	return cmd
}

func generateEmptyManifests(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	emptyManifestCfg := manifestgen.EmptyImplementationConfig{
		ImplementationConfig: manifestgen.ImplementationConfig{
			Config: manifestgen.Config{
				ManifestRef: types.ManifestRef{
					Path:     common.CreateManifestPath(types.ImplementationManifestKind, opts.ManifestPath),
					Revision: opts.Revision,
				},
			},
			Metadata: types.ImplementationMetadata{
				DocumentationURL: opts.Metadata.DocumentationURL,
				SupportURL:       opts.Metadata.SupportURL,
				IconURL:          opts.Metadata.IconURL,
				Maintainers:      opts.Metadata.Maintainers,
				License:          opts.Metadata.License,
			},
			InterfacePathWithRevision: opts.InterfacePath,
		},
		AdditionalInputTypeName: "additional-parameters",
	}
	files, err := manifestgen.GenerateEmptyManifests(&emptyManifestCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating empty implementation manifests")
	}
	return files, nil
}
