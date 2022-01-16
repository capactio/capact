package implementations

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
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
			emptyCfg.ManifestPath = args[0]
			emptyCfg.ManifestMetadata = common.GetDefaultMetadata()
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
	cmd.Flags().StringVarP(&emptyCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")

	return cmd
}

func generateEmptyManifests(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	emptyManifestCfg := manifestgen.EmptyImplementationConfig{
		ImplementationConfig: manifestgen.ImplementationConfig{
			Config: manifestgen.Config{
				ManifestMetadata: opts.Metadata,
				ManifestPath:     common.CreateManifestPath(types.ImplementationManifestKind, opts.ManifestPath),
				ManifestRevision: opts.Revision,
			},
			InterfacePathWithRevision: opts.InterfacePath,
		},
		AdditionalInputTypeName: "additional-parameters",
	}
	if slices.Contains(opts.ManifestsType, string(types.InterfaceManifestKind)) {
		emptyManifestCfg.GenerateInputType = true
	}
	files, err := manifestgen.GenerateEmptyManifests(&emptyManifestCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating empty implementation manifests")
	}
	return files, nil
}
